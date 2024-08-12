package controller

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rrethyv1 "github.com/RRethy/horizontalreplicascaler/api/v1"
	"github.com/RRethy/horizontalreplicascaler/internal/stabilization"
)

const (
	// EventReasonFailedGetScaleSubresource is the reason for the event when the scale subresource cannot be retrieved.
	EventReasonFailedGetScaleSubresource = "FailedGetScaleSubresource"
)

type metricValue struct {
	metric rrethyv1.MetricSpec
	value  float64
}

// HorizontalReplicaScalerReconciler reconciles a HorizontalReplicaScaler object
type HorizontalReplicaScalerReconciler struct {
	client.Client
	Scheme                       *runtime.Scheme
	Recorder                     record.EventRecorder
	ScaleClient                  scale.ScalesGetter
	PromAPI                      promv1.API
	ScaleDownStabilizationWindow *stabilization.Window
	ScaleUpStabilizationWindow   *stabilization.Window
}

// +kubebuilder:rbac:groups=scaling.rrethy.com,resources=horizontalreplicascalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scaling.rrethy.com,resources=horizontalreplicascalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=scaling.rrethy.com,resources=horizontalreplicascalers/finalizers,verbs=update
// +kubebuilder:rbac:groups="*",resources="*/scale",verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
// This Reconcile method uses the ObjectReconciler interface from https://github.com/kubernetes-sigs/controller-runtime/pull/2592.
func (r *HorizontalReplicaScalerReconciler) Reconcile(ctx context.Context, horizontalReplicaScaler *rrethyv1.HorizontalReplicaScaler) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if !horizontalReplicaScaler.DeletionTimestamp.IsZero() {
		// The object is being deleted, don't do anything.
		return ctrl.Result{}, nil
	}

	defer func() {
		err := r.Status().Update(ctx, horizontalReplicaScaler)
		if err != nil {
			log.Error(err, "updating status")
		}
	}()

	pollingInterval := horizontalReplicaScaler.Spec.PollingInterval.Duration

	scaleSubresource, err := r.getScaleSubresource(ctx, horizontalReplicaScaler)
	if err != nil {
		log.Error(err, "getting scale subresource")
		r.Recorder.Event(horizontalReplicaScaler, corev1.EventTypeWarning, EventReasonFailedGetScaleSubresource, err.Error())
		return ctrl.Result{RequeueAfter: pollingInterval}, client.IgnoreNotFound(err)
	}

	metricResults, err := r.getMetricValues(horizontalReplicaScaler)
	if err != nil {
		log.Error(err, "getting metric results")
		return ctrl.Result{RequeueAfter: pollingInterval}, err
	}

	desiredReplicas := r.getMaxMetricValues(metricResults)
	desiredReplicas = r.applyScalingBehavior(horizontalReplicaScaler, scaleSubresource.Spec.Replicas, desiredReplicas)
	desiredReplicas = r.applyMinMaxReplicas(horizontalReplicaScaler, desiredReplicas)

	err = r.updateScaleSubresource(ctx, horizontalReplicaScaler, scaleSubresource, desiredReplicas)
	if err != nil {
		log.Error(err, "updating scale subresource")
		return ctrl.Result{RequeueAfter: pollingInterval}, nil
	}

	return ctrl.Result{RequeueAfter: pollingInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalReplicaScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rrethyv1.HorizontalReplicaScaler{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})).
		Complete(reconcile.AsReconciler(mgr.GetClient(), r))
}

// getScaleSubresource returns the scale subresource for the target resource.
func (r *HorizontalReplicaScalerReconciler) getScaleSubresource(ctx context.Context, horizontalReplicaScaler *rrethyv1.HorizontalReplicaScaler) (*autoscalingv1.Scale, error) {
	gr := schema.GroupResource{Group: horizontalReplicaScaler.Spec.ScaleTargetRef.Group, Resource: horizontalReplicaScaler.Spec.ScaleTargetRef.Kind}
	return r.ScaleClient.Scales(horizontalReplicaScaler.Namespace).Get(ctx, gr, horizontalReplicaScaler.Spec.ScaleTargetRef.Name, metav1.GetOptions{})
}

