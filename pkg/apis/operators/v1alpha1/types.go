package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SocialBook struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec SocialBookSpec
}

type SocialBookSpec struct {
	Replicas int32
	UserName string
	Password string
	Port     string
}
