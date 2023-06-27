package eks

import (
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	defaultClientID = "sts.amazonaws.com"
)

func (ec *eks) ReconcileOIDCProvider(clusterOutput *awseks.DescribeClusterOutput) (*awsiam.CreateOpenIDConnectProviderOutput, error) {
	if clusterOutput == nil || clusterOutput.Cluster == nil ||
		clusterOutput.Cluster.Identity == nil || clusterOutput.Cluster.Identity.Oidc == nil {
		return nil, errors.New("oidc provider url not found in cluster output")
	}
	oidcProviderUrl := *clusterOutput.Cluster.Identity.Oidc.Issuer

	// Compute the SHA-1 thumbprint of the OIDC provider certificate
	thumbprint, err := getIssuerCAThumbprint(oidcProviderUrl)
	if err != nil {
		return nil, err
	}

	input := &createOIDCProviderInput{
		URL:            oidcProviderUrl,
		ThumbPrintList: []string{thumbprint},
	}

	oidcProviderArn := ec.environment.Status.CloudInfraStatus.EksStatus.OIDCProviderArn

	if oidcProviderArn != "" {
		// oidc provider is previously created
		// looking for it
		providers, err := ec.listOIDCProvider()
		if err != nil {
			return nil, err
		}

		for _, oidc := range providers.OpenIDConnectProviderList {
			if *oidc.Arn == ec.environment.Status.CloudInfraStatus.EksStatus.OIDCProviderArn {
				// oidc provider is already created and existed
				return nil, nil
			}
		}
	}

	result, err := ec.createOIDCProvider(input)
	if err != nil {
		return nil, err
	}
	// _, _, err = utils.PatchStatus(ec.ctx, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
	// 	in := obj.(*v1.Environment)
	// 	in.Status.CloudInfraStatus.EksStatus.OIDCProviderArn = *result.OpenIDConnectProviderArn

	// 	return in
	// })
	return result, nil
}

type createOIDCProviderInput struct {
	URL            string   `json:"url"`
	ThumbPrintList []string `json:"thumbPrintList"`
}

func (ec *eks) listOIDCProvider() (*awsiam.ListOpenIDConnectProvidersOutput, error) {

	result, err := ec.awsIamClient.ListOpenIDConnectProviders(ec.ctx, &awsiam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ec *eks) createOIDCProvider(param *createOIDCProviderInput) (*awsiam.CreateOpenIDConnectProviderOutput, error) {

	result, err := ec.awsIamClient.CreateOpenIDConnectProvider(ec.ctx, &awsiam.CreateOpenIDConnectProviderInput{
		ThumbprintList: param.ThumbPrintList,
		Url:            aws.String(param.URL),
		ClientIDList:   []string{defaultClientID},
	})
	if err != nil {
		return nil, err
	}

	return result, nil
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
