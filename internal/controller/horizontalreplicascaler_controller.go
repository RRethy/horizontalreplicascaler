package controller

import (
	"context"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	rrethyv1 "github.com/RRethy/horizontalrpelicascaler/api/v1"
)

const (
	EventReasonFailedGetScaleSubresource = "FailedGetScaleSubresource"
)

type MetricResult struct {
	Type  string
	Value float64
}

// HorizontalReplicaScalerReconciler reconciles a HorizontalReplicaScaler object
type HorizontalReplicaScalerReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
	ScaleClient scale.ScalesGetter
	PromAPI     promv1.API
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
	nsName := types.NamespacedName{Namespace: horizontalReplicaScaler.Namespace, Name: horizontalReplicaScaler.Name}
	log := log.FromContext(ctx).WithValues("horizontalreplicascaler", nsName)

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

	scaleSubresource, err := r.getScaleSubresource(ctx, horizontalReplicaScaler)
	if err != nil {
		log.Error(err, "failed to get scale subresource")
		r.Recorder.Event(horizontalReplicaScaler, corev1.EventTypeWarning, EventReasonFailedGetScaleSubresource, err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("scale subresource", ".spec.replicas", scaleSubresource.Spec.Replicas)

	log.Info("reconciling HorizontalReplicaScaler")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalReplicaScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rrethyv1.HorizontalReplicaScaler{}).
		Complete(reconcile.AsReconciler(mgr.GetClient(), r))
}

// getScaleSubresource returns the scale subresource for the target resource.
func (r *HorizontalReplicaScalerReconciler) getScaleSubresource(ctx context.Context, horizontalReplicaScaler *rrethyv1.HorizontalReplicaScaler) (*autoscalingv1.Scale, error) {
	gr := schema.GroupResource{Group: horizontalReplicaScaler.Spec.ScaleTargetRef.Group, Resource: horizontalReplicaScaler.Spec.ScaleTargetRef.Kind}
	return r.ScaleClient.Scales(horizontalReplicaScaler.Namespace).Get(ctx, gr, horizontalReplicaScaler.Spec.ScaleTargetRef.Name, metav1.GetOptions{})
}
