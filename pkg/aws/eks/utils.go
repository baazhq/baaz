package eks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

func makeEksClusterRoleName(clusterName string) string { return clusterName + "-" + "cluster-role" }

func newAwsClient(ctx context.Context, region string) *awseks.Client {
	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to build AWS client, %v", err)
	}

	return awseks.NewFromConfig(config)
}

func newAwsIamClient(ctx context.Context, region string) *awsiam.Client {
	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to build AWS client, %v", err)
	}

	return awsiam.NewFromConfig(config)
}
