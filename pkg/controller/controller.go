package controller

import (
	"fmt"

	"github.com/ashwin901/social-book-operator/pkg/client/clientset/versioned"
	informers "github.com/ashwin901/social-book-operator/pkg/client/informers/externalversions/ashwin901.operators/v1alpha1"
	lister "github.com/ashwin901/social-book-operator/pkg/client/listers/ashwin901.operators/v1alpha1"
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

	<-ch
}

func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("New Social Book added")
}

func (c *Controller) handleDelete(obj interface{}) {
	fmt.Println("Social Book deleted")
}
