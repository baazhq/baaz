package eks

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"k8s.io/klog/v2"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	defaultClientID = "sts.amazonaws.com"
)

type CreateOIDCProviderInput struct {
	URL            string   `json:"url"`
	ThumbPrintList []string `json:"thumbPrintList"`
}

func (ec *eks) ListOIDCProvider() (*awsiam.ListOpenIDConnectProvidersOutput, error) {

	result, err := ec.awsIamClient.ListOpenIDConnectProviders(ec.ctx, &awsiam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ec *eks) CreateOIDCProvider(param *CreateOIDCProviderInput) (*awsiam.CreateOpenIDConnectProviderOutput, error) {

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

func (ec *eks) DeleteOIDCProvider(providerArn string) (*awsiam.DeleteOpenIDConnectProviderOutput, error) {
	klog.Infof("Deleting Oidc Provider [%s]", providerArn)

	output, err := ec.awsIamClient.DeleteOpenIDConnectProvider(ec.ctx, &iam.DeleteOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: &providerArn,
	})
	if err != nil {
		klog.Infof("Response Deleting Oidc Provider [%s]", err.Error())
		return output, nil
	}

	return output, nil
}
