package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SocialBookStatus",type=boolean,JSONPath=`.status.SocialBook`
// +kubebuilder:printcolumn:name="MongoDBStatus",type=boolean,JSONPath=`.status.MongoDB`
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
	SocialBook bool `json:"configMap,omitempty"`
	MongoDB    bool `json:"mongo,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SocialBookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SocialBook `json:"items,omitempty"`
}
