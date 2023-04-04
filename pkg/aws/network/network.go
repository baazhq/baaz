package aws_network

import (
	"context"
	"fmt"
	"log"

	v1 "datainfra.io/ballastdata/api/v1"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type NetworkEnvironment struct {
	Env    *v1.Environment
	Config aws.Config
}

type NetworkOutput struct {
	Properties map[string]string
}

func NewConfig(awsRegion string) *aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &cfg
}

func (e *NetworkEnvironment) CreateNetwork() (networkOutput NetworkOutput, err error) {
	vpcOutput, err := e.makeVpc()
	if err != nil {
		return NetworkOutput{}, err
	}

	subnetId, err := e.makeSubnet(vpcOutput)
	if err != nil {
		return NetworkOutput{}, err
	}

	fmt.Println(subnetId)

	return
}

func (e *NetworkEnvironment) makeVpc() (*ec2.CreateVpcOutput, error) {
	ec2Client := ec2.NewFromConfig(e.Config)

	vpcInput := ec2.CreateVpcInput{
		CidrBlock: aws.String(e.Env.Spec.CloudInfra.Network.CidrBlock),
		TagSpecifications: []types.TagSpecification{

			{
				ResourceType: types.ResourceTypeVpc,
				Tags: []types.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String(e.Env.Spec.CloudInfra.Network.VpcName),
					},
				},
			},
		},
	}

	vpcOutput, err := ec2Client.CreateVpc(context.TODO(), &vpcInput)
	if err != nil {
		return nil, err
	}

	return vpcOutput, nil
}

func (e *NetworkEnvironment) makeSubnet(vpcOutput *ec2.CreateVpcOutput) ([]string, error) {
	ec2Client := ec2.NewFromConfig(e.Config)

	var subnetId []string
	for _, subnet := range e.Env.Spec.CloudInfra.Network.Subnets {
		subnet := ec2.CreateSubnetInput{
			VpcId:     vpcOutput.Vpc.VpcId,
			CidrBlock: &subnet.CidrBlock,
			TagSpecifications: []types.TagSpecification{
				{
					ResourceType: types.ResourceTypeSubnet,
					Tags: []types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(subnet.Name),
						},
					},
				},
			},
		}
		subnetOutput, err := ec2Client.CreateSubnet(context.TODO(), &subnet)
		if err != nil {
			return nil, err
		}

		subnetId = append(subnetId, *subnetOutput.Subnet.SubnetId)

	}

	return subnetId, nil

}
