package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

func (ec *eks) DescribeAddon(addonName string) (*awseks.DescribeAddonOutput, error) {

	input := &awseks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(ec.environment.Spec.CloudInfra.Eks.Name),
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

func (ec *eks) CreateAddon(ctx context.Context, params *CreateAddonInput) (*awseks.CreateAddonOutput, error) {

	role, err := ec.createEbsCSIRole(ctx)
	if err != nil {
		return nil, err
	}

	input := &awseks.CreateAddonInput{
		AddonName:             aws.String(params.Name),
		ClusterName:           aws.String(params.ClusterName),
		ResolveConflicts:      types.ResolveConflictsOverwrite,
		ServiceAccountRoleArn: role.Role.Arn,
	}

	result, err := ec.awsClient.CreateAddon(ctx, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}
