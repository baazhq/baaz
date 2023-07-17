package app_controller

import (
	"context"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"k8s.io/klog/v2"
)

func (r *ApplicationReconciler) do(ctx context.Context, app *v1.Application, env *v1.Environment) error {

	klog.Info("Reconciling Application")
	eksIc := eks.NewEks(ctx, env)
	clientset, err := eksIc.GetEksClientSet()
	if err != nil {
		return err
	}
	applications := NewApplication(ctx, app, env, clientset)

	if err := applications.ReconcileApplicationDeployer(); err != nil {
		return err
	}

	return nil
}
