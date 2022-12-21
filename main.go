package main

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ashwin901/social-book-operator/pkg/client/clientset/versioned"
	"github.com/ashwin901/social-book-operator/pkg/client/informers/externalversions"
	"github.com/ashwin901/social-book-operator/pkg/controller"
)

func main() {

	configFile := "/home/ashwin901/.kube/config"

	config, err := clientcmd.BuildConfigFromFlags("", configFile)

	if err != nil {
		fmt.Println("Error while building config: ", err.Error())
		return
	}

	// kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		fmt.Println("Error while creating clientset: ", err.Error())
		return
	}

	// clientset from the generated code for the new group
	customClientset, err := versioned.NewForConfig(config)

	if err != nil {
		fmt.Println("Error while creating clientset: ", err.Error())
		return
	}

	ch := make(chan struct{})
	factory := externalversions.NewSharedInformerFactory(customClientset, 10*time.Minute)

	// initializing controller
	controller := controller.NewController(clientset, customClientset, factory.Operators().V1alpha1().SocialBooks())

	// initialising all the requested informers
	factory.Start(ch)

	// start the controller
	controller.Run(ch)
}
