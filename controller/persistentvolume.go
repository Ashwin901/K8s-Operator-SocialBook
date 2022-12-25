package controller

import (
	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPersistentVolume(sb *v1alpha1.SocialBook) *corev1.PersistentVolume {
	pvName := sb.Name + PersistentVolume
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pvName,
			OwnerReferences: setOwnerReference(sb),
		},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("1Gi"),
			},
			ClaimRef: &corev1.ObjectReference{
				Namespace: sb.Namespace,
				Name:      sb.Name + PersistentVolumeClaim,
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/tmp/data",
				},
			},
		},
	}

	return pv
}

func newPersistentVolumeClaim(sb *v1alpha1.SocialBook) *corev1.PersistentVolumeClaim {
	pvcName := sb.Name + PersistentVolumeClaim
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pvcName,
			Namespace:       sb.Namespace,
			OwnerReferences: setOwnerReference(sb),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				"ReadWriteOnce",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
		},
	}

	return pvc
}
