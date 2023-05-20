package eks

import (
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/utils"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	defaultClientID = "sts.amazonaws.com"
)

func (eksEnv *EksEnvironment) ReconcileOIDCProvider(clusterOutput *awseks.DescribeClusterOutput) error {
	if clusterOutput == nil || clusterOutput.Cluster == nil ||
		clusterOutput.Cluster.Identity == nil || clusterOutput.Cluster.Identity.Oidc == nil {
		return errors.New("oidc provider url not found in cluster output")
	}
	oidcProviderUrl := *clusterOutput.Cluster.Identity.Oidc.Issuer

	// Compute the SHA-1 thumbprint of the OIDC provider certificate
	thumbprint, err := getIssuerCAThumbprint(oidcProviderUrl)
	if err != nil {
		return err
	}

	input := &createOIDCProviderInput{
		URL:            oidcProviderUrl,
		ThumbPrintList: []string{thumbprint},
	}

	oidcProviderArn := eksEnv.Env.Status.CloudInfraStatus.EksStatus.OIDCProviderArn

	if oidcProviderArn != "" {
		// oidc provider is previously created
		// looking for it
		providers, err := eksEnv.listOIDCProvider()
		if err != nil {
			return err
		}

		for _, oidc := range providers.OpenIDConnectProviderList {
			if *oidc.Arn == eksEnv.Env.Status.CloudInfraStatus.EksStatus.OIDCProviderArn {
				// oidc provider is already created and existed
				return nil
			}
		}
	}

	result, err := eksEnv.createOIDCProvider(input)
	if err != nil {
		return err
	}
	_, _, err = utils.PatchStatus(eksEnv.Context, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		in.Status.CloudInfraStatus.EksStatus.OIDCProviderArn = *result.OpenIDConnectProviderArn

		return in
	})
	return err
}

type createOIDCProviderInput struct {
	URL            string   `json:"url"`
	ThumbPrintList []string `json:"thumbPrintList"`
}

func (eksEnv *EksEnvironment) listOIDCProvider() (*awsiam.ListOpenIDConnectProvidersOutput, error) {
	iamClient := awsiam.NewFromConfig(eksEnv.Config)

	result, err := iamClient.ListOpenIDConnectProviders(eksEnv.Context, &awsiam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (eksEnv *EksEnvironment) createOIDCProvider(param *createOIDCProviderInput) (*awsiam.CreateOpenIDConnectProviderOutput, error) {
	iamClient := awsiam.NewFromConfig(eksEnv.Config)

	result, err := iamClient.CreateOpenIDConnectProvider(eksEnv.Context, &awsiam.CreateOpenIDConnectProviderInput{
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
