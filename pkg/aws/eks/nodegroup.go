package eks

import (
	"context"
	"fmt"

	v1 "datainfra.io/ballastdata/api/v1"
	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

type NodegroupStatus string

type CreateNodeGroupOutput struct {
	Result *awseks.CreateNodegroupOutput `json:"result"`
}

type DescribeNodegroupOutput struct {
	Result *awseks.DescribeNodegroupOutput `json:"result"`
}

func DescribeNodegroup(ctx context.Context, eksEnv *EksEnvironment, app *v1.ApplicationConfig) (*DescribeNodegroupOutput, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)
	input := &awseks.DescribeNodegroupInput{
		ClusterName:   aws.String(eksEnv.Env.Spec.CloudInfra.Eks.Name),
		NodegroupName: aws.String(fmt.Sprintf("%s-%s", eksEnv.Env.Spec.CloudInfra.Eks.Name, app.Name)),
	}

	result, err := eksClient.DescribeNodegroup(ctx, input)
	if err != nil {
		return nil, err
	}
	return &DescribeNodegroupOutput{Result: result}, nil
}

func CreateNodeGroup(ctx context.Context, eksEnv *EksEnvironment, nodeSpec *v1.NodeGroupSpec, app *v1.ApplicationConfig) (*CreateNodeGroupOutput, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)
	input := &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(eksEnv.Env.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String("arn:aws:iam::437639712640:role/pulak-eks-node-role"),
		NodegroupName:      aws.String(fmt.Sprintf("%s-%s", eksEnv.Env.Spec.CloudInfra.Eks.Name, app.Name)),
		Subnets:            eksEnv.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds,
		AmiType:            "",
		CapacityType:       "",
		ClientRequestToken: nil,
		DiskSize:           nil,
		InstanceTypes:      []string{nodeSpec.NodeSize},
		Labels:             nil,
		LaunchTemplate:     nil,
		ReleaseVersion:     nil,
		RemoteAccess:       nil,
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(nodeSpec.Min),
			MaxSize:     aws.Int32(nodeSpec.Max),
			MinSize:     aws.Int32(nodeSpec.Min),
		},
		Tags: map[string]string{
			fmt.Sprintf("kubernetes.io/cluster/%s", eksEnv.Env.Spec.CloudInfra.Eks.Name): "owned",
		},
		Taints:       nil,
		UpdateConfig: nil,
		Version:      nil,
	}

	result, err := eksClient.CreateNodegroup(ctx, input)
	if err != nil {
		return nil, err
	}
	return &CreateNodeGroupOutput{Result: result}, nil
}
