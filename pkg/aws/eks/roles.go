package eks

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

var (
	ebsCSIPolicyARN = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
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

var ebsCSIRoleTrustJsonTemplate = `
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Federated": "arn:aws:iam::{{ .AccountID }}:oidc-provider/{{ .OIDCProvider }}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "$OIDC_PROVIDER:sub": "system:serviceaccount:kube-system:ebs-csi-controller-sa", 
                    "$OIDC_PROVIDER:aud": "sts.amazonaws.com"
                }
            }
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

type EBSCSIRoleTemplateInput struct {
	AccountID    string
	OIDCProvider string
}

func makeEksClusterRoleName(clusterName string) string { return clusterName + "-" + "cluster-role" }
func makeEksNodeRoleName(nodeGroupName string) string  { return nodeGroupName + "-" + "node-role" }
func makeEBSCSIRoleName(region, clusterName string) string {
	return region + "-" + clusterName + "ebs-role"
}

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
		// for role error it seems
		// the error is not considered as ResourceNotFoundException
		resultCreateRole, cerr := iamClient.CreateRole(eksEnv.Context, &awsiam.CreateRoleInput{
			RoleName:                 aws.String(makeEksClusterRoleName(eksEnv.Env.Spec.CloudInfra.Eks.Name)),
			AssumeRolePolicyDocument: aws.String(strings.TrimSpace(assumeClusterRolePolicy)),
		})
		if cerr != nil {
			return nil, cerr
		}

		for _, clusterRolePolicyArn := range clusterRolePolicyArns {
			_, cerr := iamClient.AttachRolePolicy(eksEnv.Context, &awsiam.AttachRolePolicyInput{
				RoleName:  resultCreateRole.Role.RoleName,
				PolicyArn: &clusterRolePolicyArn,
			})
			if cerr != nil {
				return nil, cerr
			}
		}

		return nil, err
	}

	return result, nil
}

func (eksEnv *EksEnvironment) createEbsCSIRole(ctx context.Context) (*awsiam.CreateRoleOutput, error) {
	oidcProvider := eksEnv.Env.Status.CloudInfraStatus.AwsCloudInfraConfigStatus.EksStatus.OIDCProviderArn
	accountID, err := eksEnv.getAccountID(ctx)
	if err != nil {
		return nil, err
	}

	tmpl := template.New(ebsCSIRoleTrustJsonTemplate)
	var tmplOutput bytes.Buffer

	if err := tmpl.Execute(&tmplOutput, EBSCSIRoleTemplateInput{
		AccountID:    accountID,
		OIDCProvider: oidcProvider,
	}); err != nil {
		return nil, err
	}

	roleName := makeEBSCSIRoleName(eksEnv.Env.Spec.CloudInfra.AwsRegion, eksEnv.Env.Spec.CloudInfra.Eks.Name)
	trustPolicy := tmplOutput.String()

	roleInput := awsiam.CreateRoleInput{
		AssumeRolePolicyDocument: &trustPolicy,
		RoleName:                 &roleName,
	}

	iamClient := awsiam.NewFromConfig(eksEnv.Config)
	roleOutput, err := iamClient.CreateRole(ctx, &roleInput)
	if err != nil {
		return nil, err
	}

	attachPolicyInput := awsiam.AttachRolePolicyInput{
		PolicyArn: &ebsCSIPolicyARN,
		RoleName:  roleOutput.Role.RoleName,
	}
	if _, err := iamClient.AttachRolePolicy(ctx, &attachPolicyInput); err != nil {
		return nil, err
	}
	return roleOutput, nil
}
