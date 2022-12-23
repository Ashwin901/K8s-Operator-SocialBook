package controller

import (
	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newMongoService(sb *v1alpha1.SocialBook) *corev1.Service {

	svcName := sb.Name + "-mongo-svc"

	// mongo db service
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            svcName,
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

	return svc
}

func newSocialBookService(sb *v1alpha1.SocialBook, portNumber int) *corev1.Service {

	svcName := sb.Name + "-svc"

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            svcName,
			Namespace:       sb.Namespace,
			OwnerReferences: getOwnerReference(sb),
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "mongodb-" + sb.Name,
			},
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					TargetPort: intstr.FromInt(portNumber),
					Port:       int32(portNumber),
					NodePort:   32000,
				},
			},
		},
	}

	return svc
}
