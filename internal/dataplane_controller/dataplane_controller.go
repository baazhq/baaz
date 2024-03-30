package controller

import (
	"context"
	"errors"
	"os"
	"time"

	v1 "github.com/baazhq/baaz/api/v1/types"
	"github.com/baazhq/baaz/internal/predicates"
	"github.com/baazhq/baaz/pkg/aws/eks"
	"github.com/baazhq/baaz/pkg/aws/network"
	"github.com/baazhq/baaz/pkg/store"
	"github.com/baazhq/baaz/pkg/utils"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	dataplaneFinalizer = "dataplane.datainfra.io/finalizer"
)

// DataPlaneReconciler reconciles a Environment object
type DataPlaneReconciler struct {
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

func NewDataplaneReconciler(mgr ctrl.Manager, enablePrivate bool, customerName string) *DataPlaneReconciler {
	initLogger := ctrl.Log.WithName("controllers").WithName("dataplane")

	return &DataPlaneReconciler{
		Client:        mgr.GetClient(),
		Log:           initLogger,
		Scheme:        mgr.GetScheme(),
		ReconcileWait: lookupReconcileTime(initLogger),
		Recorder:      mgr.GetEventRecorderFor("dataplane-controller"),
		Predicates:    predicates.GetPredicates(enablePrivate, customerName),
		NgStore:       store.NewInternalStore(),
	}

}

func (r *DataPlaneReconciler) initCloudAuth(ctx context.Context, dp *v1.DataPlanes) error {
	awsSecret, err := getSecret(ctx, r.Client, client.ObjectKey{
		Name:      dp.Spec.CloudInfra.AuthSecretRef.SecretName,
		Namespace: dp.Namespace,
	})
	if err != nil {
		return err
	}

	accessKey, found := awsSecret.Data[dp.Spec.CloudInfra.AuthSecretRef.AccessKeyName]
	if !found {
		return errors.New("access key not found in the secret")
	}

	if err := os.Setenv(aws_access_key, string(accessKey)); err != nil {
		return err
	}

	secretKey, found := awsSecret.Data[dp.Spec.CloudInfra.AuthSecretRef.SecretKeyName]
	if !found {
		return errors.New("secret key not found in the secret")
	}

	if err := os.Setenv(aws_secret_key, string(secretKey)); err != nil {
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups=datainfra.io,resources=dataplanes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datainfra.io,resources=dataplanes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=datainfra.io,resources=dataplanes/finalizers,verbs=update
func (r *DataPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	desiredObj := &v1.DataPlanes{}
	err := r.Get(ctx, req.NamespacedName, desiredObj)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.initCloudAuth(ctx, desiredObj); err != nil {
		return ctrl.Result{}, err
	}

	klog.Infof("Reconciling Dataplane: %s/%s", desiredObj.Namespace, desiredObj.Name)
	networkMgr, err := network.NewProvisioner(ctx, desiredObj.Spec.CloudInfra.Region)
	if err != nil {
		return ctrl.Result{}, err
	}
	// check for deletion time stamp
	if desiredObj.DeletionTimestamp != nil {
		// object is going to be deleted
		awsEnv := awsEnv{
			ctx:     ctx,
			dp:      desiredObj,
			eksIC:   eks.NewEks(ctx, desiredObj),
			client:  r.Client,
			store:   r.NgStore,
			network: networkMgr,
		}

		return r.reconcileDelete(&awsEnv)
	}

	// if it is normal reconcile, then add finalizer if not already
	if !controllerutil.ContainsFinalizer(desiredObj, dataplaneFinalizer) {
		controllerutil.AddFinalizer(desiredObj, dataplaneFinalizer)
		if err := r.Update(ctx, desiredObj); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.do(ctx, desiredObj); err != nil {
		if _, _, upErr := utils.PatchStatus(ctx, r.Client, desiredObj, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			in.Status.Phase = v1.FailedD
			return in
		}); upErr != nil {
			return ctrl.Result{}, upErr
		}
		klog.Errorf("failed to reconcile environment: reason: %s", err.Error())
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	} else {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}
}

func (r *DataPlaneReconciler) reconcileDelete(ae *awsEnv) (ctrl.Result, error) {
	// update phase to terminating
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
		in := obj.(*v1.DataPlanes)
		in.Status.Phase = v1.TerminatingD
		return in
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	systemNodeGroupName := ae.dp.Spec.CloudInfra.Eks.Name + "-system"

	_, found, _ := ae.eksIC.DescribeNodegroup(systemNodeGroupName)
	if found {

		if ae.dp.Status.NodegroupStatus[systemNodeGroupName] != "DELETING" {
			_, _ = ae.eksIC.DeleteNodeGroup(systemNodeGroupName)
			// update status with current nodegroup status
			_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				if in.Status.NodegroupStatus == nil {
					in.Status.NodegroupStatus = make(map[string]string)
				}
				in.Status.NodegroupStatus[systemNodeGroupName] = "DELETING"
				return in
			})
			if err != nil {
				return ctrl.Result{}, err
			}

		}
		klog.Infof("waiting for nodegroup %s to be deleted", systemNodeGroupName)
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	// delete oidc provider associated with the cluster(if any)
	if ae.dp.Status.CloudInfraStatus.EksStatus.OIDCProviderArn != "" {
		_, err := ae.eksIC.DeleteOIDCProvider(ae.dp.Status.CloudInfraStatus.EksStatus.OIDCProviderArn)
		if err != nil {
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
	}

	if _, err := ae.eksIC.DeleteEKS(); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}

	if ae.dp.Spec.CloudInfra.ProvisionNetwork {
		if err := deleteNetworkComponent(ae); err != nil {
			return ctrl.Result{}, err
		}
	}

	// remove our finalizer from the list and update it.
	controllerutil.RemoveFinalizer(ae.dp, dataplaneFinalizer)
	klog.Infof("Deleted Dataplane [%s]", ae.dp.GetName())
	if err := ae.client.Update(ae.ctx, ae.dp.DeepCopyObject().(*v1.DataPlanes)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func deleteNetworkComponent(ae *awsEnv) error {
	if err := ae.network.DeleteVpcLBs(ae.ctx, ae.dp.Status.CloudInfraStatus.Vpc); err != nil {
		return err
	}
	if ae.dp.Status.CloudInfraStatus.InternetGatewayId != "" {
		if err := ae.network.DetachInternetGateway(ae.ctx,
			ae.dp.Status.CloudInfraStatus.InternetGatewayId, ae.dp.Status.CloudInfraStatus.Vpc); err != nil {
			return err
		}

		if err := ae.network.DeleteInternetGateway(ae.ctx, ae.dp.Status.CloudInfraStatus.InternetGatewayId); err != nil {
			return err
		}
	}
	if ae.dp.Status.CloudInfraStatus.NATGatewayId != "" {
		if err := ae.network.DeleteNatGateway(ae.ctx, ae.dp.Status.CloudInfraStatus.NATGatewayId); err != nil {
			klog.Error(err)
		}
	}
	return ae.network.DeleteVPC(ae.ctx, ae.dp.Status.CloudInfraStatus.Vpc)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataPlaneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.DataPlanes{}).
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
