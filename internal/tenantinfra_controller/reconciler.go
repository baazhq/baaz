package tenantinfra_controller

import (
	"context"

	v1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/pkg/aws/eks"
)

func (r *TenantsInfraReconciler) do(ctx context.Context, tenantsInfra *v1.TenantsInfra, dp *v1.DataPlanes) error {

	eksClient := eks.NewEks(ctx, dp)

	awsEnv := awsEnv{
		ctx:          ctx,
		dp:           dp,
		eksIC:        eksClient,
		tenantsInfra: tenantsInfra,
		client:       r.Client,
		store:        r.NgStore,
	}

	if err := awsEnv.ReconcileInfraTenants(); err != nil {
		return err
	}

	return nil
}
