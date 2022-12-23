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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsLister "k8s.io/client-go/listers/apps/v1"
	coreLister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	clientset        kubernetes.Interface
	customClientset  versioned.Interface
	socialbookLister lister.SocialBookLister
	deploymentLister appsLister.DeploymentLister
	serviceLister    coreLister.ServiceLister
	configMapLister  coreLister.ConfigMapLister
	pvLister         coreLister.PersistentVolumeLister
	pvcLister        coreLister.PersistentVolumeClaimLister
	socialbookSynced cache.InformerSynced
	deploymentSynced cache.InformerSynced
	serviceSynced    cache.InformerSynced
	configMapSynced  cache.InformerSynced
	pvSynced         cache.InformerSynced
	pvcSynced        cache.InformerSynced
	queue            workqueue.RateLimitingInterface
}

func NewController(clientset kubernetes.Interface, customClientset versioned.Interface, socialBookInformer informers.SocialBookInformer, factory kubeInformers.SharedInformerFactory) *Controller {

	controller := &Controller{
		clientset:        clientset,
		customClientset:  customClientset,
		socialbookLister: socialBookInformer.Lister(),
		deploymentLister: factory.Apps().V1().Deployments().Lister(),
		serviceLister:    factory.Core().V1().Services().Lister(),
		configMapLister:  factory.Core().V1().ConfigMaps().Lister(),
		pvLister:         factory.Core().V1().PersistentVolumes().Lister(),
		pvcLister:        factory.Core().V1().PersistentVolumeClaims().Lister(),
		socialbookSynced: socialBookInformer.Informer().HasSynced,
		deploymentSynced: factory.Apps().V1().Deployments().Informer().HasSynced,
		serviceSynced:    factory.Core().V1().Services().Informer().HasSynced,
		configMapSynced:  factory.Core().V1().ConfigMaps().Informer().HasSynced,
		pvSynced:         factory.Core().V1().PersistentVolumes().Informer().HasSynced,
		pvcSynced:        factory.Core().V1().PersistentVolumeClaims().Informer().HasSynced,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "socialbookController"),
	}

	// when socialbook custom resource is deleted all items created by the controller because of it are also deleted (because of owner reference)
	// so no need to handle delete event
	socialBookInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: controller.addItemToQueue,
		},
	)

	// adding event handler functions for all the other resources
	factory.Apps().V1().Deployments().Informer().AddEventHandler(
		controller.getEventHandlerFunctions(),
	)

	factory.Core().V1().Services().Informer().AddEventHandler(
		controller.getEventHandlerFunctions(),
	)

	factory.Core().V1().ConfigMaps().Informer().AddEventHandler(
		controller.getEventHandlerFunctions(),
	)

	factory.Core().V1().PersistentVolumes().Informer().AddEventHandler(
		controller.getEventHandlerFunctions(),
	)

	factory.Core().V1().PersistentVolumeClaims().Informer().AddEventHandler(
		controller.getEventHandlerFunctions(),
	)

	return controller
}

func (c *Controller) getEventHandlerFunctions() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleObject,
		DeleteFunc: c.handleObject,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(metav1.Object)
			new := newObj.(metav1.Object)
			// if it has the same resource version we ignore it
			if old.GetResourceVersion() == new.GetResourceVersion() {
				return
			}
			c.handleObject(newObj)
		},
	}
}

func (c *Controller) Run(ch chan struct{}) {

	fmt.Println("Starting Controller")

	defer c.queue.ShutDown()

	if !cache.WaitForCacheSync(ch, c.socialbookSynced) {
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
		return true
	}

	if c.reconcile(key) != nil {
		// requeue the item
		c.queue.AddRateLimited(item)
		fmt.Println("Error during reconcile, item requeued")
		return true
	}

	// if there were no errors we forget the item from queue
	c.queue.Forget(item)
	return true
}

func (c *Controller) reconcile(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		return nil // no need to requeue as the key is invalid
	}

	// get the SocialBook CR using lister
	sb, err := c.socialbookLister.SocialBooks(ns).Get(name)

	if err != nil {
		if errors.IsNotFound(err) {
			return nil // object not present, so no need to requeue
		}
		fmt.Println("Error while getting resource from the lister: ", err.Error())
		return err
	}
	err = c.handleMongoDbDeployment(sb)
	if err == nil {
		err = c.handleSocialBookDeployment(sb)
	}
	return err
}

