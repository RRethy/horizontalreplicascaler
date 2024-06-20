package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/scale"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	rrethycomv1 "github.com/RRethy/horizontalrpelicascaler/api/v1"
)

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
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HorizontalReplicaScaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (r *HorizontalReplicaScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HorizontalReplicaScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rrethycomv1.HorizontalReplicaScaler{}).
		Complete(r)
}
