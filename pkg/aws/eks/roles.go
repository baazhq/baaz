package eks

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

var assumeClusterRolePolicy string = `
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
               "Service": "eks.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
`
var assumeNodeRolePolicy string = `
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
`

var nodeRolePolicyArns = []string{
	"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
	"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
	"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
}

var clusterRolePolicyArns = []string{
	"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	"arn:aws:iam::aws:policy/AmazonEKSVPCResourceController",
}

func makeEksClusterRoleName(clusterName string) string { return clusterName + "-" + "cluster-role" }
func makeEksNodeRoleName(nodeGroupName string) string  { return nodeGroupName + "-" + "node-role" }

func (eksEnv *EksEnvironment) createNodeIamRole(name string) (*awsiam.GetRoleOutput, error) {
	iamClient := awsiam.NewFromConfig(eksEnv.Config)

	result, err := iamClient.GetRole(eksEnv.Context, &awsiam.GetRoleInput{
		RoleName: aws.String(makeEksNodeRoleName(eksEnv.Env.Spec.CloudInfra.Eks.Name)),
	})
	if err != nil {
		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {

			resultCreateRole, err := iamClient.CreateRole(eksEnv.Context, &awsiam.CreateRoleInput{
				RoleName:                 aws.String(makeEksNodeRoleName(name)),
				AssumeRolePolicyDocument: aws.String(strings.TrimSpace(assumeNodeRolePolicy)),
			})
			if err != nil {
				return nil, err
			}

			for _, nodeRolePolicyArn := range nodeRolePolicyArns {
				_, err := iamClient.AttachRolePolicy(eksEnv.Context, &awsiam.AttachRolePolicyInput{
					RoleName:  resultCreateRole.Role.RoleName,
					PolicyArn: &nodeRolePolicyArn,
				})
				if err != nil {
					return nil, err
				}
			}

			if err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	return result, nil
}

func (eksEnv *EksEnvironment) createClusterIamRole() (*awsiam.GetRoleOutput, error) {
	iamClient := awsiam.NewFromConfig(eksEnv.Config)

	result, err := iamClient.GetRole(eksEnv.Context, &awsiam.GetRoleInput{
		RoleName: aws.String(makeEksClusterRoleName(eksEnv.Env.Spec.CloudInfra.Eks.Name)),
	})
	if err != nil {
		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {

			resultCreateRole, err := iamClient.CreateRole(eksEnv.Context, &awsiam.CreateRoleInput{
				RoleName:                 aws.String(makeEksClusterRoleName(eksEnv.Env.Spec.CloudInfra.Eks.Name)),
				AssumeRolePolicyDocument: aws.String(strings.TrimSpace(assumeClusterRolePolicy)),
			})
			if err != nil {
				return nil, err
			}

			for _, clusterRolePolicyArn := range clusterRolePolicyArns {
				_, err := iamClient.AttachRolePolicy(eksEnv.Context, &awsiam.AttachRolePolicyInput{
					RoleName:  resultCreateRole.Role.RoleName,
					PolicyArn: &clusterRolePolicyArn,
				})
				if err != nil {
					return nil, err
				}
			}

			if err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	return result, nil
}
