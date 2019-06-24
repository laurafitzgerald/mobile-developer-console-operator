package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MobileDeveloperConsoleSpec defines the desired state of MobileDeveloperConsole
// +k8s:openapi-gen=true
type MobileDeveloperConsoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// MobileDeveloperConsoleStatus defines the observed state of MobileDeveloperConsole
// +k8s:openapi-gen=true
type MobileDeveloperConsoleStatus struct {
	Phase StatusPhase `json:"phase"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MobileDeveloperConsole is the Schema for the mobiledeveloperconsoles API
// +k8s:openapi-gen=true
type MobileDeveloperConsole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MobileDeveloperConsoleSpec   `json:"spec,omitempty"`
	Status MobileDeveloperConsoleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MobileDeveloperConsoleList contains a list of MobileDeveloperConsole
type MobileDeveloperConsoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MobileDeveloperConsole `json:"items"`
}

type StatusPhase string

var (
	PhaseEmpty     StatusPhase = ""
	PhaseComplete  StatusPhase = "Complete"
	PhaseProvision StatusPhase = "Provisioning"
)

func init() {
	SchemeBuilder.Register(&MobileDeveloperConsole{}, &MobileDeveloperConsoleList{})
}
