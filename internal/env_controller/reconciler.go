package controller

import (
	"context"
	"errors"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "datainfra.io/ballastdata/api/v1"
)

func (r *EnvironmentReconciler) do(ctx context.Context, env *v1.Environment) error {
	switch env.Spec.CloudInfra.Type {
	case v1.AWS:
		awsSecret, err := getSecret(ctx, r.Client, client.ObjectKey{
			Name:      env.Spec.CloudInfra.AuthSecretRef.SecretName,
			Namespace: env.Namespace,
		})
		if err != nil {
			return err
		}
		accessKey, found := awsSecret.Data[env.Spec.CloudInfra.AuthSecretRef.AccessKeyName]
		if !found {
			return errors.New("access key not found in the secret")
		}
		if err := os.Setenv("AWS_ACCESS_KEY_ID", string(accessKey)); err != nil {
			return err
		}

		secretKey, found := awsSecret.Data[env.Spec.CloudInfra.AuthSecretRef.SecretKeyName]
		if !found {
			return errors.New("secret key not found in the secret")
		}
		if err := os.Setenv("AWS_SECRET_ACCESS_KEY", string(secretKey)); err != nil {
			return err
		}
		return createOrUpdateAwsEksEnvironment(ctx, env, r.Client, r.Recorder)
	}
	return nil
}
