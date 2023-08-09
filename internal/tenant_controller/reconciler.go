package tenant_controller

import (
	"context"

	v1 "datainfra.io/ballastdata/api/v1/types"
	"datainfra.io/ballastdata/pkg/aws/eks"
)

func (r *TenantsReconciler) do(ctx context.Context, tenant *v1.Tenants, env *v1.Environment) error {

	eksClient := eks.NewEks(ctx, env)

	awsEnv := awsEnv{
		ctx:    ctx,
		env:    env,
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
