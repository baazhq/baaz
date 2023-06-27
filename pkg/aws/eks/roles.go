package eks

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

var clusterRolePolicyArns = []string{
	"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	"arn:aws:iam::aws:policy/AmazonEKSVPCResourceController",
}

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

func (ec *eks) createClusterIamRole() (*awsiam.GetRoleOutput, error) {

	awsIamGetRoleOutput, err := ec.awsIamClient.GetRole(ec.ctx, &awsiam.GetRoleInput{
		RoleName: aws.String(makeEksClusterRoleName(ec.environment.Spec.CloudInfra.Eks.Name)),
	})

	if err != nil {
		// for role error it seems
		// the error is not considered as ResourceNotFoundException
		resultCreateRole, cerr := ec.awsIamClient.CreateRole(ec.ctx, &awsiam.CreateRoleInput{
			RoleName:                 aws.String(makeEksClusterRoleName(ec.environment.Spec.CloudInfra.Eks.Name)),
			AssumeRolePolicyDocument: aws.String(strings.TrimSpace(assumeClusterRolePolicy)),
		})
		if cerr != nil {
			return nil, cerr
		}

		for _, clusterRolePolicyArn := range clusterRolePolicyArns {
			_, cerr := ec.awsIamClient.AttachRolePolicy(ec.ctx, &awsiam.AttachRolePolicyInput{
				RoleName:  resultCreateRole.Role.RoleName,
				PolicyArn: &clusterRolePolicyArn,
			})
			if cerr != nil {
				return nil, cerr
			}
		}

		return nil, err
	}

	return awsIamGetRoleOutput, nil
}
