package controller

import (
	"strconv"

	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newDeployment(sb *v1alpha1.SocialBook, appType string) *appsv1.Deployment {
	if appType == MongoDB {
		return newMongoDeployment(sb)
	}

	return newSocialBookDeployment(sb)
}

func newMongoDeployment(sb *v1alpha1.SocialBook) *appsv1.Deployment {

	var replicas int32
	replicas = 1

	depName := sb.Name + MongoDB
	cmName := sb.Name + ConfigMap

	// mongo db deployment
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            depName,
			Namespace:       sb.Namespace,
			OwnerReferences: setOwnerReference(sb),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": sb.Name + MongoDB,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: sb.Name + MongoDB,
					Labels: map[string]string{
						"app": sb.Name + MongoDB,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: sb.Name + PersistentVolume,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: sb.Name + PersistentVolumeClaim,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  sb.Name + MongoDB,
							Image: "mongo",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 27017,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "MONGO_INITDB_ROOT_USERNAME",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "mongo-root-username",
										},
									},
								},
								{
									Name: "MONGO_INITDB_ROOT_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "mongo-root-password",
										},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      sb.Name + PersistentVolume,
									MountPath: "/data/db",
								},
							},
						},
					},
				},
			},
		},
	}

	return dep
}

func newSocialBookDeployment(sb *v1alpha1.SocialBook) *appsv1.Deployment {
	portNumber, _ := strconv.Atoi(sb.Spec.Port)
	cmName := sb.Name + ConfigMap

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name,
			Namespace:       sb.Namespace,
			OwnerReferences: setOwnerReference(sb),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &sb.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": sb.Name + SocialBook,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: sb.Name + SocialBook,
					Labels: map[string]string{
						"app": sb.Name + SocialBook,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  sb.Name,
							Image: Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: int32(portNumber),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "PORT",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "port",
										},
									},
								},
								{
									Name: "MONGODB_URI",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "mongodb-uri",
										},
									},
								},
								{
									Name: "SECRET",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "secret",
										},
									},
								},
								{
									Name: "STRIPE_API_KEY",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "stripe-api-key",
										},
									},
								},
								{
									Name: "USER_EMAIL",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "user-email",
										},
									},
								},
								{
									Name: "USER_PASSWORD",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "user-pwd",
										},
									},
								},
								{
									Name: "CLIENT_URL",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: cmName,
											},
											Key: "client-url",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return dep
}
