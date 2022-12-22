package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	"github.com/ashwin901/social-book-operator/pkg/client/clientset/versioned"
	informers "github.com/ashwin901/social-book-operator/pkg/client/informers/externalversions/ashwin901.operators/v1alpha1"
	lister "github.com/ashwin901/social-book-operator/pkg/client/listers/ashwin901.operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	clientset       kubernetes.Interface
	customClientset versioned.Interface
	lister          lister.SocialBookLister
	hasSynced       cache.InformerSynced
	queue           workqueue.RateLimitingInterface
}

func NewController(clientset kubernetes.Interface, customClientset versioned.Interface, socialBookInformer informers.SocialBookInformer) *Controller {
	controller := &Controller{
		clientset:       clientset,
		customClientset: customClientset,
		lister:          socialBookInformer.Lister(),
		hasSynced:       socialBookInformer.Informer().HasSynced,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "socialbookController"),
	}

	socialBookInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: controller.handleAdd,
			// DeleteFunc: controller.handleDelete,
		},
	)

	return controller
}

func (c *Controller) Run(ch chan struct{}) {

	fmt.Println("Starting Controller")

	defer c.queue.ShutDown()

	if !cache.WaitForCacheSync(ch, c.hasSynced) {
		fmt.Println("Cache not synced")
		return
	}

	fmt.Println("Cache synced")

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c *Controller) worker() {
	for c.processItem() {

	}
}

func (c *Controller) processItem() bool {
	item, shutdown := c.queue.Get()

	//queue is no longer used
	if shutdown {
		return false
	}

	defer c.queue.Done(item)

	key, ok := item.(string)

	if !ok {
		// key here is invalid, so we forget the item
		c.queue.Forget(item)
		return false
	}

	if c.reconcile(key) != nil {
		// TODO: handle requeing logic
		fmt.Println("Error during reconcile")
		return false
	}

	// if there were no errors we forget the item from queue
	c.queue.Forget(item)
	return true
}

func (c *Controller) reconcile(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)

	// invalid key
	if err != nil {
		return err
	}

	// get the SocialBook CR using lister
	sb, err := c.lister.SocialBooks(ns).Get(name)

	if err != nil {
		fmt.Println("Error while getting resource from the lister: ", err.Error())
		return err
	}

	err = c.handleMongoDbDeployment(sb)
	if err == nil {
		err = c.handleSocialBookDeployment(sb)
	}
	return err
}

// TODO: even if one of the resource fails the resources created before this will still remain, handle this
// creating a configmap, pv, pvc, deployment and service for MongoDB
func (c *Controller) handleMongoDbDeployment(sb *v1alpha1.SocialBook) error {
	// config map
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + "-mongo-cm",
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Data: map[string]string{
			"mongo-root-username": sb.Spec.UserName,
			"mongo-root-password": sb.Spec.Password,
		},
	}

	cm, err := c.clientset.CoreV1().ConfigMaps(sb.Namespace).Create(context.Background(), cm, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("Error while creating config map: ", err.Error())
		return err
	}

	fmt.Println("Config map created")

	// create a persistent volume
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + "-mongo-pv",
			OwnerReferences: getOwnerReference(sb),
		},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("1Gi"),
			},
			ClaimRef: &corev1.ObjectReference{
				Namespace: sb.Namespace,
				Name:      sb.Name + "-mongo-pvc",
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/tmp/data",
				},
			},
		},
	}

	_, err = c.clientset.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("Error while creating persistent volume: ", err.Error())
		return err
	}

	fmt.Println("PV created")

	// create a persistent volume claim
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + "-mongo-pvc",
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
		},
	}
	_, err = c.clientset.CoreV1().PersistentVolumeClaims(sb.Namespace).Create(context.Background(), pvc, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("Error while creating persistent volume claim: ", err.Error())
		return err
	}

	fmt.Println("PVC created")

	var replicas int32
	replicas = 1

	// mongo db deployment
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + "-mongodb",
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "mongodb-" + sb.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mongodb-" + sb.Name,
					Labels: map[string]string{
						"app": "mongodb-" + sb.Name,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "mongo-volume-" + sb.Name,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: sb.Name + "-mongo-pvc",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "mongodb",
							Image: "mongo",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 27017,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "MONGO_INITDB_ROOT_USERNAME",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "mongo-root-username",
										},
									},
								},
								{
									Name: "MONGO_INITDB_ROOT_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "mongo-root-password",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "mongo-volume-" + sb.Name,
									MountPath: "/data/db",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = c.clientset.AppsV1().Deployments(sb.Namespace).Create(context.Background(), dep, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("Error while creating mongo deployment: ", err.Error())
		return err
	}

	fmt.Println("Mongo Deployment created")

	// mongo db service
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + "-mongo-svc",
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "mongodb-" + sb.Name,
			},
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.FromInt(27017),
					Port:       27017,
				},
			},
		},
	}

	_, err = c.clientset.CoreV1().Services(sb.Namespace).Create(context.Background(), svc, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error while creating mongo service: ", err.Error())
		return err
	}

	fmt.Println("Mongo Service created")
	return nil
}

