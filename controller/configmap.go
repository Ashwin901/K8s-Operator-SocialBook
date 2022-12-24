package controller

import (
	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newConfigMap(sb *v1alpha1.SocialBook) *corev1.ConfigMap {
	cmName := sb.Name + "-cm"
	mongodbUri := "mongodb://" + sb.Spec.MongoUsername + ":" + sb.Spec.MongoPassword + "@" + sb.Name + "-mongo-svc" + ":27017"

	// config map
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            cmName,
			Namespace:       sb.Namespace,
			OwnerReferences: setOwnerReference(sb),
		},
		Data: map[string]string{
			"mongo-root-username": sb.Spec.MongoUsername,
			"mongo-root-password": sb.Spec.MongoPassword,
			"port":                sb.Spec.Port,
			"secret":              sb.Spec.JwtSecret,    // any random string (used for jwt token)
			"stripe-api-key":      sb.Spec.StripeApiKey, // api key used for payments
			"user-email":          sb.Spec.EmailId,
			"user-pwd":            sb.Spec.Password,
			"client-url":          sb.Spec.ClientUrl, // redirect url after email verification
			"mongodb-uri":         mongodbUri,
		},
	}

	return cm
}