// creating a configmap, pv, pvc, deployment and service for MongoDB
func (c *Controller) handleMongoDbDeployment(sb *v1alpha1.SocialBook) error {

	cmName := sb.Name + "-mongo-cm"
	pvName := sb.Name + "-mongo-pv"
	pvcName := sb.Name + "-mongo-pvc"
	depName := sb.Name + "-mongodb"
	svcName := sb.Name + "-mongo-svc"

	cm, err := c.configMapLister.ConfigMaps(sb.Namespace).Get(cmName)
	if errors.IsNotFound(err) {
		cm, err = c.clientset.CoreV1().ConfigMaps(sb.Namespace).Create(context.Background(), newMongoConfigMap(sb), metav1.CreateOptions{})
	}
	if err != nil {
		fmt.Println("Error while creating/getting config map: ", err.Error())
		return err
	}
	// configmap with the same name exists but is not controlled by current sb
	if !metav1.IsControlledBy(cm, sb) {
		return fmt.Errorf("%s", "Config Map already exists")
	}
	fmt.Println("Config map configured")

	pv, err := c.pvLister.Get(pvName)
	if errors.IsNotFound(err) {
		// create a persistent volume
		pv, err = c.clientset.CoreV1().PersistentVolumes().Create(context.Background(), newPersistentVolume(sb), metav1.CreateOptions{})
	}
	if err != nil {
		fmt.Println("Error while creating persistent volume: ", err.Error())
		return err
	}
	if !metav1.IsControlledBy(pv, sb) {
		return fmt.Errorf("%s", "Persistent volume already exists")
	}
	fmt.Println("PV created")

	pvc, err := c.pvcLister.PersistentVolumeClaims(sb.Namespace).Get(pvcName)
	if errors.IsNotFound(err) {
		// create a persistent volume claim
		pvc, err = c.clientset.CoreV1().PersistentVolumeClaims(sb.Namespace).Create(context.Background(), newPersistentVolumeClaim(sb), metav1.CreateOptions{})
	}
	if err != nil {
		fmt.Println("Error while creating persistent volume claim: ", err.Error())
		return err
	}
	if !metav1.IsControlledBy(pvc, sb) {
		return fmt.Errorf("%s", "Persistent volume claim already exists")
	}
	fmt.Println("PVC created")

	dep, err := c.deploymentLister.Deployments(sb.Namespace).Get(depName)

	if errors.IsNotFound(err) {
		dep, err = c.clientset.AppsV1().Deployments(sb.Namespace).Create(context.Background(), newMongoDeployment(sb), metav1.CreateOptions{})
	}

	if err != nil {
		fmt.Println("Error while creating mongo deployment: ", err.Error())
		return err
	}

	if !metav1.IsControlledBy(dep, sb) {
		return fmt.Errorf("%s", "Deployment already exists")
	}

	fmt.Println("Mongo Deployment created")

	svc, err := c.serviceLister.Services(sb.Namespace).Get(svcName)
	if errors.IsNotFound(err) {
		svc, err = c.clientset.CoreV1().Services(sb.Namespace).Create(context.Background(), newMongoService(sb), metav1.CreateOptions{})
	}
	if err != nil {
		fmt.Println("Error while creating mongo service: ", err.Error())
		return err
	}
	if !metav1.IsControlledBy(svc, sb) {
		return fmt.Errorf("%s", "Service already exists")
	}

	fmt.Println("Mongo Service created")
	return nil
}

// creating a configmap, deployment and service for socialbook(image: ashwin901/social-book-server)
func (c *Controller) handleSocialBookDeployment(sb *v1alpha1.SocialBook) error {
	portNumber, err := strconv.Atoi(sb.Spec.Port)
	cmName := sb.Name + "-cm"
	svcName := sb.Name + "-svc"

	if err != nil {
		fmt.Println("invalid port number: ", err.Error())
		return err
	}
	cm, err := c.configMapLister.ConfigMaps(sb.Namespace).Get(cmName)
	if errors.IsNotFound(err) {
		cm, err = c.clientset.CoreV1().ConfigMaps(sb.Namespace).Create(context.Background(), newSocialBookConfigMap(sb), metav1.CreateOptions{})
	}
	if err != nil {
		fmt.Println("Error while creating socialbook configmap: ", err.Error())
		return err
	}
	if !metav1.IsControlledBy(cm, sb) {
		return fmt.Errorf("%s", "Service already exists")
	}
	fmt.Println("Config map created")

	dep, err := c.deploymentLister.Deployments(sb.Namespace).Get(sb.Name)

	if errors.IsNotFound(err) {
		// social book deployment, image: ashwin901/social-book-server
		dep, err = c.clientset.AppsV1().Deployments(sb.Namespace).Create(context.Background(), newSocialBookDeployment(sb, portNumber), metav1.CreateOptions{})
	}

	if err != nil {
		fmt.Println("Error while creating socialbook deployment: ", err.Error())
		return err
	}

	if !metav1.IsControlledBy(dep, sb) {
		return fmt.Errorf("%s", "Service already exists")
	}

	fmt.Println("SocialBook Deployment created")

	svc, err := c.serviceLister.Services(sb.Namespace).Get(svcName)

	if errors.IsNotFound(err) {
		svc, err = c.clientset.CoreV1().Services(sb.Namespace).Create(context.Background(), newSocialBookService(sb, portNumber), metav1.CreateOptions{})
	}

	if err != nil {
		fmt.Println("Error while creating socialbook service: ", err.Error())
		return err
	}

	if !metav1.IsControlledBy(svc, sb) {
		return fmt.Errorf("%s", "Service already exists")
	}

	fmt.Println("SocailBook Service created")

	return nil
}

func (c *Controller) addItemToQueue(obj interface{}) {
	fmt.Println("Add event")
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		fmt.Println("Error while getting key for object: ", err.Error())
		return
	}

	c.queue.Add(key)
}

func (c *Controller) handleObject(obj interface{}) {
	fmt.Println("Change detected")
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		fmt.Println("Invalid object")
		return
	}

	if owner := metav1.GetControllerOf(object); owner != nil {
		if owner.Kind != "SocialBook" {
			return
		}

		sb, err := c.socialbookLister.SocialBooks(object.GetNamespace()).Get(owner.Name)

		if err != nil {
			fmt.Println("Error while getting socialbook")
			return
		}

		c.addItemToQueue(sb)
	}
}

func setOwnerReference(sb *v1alpha1.SocialBook) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(sb, v1alpha1.SchemeGroupVersion.WithKind("SocialBook")),
	}
}
