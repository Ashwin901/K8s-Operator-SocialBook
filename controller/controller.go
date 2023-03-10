package controller

import (
	"context"
	"fmt"
	"log"
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
	networkingLister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	ConfigMap             = "-cm"
	Service               = "-svc"
	PersistentVolume      = "-pv"
	PersistentVolumeClaim = "-pvc"
	NetworkPolicy         = "-np"
	Deployment            = "-dep"
	MongoDB               = "-mongo"
	SocialBook            = "-sb"
	Kind                  = "SocialBook"
	Success               = "Success"
	Pending               = "Pending"
	Failure               = "Failed"
	Image                 = "ashwin901/social-book-server"
)

type Controller struct {
	clientset           kubernetes.Interface
	customClientset     versioned.Interface
	socialbookLister    lister.SocialBookLister
	deploymentLister    appsLister.DeploymentLister
	serviceLister       coreLister.ServiceLister
	configMapLister     coreLister.ConfigMapLister
	pvLister            coreLister.PersistentVolumeLister
	pvcLister           coreLister.PersistentVolumeClaimLister
	networkPolicyLister networkingLister.NetworkPolicyLister
	socialbookSynced    cache.InformerSynced
	deploymentSynced    cache.InformerSynced
	serviceSynced       cache.InformerSynced
	configMapSynced     cache.InformerSynced
	pvSynced            cache.InformerSynced
	pvcSynced           cache.InformerSynced
	networkPolicySynced cache.InformerSynced
	queue               workqueue.RateLimitingInterface
}

func NewController(clientset kubernetes.Interface, customClientset versioned.Interface, socialBookInformer informers.SocialBookInformer, factory kubeInformers.SharedInformerFactory) *Controller {

	controller := &Controller{
		clientset:           clientset,
		customClientset:     customClientset,
		socialbookLister:    socialBookInformer.Lister(),
		deploymentLister:    factory.Apps().V1().Deployments().Lister(),
		serviceLister:       factory.Core().V1().Services().Lister(),
		configMapLister:     factory.Core().V1().ConfigMaps().Lister(),
		pvLister:            factory.Core().V1().PersistentVolumes().Lister(),
		pvcLister:           factory.Core().V1().PersistentVolumeClaims().Lister(),
		networkPolicyLister: factory.Networking().V1().NetworkPolicies().Lister(),
		socialbookSynced:    socialBookInformer.Informer().HasSynced,
		deploymentSynced:    factory.Apps().V1().Deployments().Informer().HasSynced,
		serviceSynced:       factory.Core().V1().Services().Informer().HasSynced,
		configMapSynced:     factory.Core().V1().ConfigMaps().Informer().HasSynced,
		pvSynced:            factory.Core().V1().PersistentVolumes().Informer().HasSynced,
		pvcSynced:           factory.Core().V1().PersistentVolumeClaims().Informer().HasSynced,
		networkPolicySynced: factory.Networking().V1().NetworkPolicies().Informer().HasSynced,
		queue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "socialbookController"),
	}

	// when socialbook custom resource is deleted all items created by the controller because of it are also deleted (because of owner reference)
	// so no need to handle delete event
	socialBookInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: controller.addItemToQueue,
			UpdateFunc: func(oldObj, newObj interface{}) {
				old := oldObj.(metav1.Object)
				new := newObj.(metav1.Object)
				// if it has the same resource version we ignore it
				if old.GetResourceVersion() == new.GetResourceVersion() {
					return
				}
				controller.addItemToQueue(newObj)
			},
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

	factory.Networking().V1().NetworkPolicies().Informer().AddEventHandler(
		controller.getEventHandlerFunctions(),
	)

	return controller
}