// getMetricResults returns the result of calculating each metric.
func (r *HorizontalReplicaScalerReconciler) getMetricValues(horizontalReplicaScaler *rrethyv1.HorizontalReplicaScaler) ([]metricValue, error) {
	var values []metricValue
	for _, metric := range horizontalReplicaScaler.Spec.Metrics {
		value, err := r.getMetricValue(metric)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// getMetricResult returns the result of calculating a metric.
func (r *HorizontalReplicaScalerReconciler) getMetricValue(metric rrethyv1.MetricSpec) (metricValue, error) {
	switch metric.Type {
	case "static":
		target, err := strconv.ParseFloat(metric.Target.Value, 64)
		if err != nil {
			return metricValue{}, fmt.Errorf("failed parsing target value %s: %w", metric.Target, err)
		}
		return metricValue{metric: metric, value: target}, nil
	default:
		return metricValue{}, fmt.Errorf("unknown metric type %s", metric.Type)
	}
}

func (r *HorizontalReplicaScalerReconciler) getMaxMetricValues(metricValues []metricValue) int32 {
	var maxResult float64
	for _, metricResult := range metricValues {
		if metricResult.value > maxResult {
			maxResult = metricResult.value
		}
	}
	return int32(maxResult)
}

func (r *HorizontalReplicaScalerReconciler) applyMinMaxReplicas(horizontalReplicaScaler *rrethyv1.HorizontalReplicaScaler, desiredReplicas int32) int32 {
	if desiredReplicas < horizontalReplicaScaler.Spec.MinReplicas {
		return horizontalReplicaScaler.Spec.MinReplicas
	}
	if desiredReplicas > horizontalReplicaScaler.Spec.MaxReplicas {
		return horizontalReplicaScaler.Spec.MaxReplicas
	}
	return desiredReplicas
}

func (r *HorizontalReplicaScalerReconciler) applyScalingBehavior(horizontalReplicaScaler *rrethyv1.HorizontalReplicaScaler, currentReplicas, desiredReplicas int32) int32 {
	stabilizationWindowKey := stabilization.KeyFor(
		horizontalReplicaScaler.Namespace,
		horizontalReplicaScaler.Name,
		horizontalReplicaScaler.Spec.ScaleTargetRef.Name,
		horizontalReplicaScaler.Spec.ScaleTargetRef.Kind,
		horizontalReplicaScaler.Spec.ScaleTargetRef.Group,
	)

	stabilizedDownScale, ok := r.ScaleDownStabilizationWindow.Stabilize(stabilizationWindowKey, desiredReplicas, horizontalReplicaScaler.Spec.ScalingBehavior.ScaleDown.StabilizationWindow.Duration)
	if !ok {
		stabilizedDownScale = currentReplicas
	}

	stabilizedUpScale, ok := r.ScaleUpStabilizationWindow.Stabilize(stabilizationWindowKey, desiredReplicas, horizontalReplicaScaler.Spec.ScalingBehavior.ScaleUp.StabilizationWindow.Duration)
	if !ok {
		stabilizedUpScale = currentReplicas
	}

	if desiredReplicas < currentReplicas {
		return slices.Min([]int32{stabilizedDownScale, currentReplicas})
	} else if desiredReplicas > currentReplicas {
		return slices.Max([]int32{stabilizedUpScale, currentReplicas})
	}
	return currentReplicas
}

func (r *HorizontalReplicaScalerReconciler) updateScaleSubresource(ctx context.Context, horizontalReplicaScaler *rrethyv1.HorizontalReplicaScaler, scaleSubresource *autoscalingv1.Scale, desiredReplicas int32) error {
	var err error
	if horizontalReplicaScaler.Spec.DryRun {
		horizontalReplicaScaler.Status.DesiredReplicas = desiredReplicas
	} else {
		scaleSubresource.Spec.Replicas = desiredReplicas
		gr := schema.GroupResource{Group: horizontalReplicaScaler.Spec.ScaleTargetRef.Group, Resource: horizontalReplicaScaler.Spec.ScaleTargetRef.Kind}
		_, err = r.ScaleClient.Scales(horizontalReplicaScaler.Namespace).Update(ctx, gr, scaleSubresource, metav1.UpdateOptions{})
	}
	return err
}
