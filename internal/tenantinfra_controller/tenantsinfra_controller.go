package tenantinfra_controller

import (
	"context"
	"os"
	"time"

	v1 "github.com/baazhq/baaz/api/v1/types"
	"github.com/baazhq/baaz/internal/predicates"
	"github.com/baazhq/baaz/pkg/aws/eks"
	"github.com/baazhq/baaz/pkg/store"
	"github.com/baazhq/baaz/pkg/utils"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var tenantsFinalizer = "tenantsinfra.baaz.dev/finalizer"

// TenantsInfraReconciler reconciles a Tenants object
type TenantsInfraReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// reconcile time duration, defaults to 10s
	ReconcileWait time.Duration
	Recorder      record.EventRecorder
	Predicates    predicate.Predicate
	NgStore       store.Store
	CustomerName  string
	EnablePrivate bool
}

func NewTenantsInfraReconciler(mgr ctrl.Manager, enablePrivate bool, customerName string) *TenantsInfraReconciler {
	initLogger := ctrl.Log.WithName("controllers").WithName("tenant_infra")
	return &TenantsInfraReconciler{
		Client:        mgr.GetClient(),
		Log:           initLogger,
		Scheme:        mgr.GetScheme(),
		ReconcileWait: lookupReconcileTime(initLogger),
		Recorder:      mgr.GetEventRecorderFor("tenantinfra-controller"),
		Predicates:    predicates.GetPredicates(enablePrivate, customerName),
		NgStore:       store.NewInternalStore(),
	}
}

//+kubebuilder:rbac:groups=baaz.dev,resources=tenantsinfra,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=baaz.dev,resources=tenantsinfra/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=baaz.dev,resources=tenantsinfra/finalizers,verbs=update

func (r *TenantsInfraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	tenantInfraObj := &v1.TenantsInfra{}
	err := r.Get(ctx, req.NamespacedName, tenantInfraObj)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dataplane := &v1.DataPlanes{}
	err = r.Get(ctx, k8stypes.NamespacedName{Name: tenantInfraObj.Spec.Dataplane, Namespace: req.Namespace}, dataplane)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	klog.Infof("Reconciling Tenants Infra Objects: %s/%s", tenantInfraObj.Namespace, tenantInfraObj.Name)

	if tenantInfraObj.DeletionTimestamp != nil {
		// object is going to be deleted
		awsEnv := awsEnv{
			ctx:          ctx,
			dp:           dataplane,
			tenantsInfra: tenantInfraObj,
			eksIC:        eks.NewEks(ctx, dataplane),
			client:       r.Client,
			store:        r.NgStore,
		}

		return r.reconcileDelete(&awsEnv)
	}

	// if it is normal reconcile, then add finalizer if not already
	if !controllerutil.ContainsFinalizer(tenantInfraObj, tenantsFinalizer) {
		controllerutil.AddFinalizer(tenantInfraObj, tenantsFinalizer)
		if err := r.Update(ctx, tenantInfraObj); err != nil {
			return ctrl.Result{}, err
		}
	}
	// If first time reconciling set status to pending
	if tenantInfraObj.Status.Phase == "" {
		if _, _, err := utils.PatchStatus(ctx, r.Client, tenantInfraObj, func(obj client.Object) client.Object {
			in := obj.(*v1.TenantsInfra)
			in.Status.Phase = v1.PendingT
			return in
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.do(ctx, tenantInfraObj, dataplane); err != nil {
		if _, _, patchErr := utils.PatchStatus(ctx, r.Client, tenantInfraObj, func(obj client.Object) client.Object {
			in := obj.(*v1.TenantsInfra)
			in.Status.Phase = v1.FailedT
			return in
		}); patchErr != nil {
			return ctrl.Result{}, patchErr
		}
		klog.Errorf("failed to reconcile tenant: reason: %s", err.Error())
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	} else {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

}

// SetupWithManager sets up the controller with the Manager.
func (r *TenantsInfraReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.TenantsInfra{}).
		WithEventFilter(r.Predicates).
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

func (r *TenantsInfraReconciler) reconcileDelete(ae *awsEnv) (ctrl.Result, error) {
	// update phase to terminating
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.tenantsInfra, func(obj client.Object) client.Object {
		in := obj.(*v1.TenantsInfra)
		in.Status.Phase = v1.TerminatingT
		return in
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	for ng, ngStatus := range ae.tenantsInfra.Status.NodegroupStatus {

		if ngStatus.Status != "DELETING" {
			_, err := ae.eksIC.DeleteNodeGroup(ng)
			if err != nil {
				return ctrl.Result{}, err
			}
			// update status with current nodegroup status
			_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.tenantsInfra, func(obj client.Object) client.Object {
				in := obj.(*v1.TenantsInfra)
				if in.Status.NodegroupStatus == nil {
					in.Status.NodegroupStatus = make(map[string]v1.NodegroupStatus)
				}
				in.Status.NodegroupStatus[ng] = v1.NodegroupStatus{Status: "DELETING"}
				return in
			})
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		_, found, err := ae.eksIC.DescribeNodegroup(ng)
		if err != nil {
			return ctrl.Result{}, err
		}
		if found {
			klog.Infof("waiting for nodegroup %s to be deleted", ng)
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}
	}

	// remove our finalizer from the list and update it.
	controllerutil.RemoveFinalizer(ae.tenantsInfra, tenantsFinalizer)
	klog.Infof("Deleted Tenant Infra [%s]", ae.tenantsInfra.GetName())
	if err := ae.client.Update(ae.ctx, ae.tenantsInfra.DeepCopyObject().(*v1.TenantsInfra)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
