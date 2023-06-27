package eks

import (
	"context"

	v1 "datainfra.io/ballastdata/api/v1"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

type Eks interface {
	DescribeEks() (*awseks.DescribeClusterOutput, error)
	CreateEks() *EksInternalOutput
	UpdateEks() *EksInternalOutput
	ReconcileOIDCProvider(clusterOutput *awseks.DescribeClusterOutput) (*awsiam.CreateOpenIDConnectProviderOutput, error)
}

type eks struct {
	ctx          context.Context
	awsClient    *awseks.Client
	awsIamClient *awsiam.Client
	environment  *v1.Environment
}

func NewEks(
	ctx context.Context,
	environment *v1.Environment,
) Eks {
	return &eks{
		awsClient:    newAwsClient(ctx, environment.Spec.CloudInfra.AwsRegion),
		awsIamClient: newAwsIamClient(ctx, environment.Spec.CloudInfra.AwsRegion),
		ctx:          ctx,
		environment:  environment,
	}
}
