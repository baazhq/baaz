package controller

import (
	"context"

	v1 "github.com/baazhq/baaz/api/v1/types"
)

const (
	aws_access_key string = "AWS_ACCESS_KEY_ID"
	aws_secret_key string = "AWS_SECRET_ACCESS_KEY"
)

func (r *DataPlaneReconciler) do(ctx context.Context, dp *v1.DataPlanes) error {
	switch dp.Spec.CloudInfra.CloudType {

	case v1.CloudType(v1.AWS):
		if err := r.reconcileAwsEnvironment(ctx, dp); err != nil {
			return err
		}

	}

	return nil
}
