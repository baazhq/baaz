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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	datainfraiov1 "datainfra.io/ballastdata/api/v1"
)

// EnvironmentReconciler reconciles a Environment object
type EnvironmentReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=datainfra.io,resources=environments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datainfra.io,resources=environments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=datainfra.io,resources=environments/finalizers,verbs=update
func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	_ = context.Background()

	desiredObj := &datainfraiov1.Environment{}
	err := r.Get(context.TODO(), req.NamespacedName, desiredObj)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// If first time reconciling set status to pending
	if desiredObj.Status.Phase == "" {
		if _, _, err := PatchStatus(ctx, r.Client, desiredObj, func(obj client.Object) client.Object {
			in := obj.(*datainfraiov1.Environment)
			in.Status.Phase = datainfraiov1.Pending
			return in
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := reconcileEnvironment(ctx, r.Client, desiredObj, r.Recorder); err != nil {
		if _, _, upErr := PatchStatus(ctx, r.Client, desiredObj, func(obj client.Object) client.Object {
			in := obj.(*datainfraiov1.Environment)
			in.Status.Phase = datainfraiov1.Failed
			return in
		}); upErr != nil {
			return ctrl.Result{}, upErr
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	} else {
		return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datainfraiov1.Environment{}).
		Complete(r)
}