// monitoring only delete and update event of other resources
func (c *Controller) getEventHandlerFunctions() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
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

	log.Printf("Starting Controller")

	defer c.queue.ShutDown()

	if !cache.WaitForCacheSync(ch, c.socialbookSynced, c.configMapSynced, c.pvSynced, c.pvcSynced, c.serviceSynced, c.deploymentSynced, c.networkPolicySynced) {
		log.Printf("Cache not synced")
		return
	}

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
		// requeue the item if there were any errors
		c.queue.AddRateLimited(item)
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
		log.Printf("Error %s while getting %s from the lister", err.Error(), sb.Name)
		return err
	}

	// making a copy to update status
	sbCopy := sb.DeepCopy()
	sbCopy.Status.MongoDB = Pending
	sbCopy.Status.SocialBook = Pending

	defer c.updateSocialbookStatus(sbCopy)

	// creating all the resources required for mongodb
	if err = c.handleMongoDbDeployment(sb, sbCopy); err != nil {
		log.Printf("Error %s while creating MongoDB deployment for %s", err.Error(), sb.Name)
		sbCopy.Status.MongoDB = Failure
		return err
	}

	// creating resources for socialbook
	if err = c.handleSocialBookDeployment(sb, sbCopy); err != nil {
		log.Printf("Error %s while creating SocialBook deployment for %s", err.Error(), sb.Name)
		sbCopy.Status.SocialBook = Failure
		return err
	}

	log.Printf("MongoDB and SocalBook successfully deployed for %s", sb.Name)

	return nil
}

// creating a pv, pvc, deployment and service for MongoDB
func (c *Controller) handleMongoDbDeployment(sb *v1alpha1.SocialBook, sbCopy *v1alpha1.SocialBook) error {
	cmName := sb.Name + ConfigMap
	pvName := sb.Name + PersistentVolume
	pvcName := sb.Name + PersistentVolumeClaim
	depName := sb.Name + MongoDB
	svcName := sb.Name + MongoDB
	npName := sb.Name + MongoDB + NetworkPolicy

	// creating a configmap
	cm, err := c.configMapLister.ConfigMaps(sb.Namespace).Get(cmName)
	err = c.handleResourceCreation(err, cm, sb, "", ConfigMap)
	if err != nil {
		return err
	}

	// Creating a PV for mongoDB
	pv, err := c.pvLister.Get(pvName)
	err = c.handleResourceCreation(err, pv, sb, "", PersistentVolume)
	if err != nil {
		return err
	}

	// Creating a PVC for mongoDB
	pvc, err := c.pvcLister.PersistentVolumeClaims(sb.Namespace).Get(pvcName)
	err = c.handleResourceCreation(err, pvc, sb, "", PersistentVolumeClaim)
	if err != nil {
		return err
	}

	// Creating mongoDB deployment
	dep, err := c.deploymentLister.Deployments(sb.Namespace).Get(depName)
	err = c.handleResourceCreation(err, dep, sb, MongoDB, Deployment)
	if err != nil {
		return err
	}

	// Creating the corresponding service
	svc, err := c.serviceLister.Services(sb.Namespace).Get(svcName)
	err = c.handleResourceCreation(err, svc, sb, MongoDB, Service)
	if err != nil {
		return err
	}

	// Creating network policy for mongodb pods - Ingress rule
	np, err := c.networkPolicyLister.NetworkPolicies(sb.Namespace).Get(npName)
	err = c.handleResourceCreation(err, np, sb, MongoDB, NetworkPolicy)
	if err != nil {
		return err
	}

	sbCopy.Status.MongoDB = Success
	return nil
}

// creating deployment and service for socialbook(image: ashwin901/social-book-server)
func (c *Controller) handleSocialBookDeployment(sb *v1alpha1.SocialBook, sbCopy *v1alpha1.SocialBook) error {

	svcName := sb.Name
	npName := sb.Name + NetworkPolicy

	// Creating a deployment for image: ashwin901/social-book-server
	dep, err := c.deploymentLister.Deployments(sb.Namespace).Get(sb.Name)
	err = c.handleResourceCreation(err, dep, sb, SocialBook, Deployment)
	if err != nil {
		return err
	}

	// checking if the replicas of the deployment are same as the desired number of replicas
	if sb.Spec.Replicas != *(dep.Spec.Replicas) {
		fmt.Println("Incorrect number of replicas")
		dep, err = c.clientset.AppsV1().Deployments(sb.Namespace).Update(context.Background(), newDeployment(sb, SocialBook), metav1.UpdateOptions{})

		if err != nil {
			fmt.Println("Error while updating deployment")
			return err
		}
	}

	// Creating the corresponding service(external)
	svc, err := c.serviceLister.Services(sb.Namespace).Get(svcName)
	err = c.handleResourceCreation(err, svc, sb, SocialBook, Service)
	if err != nil {
		return err
	}

	// Creating network policy for socialbook pods - Egress rule
	np, err := c.networkPolicyLister.NetworkPolicies(sb.Namespace).Get(npName)
	err = c.handleResourceCreation(err, np, sb, SocialBook, NetworkPolicy)
	if err != nil {
		return err
	}

	sbCopy.Status.SocialBook = Success
	return nil
}

