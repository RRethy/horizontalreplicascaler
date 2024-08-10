//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Fallback) DeepCopyInto(out *Fallback) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Fallback.
func (in *Fallback) DeepCopy() *Fallback {
	if in == nil {
		return nil
	}
	out := new(Fallback)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalReplicaScaler) DeepCopyInto(out *HorizontalReplicaScaler) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalReplicaScaler.
func (in *HorizontalReplicaScaler) DeepCopy() *HorizontalReplicaScaler {
	if in == nil {
		return nil
	}
	out := new(HorizontalReplicaScaler)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HorizontalReplicaScaler) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalReplicaScalerList) DeepCopyInto(out *HorizontalReplicaScalerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HorizontalReplicaScaler, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalReplicaScalerList.
func (in *HorizontalReplicaScalerList) DeepCopy() *HorizontalReplicaScalerList {
	if in == nil {
		return nil
	}
	out := new(HorizontalReplicaScalerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HorizontalReplicaScalerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalReplicaScalerSpec) DeepCopyInto(out *HorizontalReplicaScalerSpec) {
	*out = *in
	out.ScaleTargetRef = in.ScaleTargetRef
	out.PollingInterval = in.PollingInterval
	out.ScalingBehavior = in.ScalingBehavior
	if in.Fallback != nil {
		in, out := &in.Fallback, &out.Fallback
		*out = new(Fallback)
		**out = **in
	}
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]MetricSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalReplicaScalerSpec.
func (in *HorizontalReplicaScalerSpec) DeepCopy() *HorizontalReplicaScalerSpec {
	if in == nil {
		return nil
	}
	out := new(HorizontalReplicaScalerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HorizontalReplicaScalerStatus) DeepCopyInto(out *HorizontalReplicaScalerStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HorizontalReplicaScalerStatus.
func (in *HorizontalReplicaScalerStatus) DeepCopy() *HorizontalReplicaScalerStatus {
	if in == nil {
		return nil
	}
	out := new(HorizontalReplicaScalerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetricSpec) DeepCopyInto(out *MetricSpec) {
	*out = *in
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Target = in.Target
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetricSpec.
func (in *MetricSpec) DeepCopy() *MetricSpec {
	if in == nil {
		return nil
	}
	out := new(MetricSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScaleEvent) DeepCopyInto(out *ScaleEvent) {
	*out = *in
	in.Timestamp.DeepCopyInto(&out.Timestamp)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScaleEvent.
func (in *ScaleEvent) DeepCopy() *ScaleEvent {
	if in == nil {
		return nil
	}
	out := new(ScaleEvent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScaleTargetRef) DeepCopyInto(out *ScaleTargetRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScaleTargetRef.
func (in *ScaleTargetRef) DeepCopy() *ScaleTargetRef {
	if in == nil {
		return nil
	}
	out := new(ScaleTargetRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScalingBehavior) DeepCopyInto(out *ScalingBehavior) {
	*out = *in
	out.ScaleUp = in.ScaleUp
	out.ScaleDown = in.ScaleDown
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScalingBehavior.
func (in *ScalingBehavior) DeepCopy() *ScalingBehavior {
	if in == nil {
		return nil
	}
	out := new(ScalingBehavior)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScalingRules) DeepCopyInto(out *ScalingRules) {
	*out = *in
	out.StabilizationWindow = in.StabilizationWindow
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScalingRules.
func (in *ScalingRules) DeepCopy() *ScalingRules {
	if in == nil {
		return nil
	}
	out := new(ScalingRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TargetSec) DeepCopyInto(out *TargetSec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TargetSec.
func (in *TargetSec) DeepCopy() *TargetSec {
	if in == nil {
		return nil
	}
	out := new(TargetSec)
	in.DeepCopyInto(out)
	return out
}
