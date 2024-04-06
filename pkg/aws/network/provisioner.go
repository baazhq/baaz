package network

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	v1 "github.com/baazhq/baaz/api/v1/types"
	"k8s.io/klog/v2"
)

type provisioner struct {
	awsec2Client *awsec2.Client
	elbv2Client  *elbv2.Client
}

func (p *provisioner) CreateVPC(ctx context.Context, params *awsec2.CreateVpcInput) (*awsec2.CreateVpcOutput, error) {
	return p.awsec2Client.CreateVpc(ctx, params)
}

func (p *provisioner) DeleteVPC(ctx context.Context, vpcId string) error {
	var notFoundErr *types.ResourceNotFoundException
	_, err := p.awsec2Client.DeleteVpc(ctx, &awsec2.DeleteVpcInput{
		VpcId: aws.String(vpcId),
	})
	if err != nil && !errors.As(err, &notFoundErr) {
		return err
	}
	return nil
}

func (p *provisioner) CreateSubnet(ctx context.Context, params *awsec2.CreateSubnetInput) (*awsec2.CreateSubnetOutput, error) {
	return p.awsec2Client.CreateSubnet(ctx, params)
}

func (p *provisioner) DeleteSubnets(ctx context.Context, subnetIds []string) error {
	for _, id := range subnetIds {
		_, err := p.awsec2Client.DeleteSubnet(ctx, &awsec2.DeleteSubnetInput{
			SubnetId: aws.String(id),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *provisioner) CreateSG(ctx context.Context, params *awsec2.CreateSecurityGroupInput) (*awsec2.CreateSecurityGroupOutput, error) {
	return p.awsec2Client.CreateSecurityGroup(ctx, params)
}

func (p *provisioner) DeleteSGs(ctx context.Context, vpcId string) error {
	sGroups, err := p.awsec2Client.DescribeSecurityGroups(ctx, &awsec2.DescribeSecurityGroupsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})
	if err != nil {
		return err
	}
	for _, sg := range sGroups.SecurityGroups {
		if _, err := p.awsec2Client.DeleteSecurityGroup(ctx, &awsec2.DeleteSecurityGroupInput{
			GroupId: sg.GroupId,
		}); err != nil {
			return err
		}
	}
	return nil
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

func (p *provisioner) DeleteNatGateway(ctx context.Context, id string) error {
	nats, err := p.awsec2Client.DescribeNatGateways(ctx, &awsec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{id},
	})
	if err != nil {
		return err
	}

	input := &awsec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(id),
	}

	_, err = p.awsec2Client.DeleteNatGateway(ctx, input)
	if err != nil {
		return err
	}

	if len(nats.NatGateways) == 1 {
		eips := nats.NatGateways[0].NatGatewayAddresses
		for _, ip := range eips {
			if ip.AllocationId != nil {
				_, err := p.awsec2Client.ReleaseAddress(ctx, &awsec2.ReleaseAddressInput{
					AllocationId: ip.AllocationId,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *provisioner) CreateInternetGateway(ctx context.Context) (*awsec2.CreateInternetGatewayOutput, error) {
	return p.awsec2Client.CreateInternetGateway(ctx, &awsec2.CreateInternetGatewayInput{})
}

func (p *provisioner) DeleteInternetGateway(ctx context.Context, id string) error {
	_, err := p.awsec2Client.DeleteInternetGateway(ctx, &awsec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(id),
	})
	return err
}

func (p *provisioner) AttachInternetGateway(ctx context.Context, igId, vpcId string) (*awsec2.AttachInternetGatewayOutput, error) {
	return p.awsec2Client.AttachInternetGateway(ctx, &awsec2.AttachInternetGatewayInput{
		InternetGatewayId: &igId,
		VpcId:             &vpcId,
	})
}

func (p *provisioner) DetachInternetGateway(ctx context.Context, id, vpcId string) error {
	_, err := p.awsec2Client.DetachInternetGateway(ctx, &awsec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(id),
		VpcId:             aws.String(vpcId),
	})
	return err
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

func (p *provisioner) DeleteRouteTables(ctx context.Context, vpcId string) error {
	routes, err := p.awsec2Client.DescribeRouteTables(ctx, &awsec2.DescribeRouteTablesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	})
	if err != nil {
		return err
	}

	for _, rt := range routes.RouteTables {
		for _, r := range rt.Routes {
			_, err := p.awsec2Client.DeleteRoute(ctx, &awsec2.DeleteRouteInput{
				RouteTableId:         rt.RouteTableId,
				DestinationCidrBlock: r.DestinationCidrBlock,
			})
			if err != nil {
				klog.Error(err)
				continue
			}
		}

		_, err = p.awsec2Client.DeleteRouteTable(ctx, &awsec2.DeleteRouteTableInput{
			RouteTableId: rt.RouteTableId,
		})
		if err != nil {
			klog.Error(err)
			continue
		}
	}
	return nil
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

func (p *provisioner) DeleteVpcLBs(ctx context.Context, vpcId string) error {
	lbs, err := p.elbv2Client.DescribeLoadBalancers(ctx, &elbv2.DescribeLoadBalancersInput{})
	if err != nil {
		return err
	}

	fmt.Println(lbs.LoadBalancers)

	for _, lb := range lbs.LoadBalancers {
		fmt.Println("================================")
		fmt.Println(*lb.LoadBalancerArn)
		if aws.ToString(lb.VpcId) == vpcId {
			_, err := p.elbv2Client.DeleteLoadBalancer(ctx, &elbv2.DeleteLoadBalancerInput{
				LoadBalancerArn: lb.LoadBalancerArn,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
