package controller

import (
	"context"
	"fmt"
	"strconv"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	rrethyv1 "github.com/RRethy/horizontalrpelicascaler/api/v1"
)

type MetricResult struct {
	Type  string
	Value float64
}

// HorizontalReplicaScalerReconciler reconciles a HorizontalReplicaScaler object
type HorizontalReplicaScalerReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	ScaleClient scale.ScalesGetter
}

// +kubebuilder:rbac:groups=rrethy.com.rrethy.com,resources=horizontalreplicascalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rrethy.com.rrethy.com,resources=horizontalreplicascalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=rrethy.com.rrethy.com,resources=horizontalreplicascalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *HorizontalReplicaScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("horizontalreplicascaler", req.NamespacedName)

	var horizontalReplicaScaler rrethyv1.HorizontalReplicaScaler
	if err := r.Get(ctx, req.NamespacedName, &horizontalReplicaScaler); err != nil {
		log.Error(err, "unable to fetch HorizontalReplicaScaler")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !horizontalReplicaScaler.DeletionTimestamp.IsZero() {
		// The object is being deleted, don't do anything.
		return ctrl.Result{}, nil
	}

	log.Info("reconciling horizontalreplicascaler", "name", horizontalReplicaScaler.Name)

	scaleSubresource, err := r.getScaleSubresource(&horizontalReplicaScaler)
	log.Info("getTargetScale", "targetScale", scaleSubresource, "err", err)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("targetScale not found", "targetScale", horizontalReplicaScaler.Spec.ScaleTargetRef.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// TODO: go through each metric and calculate the desired number of replicas
	metricResults, err := r.getMetricResults(horizontalReplicaScaler)
	log.Info("getMetricResults", "metricResults", metricResults, "err", err)

	// TODO: calculate the max(metric_desired_replicas[])
	var desiredReplicas int32
	for _, metricResult := range metricResults {
		if int32(metricResult.Value) > desiredReplicas {
			desiredReplicas = int32(metricResult.Value)
		}
	}
	log.Info("calculated max of metrics", "desiredReplicas", desiredReplicas)

	// TODO: apply min_replicas and max_replicas constraints
	// TODO: update the scale subresource with the desired number of replicas

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalReplicaScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rrethyv1.HorizontalReplicaScaler{}).
		Complete(r)
}

// getScaleSubresource returns the Scale subresource for the HorizontalReplicaScaler's target.
func (r *HorizontalReplicaScalerReconciler) getScaleSubresource(horizontalReplicaScaler *rrethyv1.HorizontalReplicaScaler) (*autoscalingv1.Scale, error) {
	gr := schema.GroupResource{Group: horizontalReplicaScaler.Spec.ScaleTargetRef.Group, Resource: horizontalReplicaScaler.Spec.ScaleTargetRef.Kind}
	return r.ScaleClient.Scales(horizontalReplicaScaler.Namespace).
		Get(context.TODO(), gr, horizontalReplicaScaler.Spec.ScaleTargetRef.Name, metav1.GetOptions{})
}

// getMetricResults returns the result of calculating each metric.
func (r *HorizontalReplicaScalerReconciler) getMetricResults(horizontalReplicaScaler rrethyv1.HorizontalReplicaScaler) ([]MetricResult, error) {
	var metricResults []MetricResult
	for _, metric := range horizontalReplicaScaler.Spec.Metrics {
		metricResult, err := r.getMetricResult(metric)
		if err != nil {
			return nil, err
		}
		metricResults = append(metricResults, metricResult)
	}
	return metricResults, nil
}

// getMetricResult returns the result of calculating a metric.
func (r *HorizontalReplicaScalerReconciler) getMetricResult(metric rrethyv1.MetricSpec) (MetricResult, error) {
	switch metric.Type {
	case "Static":
		if targetStr, ok := metric.Config["target"]; ok {
			// Parse the target value as a float64
			target, err := strconv.ParseFloat(targetStr, 64)
			if err != nil {
				return MetricResult{}, fmt.Errorf("parsing target value %s: %w", targetStr, err)
			}
			return MetricResult{Type: metric.Type, Value: target}, nil
		}
		return MetricResult{}, fmt.Errorf("missing target value in Static metric spec")
	default:
		return MetricResult{}, fmt.Errorf("unknown metric type %s", metric.Type)
	}
}
