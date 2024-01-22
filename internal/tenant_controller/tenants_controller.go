package tenant_controller

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
	"datainfra.io/baaz/pkg/store"
	"datainfra.io/baaz/pkg/utils"
)

var tenantsFinalizer = "tenants.datainfra.io/finalizer"

// TenantsReconciler reconciles a Tenants object
type TenantsReconciler struct {
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

func NewTenantsReconciler(mgr ctrl.Manager, enablePrivate bool, customerName string) *TenantsReconciler {
	initLogger := ctrl.Log.WithName("controllers").WithName("tenant")
	return &TenantsReconciler{
		Client:        mgr.GetClient(),
		Log:           initLogger,
		Scheme:        mgr.GetScheme(),
		ReconcileWait: lookupReconcileTime(initLogger),
		Recorder:      mgr.GetEventRecorderFor("tenant-controller"),
		Predicates:    predicates.GetPredicates(enablePrivate, customerName),
		NgStore:       store.NewInternalStore(),
	}
}

//+kubebuilder:rbac:groups=datainfra.io,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=datainfra.io,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=datainfra.io,resources=tenants/finalizers,verbs=update

func (r *TenantsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	tenantObj := &v1.Tenants{}
	err := r.Get(ctx, req.NamespacedName, tenantObj)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	dpObj := &v1.DataPlanesList{}
	err = r.List(ctx, dpObj, &client.ListOptions{})
	if err != nil {
		return ctrl.Result{}, err
	}

	var dataplane v1.DataPlanes
	for _, dp := range dpObj.Items {
		if dp.GetName() == tenantObj.Spec.DataplaneName {
			dataplane = dp
		}
	}

	klog.Infof("Reconciling Tenants: %s/%s", tenantObj.Namespace, tenantObj.Name)

	if tenantObj.DeletionTimestamp != nil {
		// object is going to be deleted
		awsEnv := awsEnv{
			ctx:    ctx,
			dp:     &dataplane,
			tenant: tenantObj,
			eksIC:  eks.NewEks(ctx, &dataplane),
			client: r.Client,
			store:  r.NgStore,
		}

		return r.reconcileDelete(&awsEnv)
	}
	// if it is normal reconcile, then add finalizer if not already
	if !controllerutil.ContainsFinalizer(tenantObj, tenantsFinalizer) {
		controllerutil.AddFinalizer(tenantObj, tenantsFinalizer)
		if err := r.Update(ctx, tenantObj); err != nil {
			return ctrl.Result{}, err
		}
	}
	// If first time reconciling set status to pending
	if tenantObj.Status.Phase == "" {
		if _, _, err := utils.PatchStatus(ctx, r.Client, tenantObj, func(obj client.Object) client.Object {
			in := obj.(*v1.Tenants)
			in.Status.Phase = v1.PendingT
			return in
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.do(ctx, tenantObj, &dataplane); err != nil {
		if _, _, patchErr := utils.PatchStatus(ctx, r.Client, tenantObj, func(obj client.Object) client.Object {
			in := obj.(*v1.Tenants)
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
func (r *TenantsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Tenants{}).
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

func (r *TenantsReconciler) reconcileDelete(ae *awsEnv) (ctrl.Result, error) {
	// update phase to terminating
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.tenant, func(obj client.Object) client.Object {
		in := obj.(*v1.Tenants)
		in.Status.Phase = v1.TerminatingT
		return in
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	for ng, ngStatus := range ae.tenant.Status.NodegroupStatus {

		if ngStatus != "DELETING" {
			_, err := ae.eksIC.DeleteNodeGroup(ng)
			if err != nil {
				return ctrl.Result{}, err
			}
			// update status with current nodegroup status
			_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.tenant, func(obj client.Object) client.Object {
				in := obj.(*v1.Tenants)
				if in.Status.NodegroupStatus == nil {
					in.Status.NodegroupStatus = make(map[string]string)
				}
				in.Status.NodegroupStatus[ng] = "DELETING"
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
	controllerutil.RemoveFinalizer(ae.tenant, tenantsFinalizer)
	klog.Infof("Deleted Tenant [%s]", ae.tenant.GetName())
	if err := ae.client.Update(ae.ctx, ae.tenant.DeepCopyObject().(*v1.Tenants)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
