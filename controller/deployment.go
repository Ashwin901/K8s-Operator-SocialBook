package controller

import (
	"strconv"

	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newMongoDeployment(sb *v1alpha1.SocialBook) *appsv1.Deployment {

	var replicas int32
	replicas = 1

	depName := sb.Name + "-mongodb"
	cmName := sb.Name + "-cm"

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
					"app": "mongodb-" + sb.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mongodb-" + sb.Name,
					Labels: map[string]string{
						"app": "mongodb-" + sb.Name,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "mongo-volume-" + sb.Name,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: sb.Name + "-mongo-pvc",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "mongodb",
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
									Name:      "mongo-volume-" + sb.Name,
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
	cmName := sb.Name + "-cm"

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
					"app": "socialbook-" + sb.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "socialbook-" + sb.Name,
					Labels: map[string]string{
						"app": "socialbook-" + sb.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "socialbook",
							Image: "ashwin901/social-book-server",
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
