package controller

import (
	"context"
	"errors"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "datainfra.io/ballastdata/api/v1/types"
)

func (r *DataPlaneReconciler) do(ctx context.Context, dp *v1.DataPlanes) error {
	switch dp.Spec.CloudInfra.CloudType {

	case v1.CloudType(v1.AWS):

		awsSecret, err := getSecret(ctx, r.Client, client.ObjectKey{
			Name:      dp.Spec.CloudInfra.AuthSecretRef.SecretName,
			Namespace: dp.Namespace,
		})
		if err != nil {
			return err
		}

		accessKey, found := awsSecret.Data[dp.Spec.CloudInfra.AuthSecretRef.AccessKeyName]
		if !found {
			return errors.New("access key not found in the secret")
		}

		if err := os.Setenv("AWS_ACCESS_KEY_ID", string(accessKey)); err != nil {
			return err
		}

		secretKey, found := awsSecret.Data[dp.Spec.CloudInfra.AuthSecretRef.SecretKeyName]
		if !found {
			return errors.New("secret key not found in the secret")
		}

		if err := os.Setenv("AWS_SECRET_ACCESS_KEY", string(secretKey)); err != nil {
			return err
		}

		if err := r.reconcileAwsEnvironment(ctx, dp); err != nil {
			return err
		}

	}
	return nil
}
