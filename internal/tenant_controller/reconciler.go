package tenant_controller

import (
	"context"

	v1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/pkg/aws/eks"
)

func (r *TenantsReconciler) do(ctx context.Context, tenant *v1.Tenants, dp *v1.DataPlanes) error {

	eksClient := eks.NewEks(ctx, dp)

	awsEnv := awsEnv{
		ctx:    ctx,
		dp:     dp,
		eksIC:  eksClient,
		tenant: tenant,
		client: r.Client,
		store:  r.NgStore,
	}

	if err := awsEnv.ReconcileTenants(); err != nil {
		return err
	}

	return nil
}