func (c *Controller) handleResourceCreation(err error, resource interface{}, sb *v1alpha1.SocialBook, appType string, resourceName string) error {
	if errors.IsNotFound(err) {
		switch resourceName {
		case ConfigMap:
			resource, err = c.clientset.CoreV1().ConfigMaps(sb.Namespace).Create(context.Background(), newConfigMap(sb), metav1.CreateOptions{})
			break
		case PersistentVolume:
			resource, err = c.clientset.CoreV1().PersistentVolumes().Create(context.Background(), newPersistentVolume(sb), metav1.CreateOptions{})
			break
		case PersistentVolumeClaim:
			resource, err = c.clientset.CoreV1().PersistentVolumeClaims(sb.Namespace).Create(context.Background(), newPersistentVolumeClaim(sb), metav1.CreateOptions{})
			break
		case Service:
			resource, err = c.clientset.CoreV1().Services(sb.Namespace).Create(context.Background(), newService(sb, appType), metav1.CreateOptions{})
			break
		case Deployment:
			resource, err = c.clientset.AppsV1().Deployments(sb.Namespace).Create(context.Background(), newDeployment(sb, appType), metav1.CreateOptions{})
			break
		case NetworkPolicy:
			resource, err = c.clientset.NetworkingV1().NetworkPolicies(sb.Namespace).Create(context.Background(), newNetworkPolicy(sb, appType), metav1.CreateOptions{})
			break
		default:
			err = fmt.Errorf("Unkown resource %s", resourceName)
			break
		}
	}

	if err != nil {
		return err
	}

	// check if the resource is controlled by current SocialBook resource
	if !metav1.IsControlledBy(resource.(metav1.Object), sb) {
		return fmt.Errorf("%s", "Resource already exists")
	}

	return err
}

// updating the status of SocialBook custom resource
func (c *Controller) updateSocialbookStatus(sbCopy *v1alpha1.SocialBook) {
	_, err := c.customClientset.OperatorsV1alpha1().SocialBooks(sbCopy.Namespace).UpdateStatus(context.Background(), sbCopy, metav1.UpdateOptions{})

	if err != nil {
		log.Printf("Error %s while updating status of %s", err.Error(), sbCopy.Name)
	}

	log.Printf("Status for %s successfully updated", sbCopy.Name)
}

// adding SocialBook items to workqueue for processing
func (c *Controller) addItemToQueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)

	if err != nil {
		log.Printf("Error %s while adding item %s to queue", err.Error(), obj)
		return
	}

	c.queue.Add(key)
}

// checks if the resource is owned by "SocialBook" kind
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		log.Printf("Invalid object %s", obj)
		return
	}

	if owner := metav1.GetControllerOf(object); owner != nil {
		if owner.Kind != Kind {
			return
		}

		sb, err := c.socialbookLister.SocialBooks(object.GetNamespace()).Get(owner.Name)

		if err != nil {
			log.Printf("Error %s while getting socialbook %s", err.Error(), owner.Name)
			return
		}

		c.addItemToQueue(sb)
	}
}

// used to set the owner reference of resources
func setOwnerReference(sb *v1alpha1.SocialBook) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(sb, v1alpha1.SchemeGroupVersion.WithKind(Kind)),
	}
}
