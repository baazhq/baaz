package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
)

const (
	EKSStatusCreating = "CREATING"
	EKSStatusACTIVE   = "ACTIVE"
	EKSStatusUpdating = "UPDATING"
)

type DescribeClusterOutput struct {
	Result *awseks.DescribeClusterOutput `json:"result"`
}

func DescribeCluster(ctx context.Context, env EksEnvironment) (*DescribeClusterOutput, error) {
	eksClient := awseks.NewFromConfig(env.Config)

	result, err := eksClient.DescribeCluster(ctx, &awseks.DescribeClusterInput{Name: aws.String(env.Env.Spec.CloudInfra.Eks.Name)})
	if err != nil {
		return nil, err
	}
	return &DescribeClusterOutput{Result: result}, nil
}
