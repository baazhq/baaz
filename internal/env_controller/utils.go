package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getSecret(ctx context.Context, c client.Client, key client.ObjectKey) (*corev1.Secret, error) {
	secret := &corev1.Secret{}

	if err := c.Get(ctx, key, secret); err != nil {
		return nil, err
	}

	return secret, nil
}
