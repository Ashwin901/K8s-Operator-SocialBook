package controller

import (
	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newMongoConfigMap(sb *v1alpha1.SocialBook) *corev1.ConfigMap {
	cmName := sb.Name + "-mongo-cm"
	// config map
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cmName,
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Data: map[string]string{
			"mongo-root-username": sb.Spec.UserName,
			"mongo-root-password": sb.Spec.Password,
		},
	}

	return cm
}

func newSocialBookConfigMap(sb *v1alpha1.SocialBook) *corev1.ConfigMap {
	cmName := sb.Name + "-cm"
	mongodbUri := "mongodb://" + sb.Spec.UserName + ":" + sb.Spec.Password + "@" + sb.Name + "-mongo-svc" + ":27017"

	// config map
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cmName,
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

	return cm
}
