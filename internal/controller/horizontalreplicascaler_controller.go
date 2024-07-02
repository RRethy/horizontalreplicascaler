package controller

import (
	"context"

	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/record"
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
	Recorder    record.EventRecorder
	ScaleClient scale.ScalesGetter
	PromAPI     promv1.API
}

// +kubebuilder:rbac:groups=scaling.rrethy.com,resources=horizontalreplicascalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=scaling.rrethy.com,resources=horizontalreplicascalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=scaling.rrethy.com,resources=horizontalreplicascalers/finalizers,verbs=update
// +kubebuilder:rbac:groups="*",resources="*/scale",verbs=get;update;patch

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

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalReplicaScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rrethyv1.HorizontalReplicaScaler{}).
		Complete(r)
}
