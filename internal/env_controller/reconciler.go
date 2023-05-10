package controller

import (
	"context"
	"os"

	v1 "datainfra.io/ballastdata/api/v1"
)

func (r *EnvironmentReconciler) do(ctx context.Context, env *v1.Environment) error {
	switch env.Spec.CloudInfra.Type {
	case v1.AWS:
		os.Setenv("AWS_ACCESS_KEY_ID", env.Spec.CloudInfra.Auth.AwsAccessKey)
		os.Setenv("AWS_SECRET_ACCESS_KEY", env.Spec.CloudInfra.Auth.AwsSecretAccessKey)
		return createOrUpdateAwsEksEnvironment(ctx, env, r.Client, r.Recorder)
	}
	return nil
}