// creating a configmap, deployment and service for socialbook(image: ashwin901/social-book-server)
func (c *Controller) handleSocialBookDeployment(sb *v1alpha1.SocialBook) error {

	mongodbUri := "mongodb://" + sb.Spec.UserName + ":" + sb.Spec.Password + "@" + sb.Name + "-mongo-svc" + ":27017"

	portNumber, err := strconv.Atoi(sb.Spec.Port)

	if err != nil {
		fmt.Println("invalid port number: ", err.Error())
		return err
	}

	// config map
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + "-cm",
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Data: map[string]string{
			"port":           sb.Spec.Port,
			"secret":         "secret", // any random string (used for jwt token)
			"stripe-api-key": "abc",    // api key used for payments
			"user-email":     "abc@email.com",
			"user-pwd":       "admin",
			"client-url":     "sb-client.com", // redirect url after email verification
			"mongodb-uri":    mongodbUri,
		},
	}

	cm, err = c.clientset.CoreV1().ConfigMaps(sb.Namespace).Create(context.Background(), cm, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("Error while creating socialbook configmap: ", err.Error())
		return err
	}

	fmt.Println("Config map created")

	// social book deployment, image: ashwin901/social-book-server
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name,
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &sb.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "socialbook-" + sb.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "socialbook-" + sb.Name,
					Labels: map[string]string{
						"app": "socialbook-" + sb.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "socialbook",
							Image: "ashwin901/social-book-server",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(portNumber),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "PORT",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "port",
										},
									},
								},
								{
									Name: "MONGODB_URI",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "mongodb-uri",
										},
									},
								},
								{
									Name: "SECRET",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "secret",
										},
									},
								},
								{
									Name: "STRIPE_API_KEY",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "stripe-api-key",
										},
									},
								},
								{
									Name: "USER_EMAIL",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "user-email",
										},
									},
								},
								{
									Name: "USER_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "user-pwd",
										},
									},
								},
								{
									Name: "CLIENT_URL",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cm.Name,
											},
											Key: "client-url",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	dep, err = c.clientset.AppsV1().Deployments(sb.Namespace).Create(context.Background(), dep, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("Error while creating socialbook deployment: ", err.Error())
		return err
	}

	fmt.Println("SocialBook Deployment created")

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + "-svc",
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Spec: corev1.ServiceSpec{
			Selector: dep.Spec.Template.Labels,
			Type:     corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.FromInt(portNumber),
					Port:       int32(portNumber),
					NodePort:   32000,
				},
			},
		},
	}

	_, err = c.clientset.CoreV1().Services(sb.Namespace).Create(context.Background(), svc, metav1.CreateOptions{})
	if err != nil {
		fmt.Println("Error while creating socialbook service: ", err.Error())
		return err
	}

	fmt.Println("SocailBook Service created")

	return nil
}

func getOwnerReference(sb *v1alpha1.SocialBook) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(sb, v1alpha1.SchemeGroupVersion.WithKind("SocialBook")),
	}
}

func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("Add event")
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		fmt.Println("Error while getting key for object: ", err.Error())
		return
	}

	c.queue.Add(key)
}

func (c *Controller) handleDelete(obj interface{}) {
	fmt.Println("Delete event")
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		fmt.Println("Error while getting key for object: ", err.Error())
		return
	}

	c.queue.Add(key)
}
