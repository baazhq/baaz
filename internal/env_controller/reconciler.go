package controller

import (
	"context"
	"fmt"
	"os"

	v1 "datainfra.io/ballastdata/api/v1"
	aws_network "datainfra.io/ballastdata/pkg/aws/network"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileEnvironment(client client.Client, env *v1.Environment) error {

	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAWLZK4B6ACNA3H43S")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "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")
	err := CreateEnvironment(context.TODO(), env)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func CreateEnvironment(ctx context.Context, env *v1.Environment) error {

	network := aws_network.NetworkEnvironment{
		Env:    env,
		Config: *aws_network.NewConfig(env.Spec.CloudInfra.AwsRegion),
	}

	networkOutput, err := network.CreateNetwork()
	if err != nil {
		return err
	}

	fmt.Println(networkOutput)

	// subnet := ec2.CreateSubnetInput{
	// 	VpcId:     vpcOutput.Vpc.VpcId,
	// 	CidrBlock: aws.String("10.0.0.0/24"),
	// }

	//	subnetOutput, _ := ec2Client.CreateSubnet(context.TODO(), &subnet)

	// input := &eks.CreateClusterInput{
	// 	ClientRequestToken: aws.String("1d2129a1-3d38-460a-9756-e5b91fddb951"),
	// 	Name:               aws.String("prod"),
	// 	ResourcesVpcConfig: &types.VpcConfigRequest{
	// 		SecurityGroupIds: []string{"sg-6979fe18"},
	// 		SubnetIds:        []string{"subnet-6782e71e", "subnet-e7e761ac"},
	// 	},
	// 	RoleArn: aws.String("arn:aws:iam::012345678910:role/eks-service-role-AWSServiceRoleForAmazonEKS-J7ONKE3BQ4PI"),
	// 	Version: aws.String("1.22"),
	// }

	// result, err := svc.CreateCluster(context.TODO(), input)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	return nil
}
