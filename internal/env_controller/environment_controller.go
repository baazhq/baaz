/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"os"
	"time"

	"datainfra.io/ballastdata/pkg/aws/eks"

	"datainfra.io/ballastdata/pkg/store"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datainfraiov1 "datainfra.io/ballastdata/api/v1"
	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/utils"
)

const (
	BallasdataFinalizer = "environment.datainfra.io/finalizer"
)

// EnvironmentReconciler reconciles a Environment object
type EnvironmentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// reconcile time duration, defaults to 10s
	ReconcileWait time.Duration
	Recorder      record.EventRecorder
	NgStore       store.Store
}

func NewEnvironmentReconciler(mgr ctrl.Manager) *EnvironmentReconciler {
	initLogger := ctrl.Log.WithName("controllers").WithName("environment")
	return &EnvironmentReconciler{
		Client:        mgr.GetClient(),
		Log:           initLogger,
		Scheme:        mgr.GetScheme(),
		ReconcileWait: lookupReconcileTime(initLogger),
		Recorder:      mgr.GetEventRecorderFor("ballastdata-control-plane"),
		NgStore:       store.NewInternalStore(),
	}
}

// +kubebuilder:rbac:groups=datainfra.io,resources=environments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datainfra.io,resources=environments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=datainfra.io,resources=environments/finalizers,verbs=update
func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	desiredObj := &datainfraiov1.Environment{}
	err := r.Get(context.TODO(), req.NamespacedName, desiredObj)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	klog.Infof("Reconciling Environment: %s/%s", desiredObj.Namespace, desiredObj.Name)
	// check for deletion time stamp
	if desiredObj.DeletionTimestamp != nil {
		// object is going to be deleted
		return r.reconcileDelete(ctx, desiredObj)
	}

	// if it is normal reconcile, then add finalizer if not already
	if !controllerutil.ContainsFinalizer(desiredObj, BallasdataFinalizer) {
		controllerutil.AddFinalizer(desiredObj, BallasdataFinalizer)
		if err := r.Update(ctx, desiredObj); err != nil {
			return ctrl.Result{}, err
		}
	}

	// If first time reconciling set status to pending
	if desiredObj.Status.Phase == "" {
		if _, _, err := utils.PatchStatus(ctx, r.Client, desiredObj, func(obj client.Object) client.Object {
			in := obj.(*datainfraiov1.Environment)
			in.Status.Phase = datainfraiov1.Pending
			return in
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.do(ctx, desiredObj); err != nil {
		if _, _, upErr := utils.PatchStatus(ctx, r.Client, desiredObj, func(obj client.Object) client.Object {
			in := obj.(*datainfraiov1.Environment)
			in.Status.Phase = datainfraiov1.Failed
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

func (r *EnvironmentReconciler) reconcileDelete(ctx context.Context, env *datainfraiov1.Environment) (ctrl.Result, error) {
	ngList := r.NgStore.List(env.Spec.CloudInfra.Eks.Name)
	eksEnv := eks.NewEksEnvironment(ctx, r.Client, env, *eks.NewConfig(env.Spec.CloudInfra.AwsRegion))

	for _, ng := range ngList {
		if env.Status.NodegroupStatus[ng] != "DELETING" {
			_, err := eksEnv.DeleteNodeGroup(ng)
			if err != nil {
				return ctrl.Result{}, err
			}
			// update status with current nodegroup status
			_, _, err = utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
				in := obj.(*v1.Environment)
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
	}

	for _, ng := range ngList {
		_, found, err := eksEnv.DescribeNodeGroup(ng)
		if err != nil {
			return ctrl.Result{}, err
		}
		if found {
			klog.Infof("waiting for nodegroup %s to be deleted", ng)
			return ctrl.Result{RequeueAfter: time.Second * 10}, nil
		}
	}

	// delete oidc provider associated with the cluster(if any)
	if env.Status.CloudInfraStatus.EksStatus.OIDCProviderArn != "" {
		_, err := eksEnv.DeleteOIDCProvider(env.Status.CloudInfraStatus.EksStatus.OIDCProviderArn)
		if err != nil {
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
	}

	if _, err := eksEnv.DeleteEKS(); err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}

	// remove our finalizer from the list and update it.
	controllerutil.RemoveFinalizer(env, envFinalizer)
	if err := r.Update(ctx, env.DeepCopyObject().(*v1.Environment)); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datainfraiov1.Environment{}).
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
