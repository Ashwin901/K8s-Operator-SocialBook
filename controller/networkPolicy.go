package controller

import (
	"github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newNetworkPolicy(sb *v1alpha1.SocialBook, appType string) *networkingv1.NetworkPolicy {
	if appType == MongoDB {
		return newMongoNetworkPolicy(sb)
	}
	return newSocialBookNetworkPolicy(sb)
}

func newMongoNetworkPolicy(sb *v1alpha1.SocialBook) *networkingv1.NetworkPolicy {
	port := intstr.FromInt(27017)
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + MongoDB + NetworkPolicy,
			Namespace:       sb.Namespace,
			OwnerReferences: setOwnerReference(sb),
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": sb.Name + MongoDB,
				},
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": sb.Name + SocialBook,
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port: &port,
						},
					},
				},
			},
		},
	}
}

func newSocialBookNetworkPolicy(sb *v1alpha1.SocialBook) *networkingv1.NetworkPolicy {
	port := intstr.FromInt(27017)
	return &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sb.Name + NetworkPolicy,
			Namespace:       sb.Namespace,
			OwnerReferences: setOwnerReference(sb),
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": sb.Name + SocialBook,
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					To: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": sb.Name + MongoDB,
								},
							},
						},
					},
					Ports: []networkingv1.NetworkPolicyPort{
						{
							Port: &port,
						},
					},
				},
			},
		},
	}
}
