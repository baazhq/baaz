package eks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"

	"github.com/aws/aws-sdk-go-v2/aws"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
)

type CreateNodeGroupOutput struct {
	Result *awseks.CreateNodegroupOutput `json:"result"`
}

func CreateNodeGroup(ctx context.Context, eksEnv EksEnvironment) (*CreateNodeGroupOutput, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)
	input := &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(eksEnv.Env.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String("arn:aws:iam::437639712640:role/pulak-eks-node-role"),
		NodegroupName:      aws.String(fmt.Sprintf("%s-%s", eksEnv.Env.Spec.CloudInfra.Eks.Name, eksEnv.Env.Spec.Application[0].Name)),
		Subnets:            eksEnv.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds,
		AmiType:            "",
		CapacityType:       "",
		ClientRequestToken: nil,
		DiskSize:           nil,
		InstanceTypes:      []string{"t3.medium"},
		Labels:             nil,
		LaunchTemplate:     nil,
		ReleaseVersion:     nil,
		RemoteAccess:       nil,
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(1),
			MaxSize:     aws.Int32(1),
			MinSize:     aws.Int32(1),
		},
		Tags:         nil,
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
