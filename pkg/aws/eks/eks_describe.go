package eks

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
)

func DescribeCluster(ctx context.Context, env EksEnvironment) error {
	eksClient := awseks.NewFromConfig(env.Config)

	_, err := eksClient.DescribeCluster(ctx, &awseks.DescribeClusterInput{Name: aws.String(env.Env.Spec.CloudInfra.Eks.Name)})
	return err
}
