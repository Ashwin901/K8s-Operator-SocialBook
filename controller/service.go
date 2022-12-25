package controller

import (
	"strconv"

	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newService(sb *v1alpha1.SocialBook, appType string) *corev1.Service {
	if appType == MongoDB {
		return newMongoService(sb)
	}

	return newSocialBookService(sb)
}

func newMongoService(sb *v1alpha1.SocialBook) *corev1.Service {

	svcName := sb.Name + MongoDB

	// mongo db service
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            svcName,
			Namespace:       sb.Namespace,
			OwnerReferences: setOwnerReference(sb),
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

func newSocialBookService(sb *v1alpha1.SocialBook) *corev1.Service {
	portNumber, _ := strconv.Atoi(sb.Spec.Port)
	svcName := sb.Name

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            svcName,
			Namespace:       sb.Namespace,
			OwnerReferences: setOwnerReference(sb),
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
