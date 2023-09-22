package app_controller

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/pkg/utils"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// reconcile time duration, defaults to 10s
	ReconcileWait time.Duration
	Recorder      record.EventRecorder
}

func NewApplicationReconciler(mgr ctrl.Manager) *ApplicationReconciler {
	initLogger := ctrl.Log.WithName("controllers").WithName("application")
	return &ApplicationReconciler{
		Client:        mgr.GetClient(),
		Log:           initLogger,
		Scheme:        mgr.GetScheme(),
		ReconcileWait: lookupReconcileTime(initLogger),
		Recorder:      mgr.GetEventRecorderFor("applications-controller"),
	}
}

// +kubebuilder:rbac:groups=datainfra.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datainfra.io,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=datainfra.io,resources=applications/finalizers,verbs=update
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	applicationObj := &v1.Applications{}
	err := r.Get(ctx, req.NamespacedName, applicationObj)
	if err != nil {
		return ctrl.Result{}, err
	}

	dpObj := &v1.DataPlanes{}
	var namespace string

	if strings.Contains(applicationObj.Spec.EnvRef, "shared") {
		namespace = "shared"
	} else {
		namespace = applicationObj.Namespace
	}

	err = r.Get(ctx, types.NamespacedName{Name: applicationObj.Spec.EnvRef, Namespace: namespace}, dpObj)
	if err != nil {
		return ctrl.Result{}, err
	}

	if applicationObj.Status.Phase == "" {
		if _, _, err := utils.PatchStatus(ctx, r.Client, applicationObj, func(obj client.Object) client.Object {
			in := obj.(*v1.Applications)
			in.Status.Phase = v1.ApplicationPhase(v1.PendingA)
			//	in.Status.ApplicationCurrentSpec = applicationObj
			return in
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.do(ctx, applicationObj, dpObj); err != nil {
		klog.Errorf("failed to reconcile application: reason: %s", err.Error())
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	} else {
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Applications{}).
		Complete(r)
}

func lookupReconcileTime(log logr.Logger) time.Duration {
	val, exists := os.LookupEnv("RECONCILE_WAIT")
	if !exists {
		return time.Second * 10
	} else {
		v, err := time.ParseDuration(val)
		if err != nil {
			log.Error(err, err.Error())
			// Exit Program if not valid
			os.Exit(1)
		}
		return v
	}
}
