package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"datainfra.io/ballastdata/pkg/aws/eks"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	defaultClientID = "sts.amazonaws.com"
)

type CreateOIDCProviderInput struct {
	URL            string   `json:"url"`
	ThumbPrintList []string `json:"thumbPrintList"`
}

type CreateOIDCProviderOutput struct {
	Result *awsiam.CreateOpenIDConnectProviderOutput `json:"result"`
}

type ListOIDCProviderOutput struct {
	Result *awsiam.ListOpenIDConnectProvidersOutput `json:"result"`
}

func ListOIDCProvider(ctx context.Context, eksEnv *eks.EksEnvironment) (*ListOIDCProviderOutput, error) {
	iamClient := awsiam.NewFromConfig(eksEnv.Config)

	result, err := iamClient.ListOpenIDConnectProviders(ctx, &awsiam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return nil, err
	}

	return &ListOIDCProviderOutput{Result: result}, nil
}

func CreateOIDCProvider(ctx context.Context, eksEnv *eks.EksEnvironment, param *CreateOIDCProviderInput) (*CreateOIDCProviderOutput, error) {
	iamClient := awsiam.NewFromConfig(eksEnv.Config)

	result, err := iamClient.CreateOpenIDConnectProvider(ctx, &awsiam.CreateOpenIDConnectProviderInput{
		ThumbprintList: param.ThumbPrintList,
		Url:            aws.String(param.URL),
		ClientIDList:   []string{defaultClientID},
	})
	if err != nil {
		return nil, err
	}

	return &CreateOIDCProviderOutput{
		Result: result,
	}, nil
}
