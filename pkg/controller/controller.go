package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	"github.com/ashwin901/social-book-operator/pkg/client/clientset/versioned"
	informers "github.com/ashwin901/social-book-operator/pkg/client/informers/externalversions/ashwin901.operators/v1alpha1"
	lister "github.com/ashwin901/social-book-operator/pkg/client/listers/ashwin901.operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			AddFunc:    controller.handleAdd,
			DeleteFunc: controller.handleDelete,
		},
	)

	return controller
}

func (c *Controller) Run(ch chan struct{}) {

	fmt.Println("Starting Controller")

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

	return c.handleMongoDbDeployment(*sb)
}

func (c *Controller) handleMongoDbDeployment(sb v1alpha1.SocialBook) error {
	// config map
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sb.Name,
			Namespace: sb.Namespace,
		},
		Data: map[string]string{
			"mongo-root-username": sb.Spec.UserName,
			"mongo-root-password": sb.Spec.Password,
		},
	}

	_, err := c.clientset.CoreV1().ConfigMaps(sb.Namespace).Create(context.Background(), cm, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("Error while creating config map: ", err.Error())
		return err
	}

	// create a persistent volume
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mongo-pv-" + sb.Name,
		},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("1Gi"),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/mnt/data",
				},
			},
		},
	}

	_, err = c.clientset.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})

	if err != nil {
		fmt.Println("Error while creating persistent volume: ", err.Error())
		return err
	}

	// create a persistent volume claim
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongo-pvc-" + sb.Name,
			Namespace: sb.Namespace,
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

	// mongo db deployment

	// mong db service
	return nil
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
