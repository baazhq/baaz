package app_controller

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/internal/predicates"
	"datainfra.io/baaz/pkg/aws/eks"
	"datainfra.io/baaz/pkg/utils"
)

const (
	applicationFinalizer = "application.datainfra.io/finalizer"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// reconcile time duration, defaults to 10s
	ReconcileWait time.Duration
	Predicates    predicate.Predicate
	Recorder      record.EventRecorder
	CustomerName  string
	EnablePrivate bool
}

func NewApplicationReconciler(mgr ctrl.Manager, enablePrivate bool, customerName string) *ApplicationReconciler {
	initLogger := ctrl.Log.WithName("controllers").WithName("application")
	return &ApplicationReconciler{
		Client:        mgr.GetClient(),
		Log:           initLogger,
		Scheme:        mgr.GetScheme(),
		ReconcileWait: lookupReconcileTime(),
		Predicates:    predicates.GetPredicates(enablePrivate, customerName),
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

	dpObj := &v1.DataPlanesList{}
	err = r.List(ctx, dpObj, &client.ListOptions{})
	if err != nil {
		return ctrl.Result{}, err
	}

	var dataplane v1.DataPlanes
	for _, dp := range dpObj.Items {
		if dp.GetName() == applicationObj.Spec.Dataplane {
			dataplane = dp
		}
	}

	// check for deletion time stamp
	if applicationObj.DeletionTimestamp != nil {
		// object is going to be deleted
		return r.reconcileDelete(ctx, applicationObj, &dataplane)
	}

	// if it is normal reconcile, then add finalizer if not already
	if !controllerutil.ContainsFinalizer(applicationObj, applicationFinalizer) {
		controllerutil.AddFinalizer(applicationObj, applicationFinalizer)
		if err := r.Update(ctx, applicationObj); err != nil {
			return ctrl.Result{}, err
		}
	}

	if applicationObj.Status.Phase == "" {
		if _, _, err := utils.PatchStatus(ctx, r.Client, applicationObj, func(obj client.Object) client.Object {
			in := obj.(*v1.Applications)
			in.Status.Phase = v1.ApplicationPhase(v1.PendingA)
			return in
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.do(ctx, applicationObj, &dataplane); err != nil {
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
		WithEventFilter(r.Predicates).
		Complete(r)
}

func lookupReconcileTime() time.Duration {
	val, exists := os.LookupEnv("RECONCILE_WAIT")
	if !exists {
		return time.Second * 10
	} else {
		v, err := time.ParseDuration(val)
		if err != nil {
			klog.Error(err, err.Error())
			// Exit Program if not valid
			os.Exit(1)
		}
		return v
	}
}

func (r *ApplicationReconciler) reconcileDelete(ctx context.Context, app *v1.Applications, dataplane *v1.DataPlanes) (ctrl.Result, error) {
	// update phase to terminating
	_, _, err := utils.PatchStatus(ctx, r.Client, app, func(obj client.Object) client.Object {
		in := obj.(*v1.Applications)
		in.Status.Phase = v1.UninstallingA
		return in
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	eksIc := eks.NewEks(ctx, dataplane)
	eksClientSet, err := eksIc.GetEksClientSet()
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}

	applications := NewApplication(ctx, app, dataplane, r.Client, eksClientSet)

	if err := applications.UninstallApplications(); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}

	// remove our finalizer from the list and update it.
	controllerutil.RemoveFinalizer(app, applicationFinalizer)
	klog.Infof("Uninstalled Application [%s]", app.GetName())
	if err := r.Client.Update(ctx, app.DeepCopyObject().(*v1.Applications)); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
