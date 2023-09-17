package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
)

func (ec *eks) DescribeAddon(addonName string) (*awseks.DescribeAddonOutput, error) {

	input := &awseks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(ec.dp.Spec.CloudInfra.Eks.Name),
	}
	result, err := ec.awsClient.DescribeAddon(ec.ctx, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type CreateAddonInput struct {
	Name        string `json:"name"`
	ClusterName string `json:"clusterName"`
}

func (ec *eks) CreateAddon(ctx context.Context, params *awseks.CreateAddonInput) (*awseks.CreateAddonOutput, error) {

	result, err := ec.awsClient.CreateAddon(ctx, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}
