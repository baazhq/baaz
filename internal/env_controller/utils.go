package controller

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

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

func getIssuerCAThumbprint(isserURL string) (string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}

	response, err := client.Get(isserURL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.TLS != nil {
		if numCerts := len(response.TLS.PeerCertificates); numCerts >= 1 {
			root := response.TLS.PeerCertificates[numCerts-1]
			return fmt.Sprintf("%x", sha1.Sum(root.Raw)), nil
		}
	}
	return "", errors.New("unable to get OIDC issuer's certificate")
}
