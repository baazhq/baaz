package eks

import (
	"context"
	"errors"

	v1 "datainfra.io/baaz/api/v1/types"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
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

func (ec *eks) CreateElasticIP(ctx context.Context) (*awsec2.AllocateAddressOutput, error) {
	return ec.awsec2Client.AllocateAddress(ctx, &awsec2.AllocateAddressInput{})
}
