package main

import (
	"log"
	"time"

	kubeInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ashwin901/social-book-operator/controller"
	"github.com/ashwin901/social-book-operator/pkg/client/clientset/versioned"
	"github.com/ashwin901/social-book-operator/pkg/client/informers/externalversions"
)

func main() {

	configFile := "/home/ashwin901/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", configFile)

	if err != nil {
		log.Printf("Error %s while building config for %s", err.Error(), configFile)

		// if there is an error, try to get config from inside the cluster
		config, err = rest.InClusterConfig()

		if err != nil {
			log.Printf("Error %s while building config from inside the cluster", err.Error())
			return
		}
	}

	// kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		log.Printf("Error %s while creating clientset", err.Error())
		return
	}

	// clientset from the generated code for the new group
	customClientset, err := versioned.NewForConfig(config)

	if err != nil {
		log.Printf("Error %s while creating customm clientset: ", err.Error())
		return
	}

	ch := make(chan struct{})
	factory := kubeInformers.NewSharedInformerFactory(clientset, 10*time.Minute)
	customFactory := externalversions.NewSharedInformerFactory(customClientset, 10*time.Minute)

	// initializing controller
	controller := controller.NewController(clientset, customClientset, customFactory.Operators().V1alpha1().SocialBooks(), factory)

	// initialising all the requested informers
	customFactory.Start(ch)
	factory.Start(ch)

	// start the controller
	controller.Run(ch)
}
