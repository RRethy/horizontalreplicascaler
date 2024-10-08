package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "make" to regenerate code after modifying this file.

// MetricType is the type of metric to use.
type MetricType string

const (
	StaticMetricType     MetricType = "static"
	PrometheusMetricType MetricType = "prometheus"
)

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
	// For scaling up, this should be 0s unless the system is known to be extremely unstable.
	// Stabilization windows are cleared when the controller restarts to error on the side of caution.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Default=0s
	StabilizationWindow metav1.Duration `json:"stabilizationWindowSeconds,omitempty"`

	// TODO: Something akin to HPA scaling policies.
}

// ScalingBehavior defines the scaling rules for scaling up and down.
type ScalingBehavior struct {
	// ScaleUp is the scaling behavior for scaling up.
	// +kubebuilder:validation:Optional
	ScaleUp ScalingRules `json:"scaleUp,omitempty"`

	// ScaleDown is the scaling behavior for scaling down.
	// +kubebuilder:validation:Optional
	ScaleDown ScalingRules `json:"scaleDown,omitempty"`
}

// Fallback defines the fallback behavior when failures occur.
type Fallback struct {
	// Replicas is the number of replicas to scale to when metrics fail.
	// +kubebuilder:validation:Required
	Replicas int32 `json:"replicas"`

	// Threshold is the number of consecutive failures before the fallback is triggered.
	Threshold int32 `json:"threshold"`
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
	Type MetricType `json:"type"`

	// Config is a map of configuration values for the metric.
	// +kubebuilder:validation:Optional
	Config map[string]string `json:"config"`

	// Target is the target specification for the metric.
	// +kubebuilder:validation:Required
	Target TargetSec `json:"target"`
}

// HorizontalReplicaScalerSpec defines the desired state of HorizontalReplicaScaler.
type HorizontalReplicaScalerSpec struct {
	// DryRun is a flag to indicate if the target workload should not actually be scaled.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Default=false
	DryRun bool `json:"dryRun"`

	// ScaleTargetRef points to the target resource to scale.
	// +kubebuilder:validation:Required
	ScaleTargetRef ScaleTargetRef `json:"scaleTargetRef"`

	// MinReplicas is the lower limit for the number of replicas to which the target can be scaled.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MinReplicas int32 `json:"minReplicas"`

	// MaxReplicas is the upper limit for the number of replicas to which the target can be scaled.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	MaxReplicas int32 `json:"maxReplicas"`

	// PollingInterval is a best-effort target for how often the autoscaler should poll the metrics.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Default=30s
	PollingInterval metav1.Duration `json:"pollingInterval"`

	// ScalingBehavior is the way in which we scale the target to the desired replicas.
	// +kubebuilder:validation:Optional
	ScalingBehavior ScalingBehavior `json:"scalingBehavior,omitempty"`

	// Fallback is the fallback behavior for the autoscaler when metrics fail.
	// The fallback applies to each metric individually.
	// +kubebuilder:validation:Optional
	Fallback *Fallback `json:"fallback,omitempty"`

	// Metrics is a list of metrics the autoscaler should use to scale the target.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Metrics []MetricSpec `json:"metrics"`
}

// ScaleEvent defines an event in the stabilization window for the scaling rule.
type ScaleEvent struct {
	// Value is the replica value for the scale event.
	// +kubebuilder:validation:Required
	Value int32 `json:"value"`

	// Timestamp is the timestamp of the scale event.
	// +kubebuilder:validation:Required
	Timestamp metav1.Time `json:"timestamp"`
}

// HorizontalReplicaScalerStatus defines the observed state of HorizontalReplicaScaler.
type HorizontalReplicaScalerStatus struct {
	// DesiredReplicas is the number of replicas the target should be scaled to.
	// +kubebuilder:validation:Optional
	DesiredReplicas int32 `json:"desiredReplicas"`
}

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
