package eks

import (
	"context"

	v1 "datainfra.io/ballastdata/api/v1"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awssts "github.com/aws/aws-sdk-go-v2/service/sts"
)

type Eks interface {
	// eks control plane
	DescribeEks() (*awseks.DescribeClusterOutput, error)
	CreateEks() *EksInternalOutput
	UpdateEks() *EksInternalOutput
	DeleteEKS() (*awseks.DeleteClusterOutput, error)
	// oidc
	ListOIDCProvider() (*awsiam.ListOpenIDConnectProvidersOutput, error)
	CreateOIDCProvider(param *CreateOIDCProviderInput) (*awsiam.CreateOpenIDConnectProviderOutput, error)
	DeleteOIDCProvider(providerArn string) (*awsiam.DeleteOpenIDConnectProviderOutput, error)
	// nodegroups
	DescribeNodegroup(nodeGroupName string) (output *awseks.DescribeNodegroupOutput, found bool, err error)
	CreateSystemNodeGroup(nodeGroupInput awseks.CreateNodegroupInput) (*awseks.CreateNodegroupOutput, error)
	DeleteNodeGroup(nodeGroupName string) (*awseks.DeleteNodegroupOutput, error)
	// iam role
	CreateNodeIamRole(name string) (*awsiam.GetRoleOutput, error)
	CreateClusterIamRole() (*awsiam.GetRoleOutput, error)
	// addons
	CreateAddon(ctx context.Context, params *CreateAddonInput) (*awseks.CreateAddonOutput, error)
	DescribeAddon(addonName string) (*awseks.DescribeAddonOutput, error)
}

type eks struct {
	ctx          context.Context
	awsClient    *awseks.Client
	awsIamClient *awsiam.Client
	awsStsClient *awssts.Client
	environment  *v1.Environment
}

func NewEks(
	ctx context.Context,
	environment *v1.Environment,
) Eks {
	return &eks{
		awsClient:    newAwsClient(ctx, environment.Spec.CloudInfra.AwsRegion),
		awsIamClient: newAwsIamClient(ctx, environment.Spec.CloudInfra.AwsRegion),
		awsStsClient: newAwsStsClient(ctx, environment.Spec.CloudInfra.AwsRegion),
		ctx:          ctx,
		environment:  environment,
	}
}
