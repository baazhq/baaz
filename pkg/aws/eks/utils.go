package eks

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awssts "github.com/aws/aws-sdk-go-v2/service/sts"
)

func MakeEksClusterRoleName(clusterName string) string { return clusterName + "-" + "cluster-role" }
func MakeEksNodeRoleName(nodeGroupName string) string  { return nodeGroupName + "-" + "node-role" }
func MakeEBSCSIRoleName(region, clusterName string) string {
	return region + "-" + clusterName + "-" + "ebs-role"
}

func MakeVpcCniRoleName(region, clusterName string) string {
	return region + "-" + clusterName + "-" + "vpccni-role"
}

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

func newAwsStsClient(ctx context.Context, region string) *awssts.Client {
	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to build AWS client, %v", err)
	}

	return awssts.NewFromConfig(config)
}

func newAwsEc2Client(ctx context.Context, region string) *awsec2.Client {
	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("unable to build AWS client, %v", err)
	}

	return awsec2.NewFromConfig(config)
}
