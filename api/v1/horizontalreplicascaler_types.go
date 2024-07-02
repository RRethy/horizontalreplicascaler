package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make" to regenerate code after modifying this file.

type ScaleTargetRef struct {
	// Group is the group of the target resource.
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

// ScalingRules defines the scaling rules for how many replicas to scaling up or down.
type ScalingRules struct {
	// StabilizationWindowSeconds is the number of seconds to wait before considering the system stable.
	// A stabilization window of 0 seconds means the replica suggestion will be applied immediately.
	// This may cause thrashing. A stabilization that is too long may cause the system to be unresponsive.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Default=0s
	StabilizationWindow metav1.Duration `json:"stabilizationWindowSeconds,omitempty"`

	// TODO: Something akin to HPA scaling policies.
}

// ScalingBehavior defines the scaling rules for scaling up and down.
type ScalingBehavior struct {
	// ScaleUp is the scaling behavior for scaling up.
	// +kubebuilder:validation:Optional
	ScaleUp *ScalingRules `json:"scaleUp,omitempty"`

	// ScaleDown is the scaling behavior for scaling down.
	// +kubebuilder:validation:Optional
	ScaleDown *ScalingRules `json:"scaleDown,omitempty"`
}

// TargetSec defines the target that should be scaled towards.
type TargetSec struct {
	// Type is the type of the target.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=pod-average;value
	Type string `json:"type"`

	// Value is the value of the target.
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// MetricSpec defines a metric to consider for scaling.
type MetricSpec struct {
	// Type is the type of metric to use.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=static;prometheus
	Type string `json:"type"`

	// Config is a map of configuration values for the metric.
	// +kubebuilder:validation:Optional
	Config map[string]string `json:"config"`

	// Target is the target specification for the metric.
	// +kubebuilder:validation:Required
	Target TargetSec `json:"target"`
}

// HorizontalReplicaScalerSpec defines the desired state of HorizontalReplicaScaler.
type HorizontalReplicaScalerSpec struct {
	// ScaleTargetRef points to the target resource to scale.
	// +kubebuilder:validation:Required
	ScaleTargetRef *ScaleTargetRef `json:"scaleTargetRef"`

	// MinReplicas is the lower limit for the number of replicas to which the target can be scaled.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MinReplicas int32 `json:"minReplicas"`

	// MaxReplicas is the upper limit for the number of replicas to which the target can be scaled.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MaxReplicas int32 `json:"maxReplicas"`

	// ScalingBehavior is the way in which we scale the target to the desired replicas.
	// +kubebuilder:validation:Optional
	ScalingBehavior *ScalingBehavior `json:"scalingBehavior,omitempty"`

	// PollingInterval is a best-effort target for how often the autoscaler should poll the metrics.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Default=30s
	PollingInterval metav1.Duration `json:"pollingInterval"`

	// Metrics is a list of metrics the autoscaler should use to scale the target.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Metrics []MetricSpec `json:"metrics"`
}

// HorizontalReplicaScalerStatus defines the observed state of HorizontalReplicaScaler.
type HorizontalReplicaScalerStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=all,shortName=hrs

// HorizontalReplicaScaler is the Schema for the horizontalreplicascalers API.
type HorizontalReplicaScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HorizontalReplicaScalerSpec   `json:"spec,omitempty"`
	Status HorizontalReplicaScalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HorizontalReplicaScalerList contains a list of HorizontalReplicaScaler.
type HorizontalReplicaScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HorizontalReplicaScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HorizontalReplicaScaler{}, &HorizontalReplicaScalerList{})
}
