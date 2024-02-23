package eks

import (
	"context"
	"errors"

	v1 "datainfra.io/baaz/api/v1/types"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func (ec *eks) CreateVPC(ctx context.Context, params *awsec2.CreateVpcInput) (*awsec2.CreateVpcOutput, error) {
	return ec.awsec2Client.CreateVpc(ctx, params)
}

func (ec *eks) CreateSubnet(ctx context.Context, params *awsec2.CreateSubnetInput) (*awsec2.CreateSubnetOutput, error) {
	return ec.awsec2Client.CreateSubnet(ctx, params)
}

func (ec *eks) CreateSG(ctx context.Context, params *awsec2.CreateSecurityGroupInput) (*awsec2.CreateSecurityGroupOutput, error) {
	return ec.awsec2Client.CreateSecurityGroup(ctx, params)
}

func (ec *eks) CreateNAT(ctx context.Context, dp *v1.DataPlanes) (*awsec2.CreateNatGatewayOutput, error) {
	var subnetId string
	if ec.dp.Spec.CloudInfra.ProvisionNetwork {
		if len(ec.dp.Status.CloudInfraStatus.SubnetIds) == 0 {
			return nil, errors.New("DataPlane has no subnet Id specified")
		}
		subnetId = ec.dp.Status.CloudInfraStatus.SubnetIds[0]
	} else {
		if len(ec.dp.Spec.CloudInfra.Eks.SubnetIds) == 0 {
			return nil, errors.New("DataPlane has no subnet Id specified")
		}
		subnetId = ec.dp.Spec.CloudInfra.Eks.SubnetIds[0]
	}

	eIP, err := ec.CreateElasticIP(ctx)
	if err != nil {
		return nil, err
	}

	input := &awsec2.CreateNatGatewayInput{
		SubnetId:     &subnetId,
		AllocationId: eIP.AllocationId,
	}

	return ec.awsec2Client.CreateNatGateway(ctx, input)
}

func (ec *eks) CreateInternetGateway(ctx context.Context) (*awsec2.CreateInternetGatewayOutput, error) {
	return ec.awsec2Client.CreateInternetGateway(ctx, &awsec2.CreateInternetGatewayInput{})
}

func (ec *eks) AttachInternetGateway(ctx context.Context, igId, vpcId string) (*awsec2.AttachInternetGatewayOutput, error) {
	return ec.awsec2Client.AttachInternetGateway(ctx, &awsec2.AttachInternetGatewayInput{
		InternetGatewayId: &igId,
		VpcId:             &vpcId,
	})
}

func (ec *eks) CreateElasticIP(ctx context.Context) (*awsec2.AllocateAddressOutput, error) {
	return ec.awsec2Client.AllocateAddress(ctx, &awsec2.AllocateAddressInput{})
}

func (ec *eks) AssociateNATWithRT(ctx context.Context, dp *v1.DataPlanes) error {
	vpc := dp.Status.CloudInfraStatus.Vpc
	natId := dp.Status.CloudInfraStatus.NATGatewayId

	routeTables, err := ec.awsec2Client.DescribeRouteTables(ctx, &awsec2.DescribeRouteTablesInput{
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
		_, err = ec.awsec2Client.CreateRoute(ctx, &awsec2.CreateRouteInput{
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
			RouteTableId:         rt.RouteTableId,
			NatGatewayId:         &natId,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (ec *eks) DescribeVpcAttribute(ctx context.Context, vpcId string) (*awsec2.DescribeVpcsOutput, error) {
	return ec.awsec2Client.DescribeVpcs(ctx, &awsec2.DescribeVpcsInput{
		VpcIds: []string{vpcId},
	})
}

func (ec *eks) AddSGInboundRule(ctx context.Context, sgGroupId, vpcId string) (*awsec2.AuthorizeSecurityGroupIngressOutput, error) {
	vpcs, err := ec.DescribeVpcAttribute(ctx, vpcId)
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

	return ec.awsec2Client.AuthorizeSecurityGroupIngress(ctx, input)
}

func (ec *eks) SubnetAutoAssignPublicIP(ctx context.Context, subnetId string) (*awsec2.ModifySubnetAttributeOutput, error) {
	input := &awsec2.ModifySubnetAttributeInput{
		SubnetId: &subnetId,
		MapPublicIpOnLaunch: &ec2types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	}

	return ec.awsec2Client.ModifySubnetAttribute(ctx, input)
}

func (ec *eks) CreateRouteTable(ctx context.Context, vpcId string) (*awsec2.CreateRouteTableOutput, error) {
	input := &awsec2.CreateRouteTableInput{
		VpcId: &vpcId,
	}

	return ec.awsec2Client.CreateRouteTable(ctx, input)
}

func (ec *eks) CreateRoute(ctx context.Context, input *awsec2.CreateRouteInput) (*awsec2.CreateRouteOutput, error) {
	return ec.awsec2Client.CreateRoute(ctx, input)
}

func (ec *eks) AssociateRTWithSubnet(ctx context.Context, rtId, subnetId string) error {
	input := &awsec2.AssociateRouteTableInput{
		RouteTableId: &rtId,
		SubnetId:     &subnetId,
	}

	_, err := ec.awsec2Client.AssociateRouteTable(ctx, input)
	return err
}
