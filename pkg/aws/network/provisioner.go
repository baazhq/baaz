package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	v1 "github.com/baazhq/baaz/api/v1/types"
)

type provisioner struct {
	awsec2Client *awsec2.Client
}

func (p *provisioner) CreateVPC(ctx context.Context, params *awsec2.CreateVpcInput) (*awsec2.CreateVpcOutput, error) {
	return p.awsec2Client.CreateVpc(ctx, params)
}

func (p *provisioner) CreateSubnet(ctx context.Context, params *awsec2.CreateSubnetInput) (*awsec2.CreateSubnetOutput, error) {
	return p.awsec2Client.CreateSubnet(ctx, params)
}

func (ec *provisioner) CreateSG(ctx context.Context, params *awsec2.CreateSecurityGroupInput) (*awsec2.CreateSecurityGroupOutput, error) {
	return ec.awsec2Client.CreateSecurityGroup(ctx, params)
}

func (p *provisioner) CreateNAT(ctx context.Context, dp *v1.DataPlanes) (*awsec2.CreateNatGatewayOutput, error) {
	var subnetId string
	if dp.Spec.CloudInfra.ProvisionNetwork {
		if len(dp.Status.CloudInfraStatus.SubnetIds) == 0 {
			return nil, errors.New("DataPlane has no subnet Id specified")
		}
		subnetId = dp.Status.CloudInfraStatus.SubnetIds[0]
	} else {
		if len(dp.Spec.CloudInfra.Eks.SubnetIds) == 0 {
			return nil, errors.New("DataPlane has no subnet Id specified")
		}
		subnetId = dp.Spec.CloudInfra.Eks.SubnetIds[0]
	}

	eIP, err := p.CreateElasticIP(ctx)
	if err != nil {
		return nil, err
	}

	input := &awsec2.CreateNatGatewayInput{
		SubnetId:     &subnetId,
		AllocationId: eIP.AllocationId,
	}

	return p.awsec2Client.CreateNatGateway(ctx, input)
}

func (p *provisioner) CreateInternetGateway(ctx context.Context) (*awsec2.CreateInternetGatewayOutput, error) {
	return p.awsec2Client.CreateInternetGateway(ctx, &awsec2.CreateInternetGatewayInput{})
}

func (p *provisioner) AttachInternetGateway(ctx context.Context, igId, vpcId string) (*awsec2.AttachInternetGatewayOutput, error) {
	return p.awsec2Client.AttachInternetGateway(ctx, &awsec2.AttachInternetGatewayInput{
		InternetGatewayId: &igId,
		VpcId:             &vpcId,
	})
}

func (p *provisioner) CreateElasticIP(ctx context.Context) (*awsec2.AllocateAddressOutput, error) {
	return p.awsec2Client.AllocateAddress(ctx, &awsec2.AllocateAddressInput{})
}

func (p *provisioner) AssociateNATWithRT(ctx context.Context, dp *v1.DataPlanes) error {
	vpc := dp.Status.CloudInfraStatus.Vpc
	natId := dp.Status.CloudInfraStatus.NATGatewayId

	routeTables, err := p.awsec2Client.DescribeRouteTables(ctx, &awsec2.DescribeRouteTablesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpc},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, rt := range routeTables.RouteTables {
		if *rt.RouteTableId == dp.Status.CloudInfraStatus.PublicRTId {
			continue
		}

		_, err = p.awsec2Client.CreateRoute(ctx, &awsec2.CreateRouteInput{
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
			RouteTableId:         rt.RouteTableId,
			NatGatewayId:         &natId,
		})
		if err != nil {
			return fmt.Errorf("failed to attach NAT with RT %s: %s", *rt.RouteTableId, err.Error())
		}
	}
	return nil
}

func (p *provisioner) DescribeVpcAttribute(ctx context.Context, vpcId string) (*awsec2.DescribeVpcsOutput, error) {
	return p.awsec2Client.DescribeVpcs(ctx, &awsec2.DescribeVpcsInput{
		VpcIds: []string{vpcId},
	})
}

func (p *provisioner) AddSGInboundRule(ctx context.Context, sgGroupId, vpcId string) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
	vpcs, err := p.DescribeVpcAttribute(ctx, vpcId)
	if err != nil {
		return nil, err
	}

	if len(vpcs.Vpcs) != 1 {
		return nil, errors.New("failed to get vpc details")
	}

	input := &awsec2.AuthorizeSecurityGroupIngressInput{
		GroupId: &sgGroupId,
		IpPermissions: []ec2types.IpPermission{
			{
				IpProtocol: aws.String(string(ec2types.TransportProtocolTcp)),
				IpRanges: []ec2types.IpRange{
					{
						CidrIp: vpcs.Vpcs[0].CidrBlock,
					},
				},
				FromPort: aws.Int32(0),
				ToPort:   aws.Int32(65535),
			},
		},
	}

	return p.awsec2Client.AuthorizeSecurityGroupIngress(ctx, input)
}

func (p *provisioner) SubnetAutoAssignPublicIP(ctx context.Context, subnetId string) (*awsec2.ModifySubnetAttributeOutput, error) {
	input := &awsec2.ModifySubnetAttributeInput{
		SubnetId: &subnetId,
		MapPublicIpOnLaunch: &ec2types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}

	return p.awsec2Client.ModifySubnetAttribute(ctx, input)
}

func (p *provisioner) CreateRouteTable(ctx context.Context, vpcId string) (*awsec2.CreateRouteTableOutput, error) {
	input := &awsec2.CreateRouteTableInput{
		VpcId: &vpcId,
	}

	return p.awsec2Client.CreateRouteTable(ctx, input)
}

func (p *provisioner) CreateRoute(ctx context.Context, input *awsec2.CreateRouteInput) (*awsec2.CreateRouteOutput, error) {
	return p.awsec2Client.CreateRoute(ctx, input)
}

func (p *provisioner) AssociateRTWithSubnet(ctx context.Context, rtId, subnetId string) error {
	input := &awsec2.AssociateRouteTableInput{
		RouteTableId: &rtId,
		SubnetId:     &subnetId,
	}

	_, err := p.awsec2Client.AssociateRouteTable(ctx, input)
	return err
}
