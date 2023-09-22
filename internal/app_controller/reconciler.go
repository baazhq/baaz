package app_controller

import (
	"context"

	v1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/pkg/aws/eks"
)

func (r *ApplicationReconciler) do(ctx context.Context, app *v1.Applications, dp *v1.DataPlanes) error {

	if app == nil {
		return nil
	}

	eksIc := eks.NewEks(ctx, dp)
	eksClientSet, err := eksIc.GetEksClientSet()
	if err != nil {
		return err
	}

	applications := NewApplication(ctx, app, dp, r.Client, eksClientSet)

	if err := applications.ReconcileApplicationDeployer(); err != nil {
		return err
	}

	return nil
}
