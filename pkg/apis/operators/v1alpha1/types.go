package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SocialBook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SocialBookSpec   `json:"spec,omitempty"`
	Status SocialBookStatus `json:"status,omitempty"`
}

type SocialBookSpec struct {
	Replicas int32  `json:"replicas,omitempty"`
	UserName string `json:"username,omitempty"`
	Password string `json:"pwd,omitempty"`
	Port     string `json:"port,omitempty"`
}

type SocialBookStatus struct {
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
}

type SocialBookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SocialBook `json:"items,omitempty"`
}
