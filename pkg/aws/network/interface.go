package network

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	v1 "github.com/baazhq/baaz/api/v1/types"
)

type Network interface {
	CreateVPC(ctx context.Context, params *awsec2.CreateVpcInput) (*awsec2.CreateVpcOutput, error)
	CreateSubnet(ctx context.Context, params *awsec2.CreateSubnetInput) (*awsec2.CreateSubnetOutput, error)
	CreateSG(ctx context.Context, params *awsec2.CreateSecurityGroupInput) (*awsec2.CreateSecurityGroupOutput, error)
	CreateNAT(ctx context.Context, dp *v1.DataPlanes) (*awsec2.CreateNatGatewayOutput, error)
	CreateElasticIP(ctx context.Context) (*awsec2.AllocateAddressOutput, error)
	AssociateNATWithRT(ctx context.Context, dp *v1.DataPlanes) error
	CreateInternetGateway(ctx context.Context) (*awsec2.CreateInternetGatewayOutput, error)
	AttachInternetGateway(ctx context.Context, igId, vpcId string) (*awsec2.AttachInternetGatewayOutput, error)
	AddSGInboundRule(ctx context.Context, sgGroupId, vpcId string) (*awsec2.AuthorizeSecurityGroupIngressOutput, error)
	SubnetAutoAssignPublicIP(ctx context.Context, subnetId string) (*awsec2.ModifySubnetAttributeOutput, error)
	CreateRouteTable(ctx context.Context, vpcId string) (*awsec2.CreateRouteTableOutput, error)
	CreateRoute(ctx context.Context, input *awsec2.CreateRouteInput) (*awsec2.CreateRouteOutput, error)
	AssociateRTWithSubnet(ctx context.Context, rtId, subnetId string) error
}

func NewProvisioner(ctx context.Context, region string) (Network, error) {
	config, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return &provisioner{
		awsec2Client: awsec2.NewFromConfig(config),
	}, nil
}