package eks

import (
	"context"

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
