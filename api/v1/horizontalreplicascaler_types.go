package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make" to regenerate code after modifying this file

type ScaleTargetRef struct {
	// Group is the group of the target resource
	// +kubebuilder:validation:Required
	Group string `json:"group"`

	// Kind is a string value representing the REST resource this object represents.
	// Servers may infer this from the endpoint the client submits requests to.
	// +kubebuilder:validation:Required
	Kind string `json:"kind"`

	// Name is the name of the resource being referred to by the scale target.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

type MetricSpec struct {
	// Type is the type of metric to use
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Static
	Type string `json:"type"`

	// Config is a map of configuration values for the metric
	// +kubebuilder:validation:Required
	Config map[string]string `json:"config"`

	// Target is the target value for the metric
	// +kubebuilder:validation:Required
	Target string `json:"target"`
}

// HorizontalReplicaScalerSpec defines the desired state of HorizontalReplicaScaler
type HorizontalReplicaScalerSpec struct {
	// ScaleTargetRef points to the target resource to scale
	// +kubebuilder:validation:Required
	ScaleTargetRef *ScaleTargetRef `json:"scaleTargetRef"`

	// MinReplicas is the lower limit for the number of replicas to which the target can be scaled
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MinReplicas int32 `json:"minReplicas"`

	// MaxReplicas is the upper limit for the number of replicas to which the target can be scaled
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MaxReplicas int32 `json:"maxReplicas"`

	// Metrics is a list of metrics the autoscaler should use to scale the target
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Metrics []MetricSpec `json:"metrics"`
}

// HorizontalReplicaScalerStatus defines the observed state of HorizontalReplicaScaler
type HorizontalReplicaScalerStatus struct {
	// Info is a human-readable message about the status of the autoscaler
	Info string `json:"info"`
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
