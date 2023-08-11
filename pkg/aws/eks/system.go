package eks

import (
	"errors"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go/aws"
)

func (ec *eks) CreateSystemNodeGroup(nodeGroupInput awseks.CreateNodegroupInput) (*awseks.CreateNodegroupOutput, error) {

	createNodeGroupOutput, err := ec.awsClient.CreateNodegroup(ec.ctx, &nodeGroupInput)
	if err != nil {
		return nil, err
	}

	return createNodeGroupOutput, nil

}

func (ec *eks) DescribeNodegroup(nodeGroupName string) (output *awseks.DescribeNodegroupOutput, found bool, err error) {

	input := &awseks.DescribeNodegroupInput{
		ClusterName:   aws.String(ec.dp.Spec.CloudInfra.Eks.Name),
		NodegroupName: aws.String(nodeGroupName),
	}

	describeNodeGroupOutput, err := ec.awsClient.DescribeNodegroup(ec.ctx, input)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return describeNodeGroupOutput, true, nil
}

func (ec *eks) CreateNodegroup(createNodegroupInput *awseks.CreateNodegroupInput) (output *awseks.CreateNodegroupOutput, err error) {

	createNodeGroupOutput, err := ec.awsClient.CreateNodegroup(ec.ctx, createNodegroupInput)
	if err != nil {
		return nil, err
	}
	return createNodeGroupOutput, nil
}

func (ec *eks) UpdateNodegroup(updateNodeGroupConfig *awseks.UpdateNodegroupConfigInput) (output *awseks.UpdateNodegroupConfigOutput, err error) {

	updateNodeGroupOutput, err := ec.awsClient.UpdateNodegroupConfig(ec.ctx, updateNodeGroupConfig)
	if err != nil {
		return nil, err
	}
	return updateNodeGroupOutput, nil
}

func (ec *eks) DeleteNodeGroup(nodeGroupName string) (*awseks.DeleteNodegroupOutput, error) {

	result, err := ec.awsClient.DeleteNodegroup(ec.ctx, &awseks.DeleteNodegroupInput{
		ClusterName:   &ec.dp.Spec.CloudInfra.Eks.Name,
		NodegroupName: &nodeGroupName,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
