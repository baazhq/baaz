package eks

import (
	"context"

	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awssts "github.com/aws/aws-sdk-go-v2/service/sts"
	v1 "github.com/baazhq/baaz/api/v1/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	CreateSystemNodeGroup(nodeGroupInput awseks.CreateNodegroupInput) (*awseks.CreateNodegroupOutput, error)
	DeleteNodeGroup(nodeGroupName string) (*awseks.DeleteNodegroupOutput, error)
	DescribeNodegroup(nodeGroupName string) (output *awseks.DescribeNodegroupOutput, found bool, err error)
	CreateNodegroup(createNodegroupInput *awseks.CreateNodegroupInput) (output *awseks.CreateNodegroupOutput, err error)
	UpdateNodegroup(updateNodeGroupConfig *awseks.UpdateNodegroupConfigInput) (output *awseks.UpdateNodegroupConfigOutput, err error)
	// iam role
	CreateNodeIamRole(name string) (*awsiam.GetRoleOutput, error)
	CreateClusterIamRole() (*awsiam.GetRoleOutput, error)
	// addons
	CreateAddon(ctx context.Context, params *awseks.CreateAddonInput) (*awseks.CreateAddonOutput, error)
	DescribeAddon(addonName string) (*awseks.DescribeAddonOutput, error)
	// auth
	GetEksClientSet() (*kubernetes.Clientset, error)
	GetRestConfig() (*rest.Config, error)
	// roles
	CreateEbsCSIRole(ctx context.Context) (*awsiam.CreateRoleOutput, error)
	CreateVpcCniRole(ctx context.Context) (roleOutput *awsiam.CreateRoleOutput, arn string, err error)
}

type eks struct {
	ctx          context.Context
	awsClient    *awseks.Client
	awsIamClient *awsiam.Client
	awsStsClient *awssts.Client
	awsec2Client *awsec2.Client
	dp           *v1.DataPlanes
}

func NewEks(
	ctx context.Context,
	dp *v1.DataPlanes,
) Eks {
	return &eks{
		awsClient:    newAwsClient(ctx, dp.Spec.CloudInfra.Region),
		awsIamClient: newAwsIamClient(ctx, dp.Spec.CloudInfra.Region),
		awsStsClient: newAwsStsClient(ctx, dp.Spec.CloudInfra.Region),
		awsec2Client: newAwsEc2Client(ctx, dp.Spec.CloudInfra.Region),
		ctx:          ctx,
		dp:           dp,
	}
}
