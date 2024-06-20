package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HorizontalReplicaScalerSpec defines the desired state of HorizontalReplicaScaler
type HorizontalReplicaScalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of HorizontalReplicaScaler. Edit horizontalreplicascaler_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// HorizontalReplicaScalerStatus defines the observed state of HorizontalReplicaScaler
type HorizontalReplicaScalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HorizontalReplicaScaler is the Schema for the horizontalreplicascalers API
type HorizontalReplicaScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HorizontalReplicaScalerSpec   `json:"spec,omitempty"`
	Status HorizontalReplicaScalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HorizontalReplicaScalerList contains a list of HorizontalReplicaScaler
type HorizontalReplicaScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HorizontalReplicaScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HorizontalReplicaScaler{}, &HorizontalReplicaScalerList{})
}
