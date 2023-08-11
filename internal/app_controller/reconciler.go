package app_controller

import (
	"context"
	"fmt"

	v1 "datainfra.io/ballastdata/api/v1/types"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"k8s.io/klog/v2"
)

func (r *ApplicationReconciler) do(ctx context.Context, app *v1.Application, dp *v1.DataPlanes) error {

	if app == nil {
		return nil
	}

	fmt.Println("app", app)
	fmt.Println("env", dp)

	klog.Info("Reconciling Application")
	eksIc := eks.NewEks(ctx, dp)
	clientset, err := eksIc.GetEksClientSet()
	if err != nil {
		return err
	}
	applications := NewApplication(ctx, app, dp, clientset)

	if err := applications.ReconcileApplicationDeployer(); err != nil {
		return err
	}

	return nil
}
