package app_controller

import (
	"context"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/application"
)

func (r *ApplicationReconciler) do(ctx context.Context, app *v1.Application) error {

	apps := application.NewApplication(ctx, r.Client, app)
	err := apps.ReconcileApplicationDeployer()

	if err != nil {
		return err
	}
	return nil
}
