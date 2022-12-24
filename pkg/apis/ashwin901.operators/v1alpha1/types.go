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
	Replicas      int32  `json:"replicas,omitempty"`      // number of pods for socialbook image
	MongoUsername string `json:"mongoUsername,omitempty"` // mongodb username
	MongoPassword string `json:"mongoPassword,omitempty"` // mongodb password
	Port          string `json:"port,omitempty"`          // container port for socialbook
	JwtSecret     string `json:"jwtSecret,omitempty"`     // used to generate jwt (any random string)
	EmailId       string `json:"email,omitempty"`         // email id used to send verification emails
	Password      string `json:"password,omitempty"`      // pwd of email id
	ClientUrl     string `json:"clientUrl,omitempty"`     // redirection url used during email verification
	StripeApiKey  string `json:"stripeApiKey,omitempty"`  // stripe api key used for payments
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
