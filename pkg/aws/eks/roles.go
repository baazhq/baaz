package eks

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
)

var clusterRolePolicyArns = []string{
	"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
	"arn:aws:iam::aws:policy/AmazonEKSVPCResourceController",
}

var nodeRolePolicyArns = []string{
	"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
	"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
	"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
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
                "Federated": "arn:aws:iam::{{.AccountID}}:oidc-provider/{{.OIDCProvider}}"
            },
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "{{.OIDCProvider}}:sub": "system:serviceaccount:kube-system:ebs-csi-controller-sa", 
                    "{{.OIDCProvider}}:aud": "sts.amazonaws.com"
                }
            }
        }
    ]
}
`

var (
	ebsCSIPolicyARN = "arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"
)

var (
	vpcCNIPolicyARN = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
)

type genericRoleTemplateInput struct {
	AccountID    string
	OIDCProvider string
}

var (
	vpcCniTrustPolicy = `
	{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Federated": "arn:aws:iam::{{.AccountID}}:oidc-provider/{{.OIDCProvider}}"
				},
				"Action": "sts:AssumeRoleWithWebIdentity",
				"Condition": {
					"StringEquals": {
						"{{.OIDCProvider}}:aud": "sts.amazonaws.com",
						"{{.OIDCProvider}}:sub": "system:serviceaccount:kube-system:aws-node"
					}
				}
			}
		]
	}`
)

func (ec *eks) CreateClusterIamRole() (*awsiam.GetRoleOutput, error) {

	awsIamGetRoleOutput, err := ec.awsIamClient.GetRole(ec.ctx, &awsiam.GetRoleInput{
		RoleName: aws.String(MakeEksClusterRoleName(ec.dp.Spec.CloudInfra.Eks.Name)),
	})

	if err != nil {
		// for role error it seems
		// the error is not considered as ResourceNotFoundException
		resultCreateRole, cerr := ec.awsIamClient.CreateRole(ec.ctx, &awsiam.CreateRoleInput{
			RoleName:                 aws.String(MakeEksClusterRoleName(ec.dp.Spec.CloudInfra.Eks.Name)),
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

		return ec.CreateClusterIamRole()
	}

	return awsIamGetRoleOutput, nil
}

func (ec *eks) CreateNodeIamRole(name string) (*awsiam.GetRoleOutput, error) {

	result, err := ec.awsIamClient.GetRole(ec.ctx, &awsiam.GetRoleInput{
		RoleName: aws.String(MakeEksNodeRoleName(name)),
	})
	if err != nil {
		resultCreateRole, cerr := ec.awsIamClient.CreateRole(ec.ctx, &awsiam.CreateRoleInput{
			RoleName:                 aws.String(MakeEksNodeRoleName(name)),
			AssumeRolePolicyDocument: aws.String(strings.TrimSpace(assumeNodeRolePolicy)),
		})
		if cerr != nil {
			return nil, cerr
		}

		for _, nodeRolePolicyArn := range nodeRolePolicyArns {
			_, cerr := ec.awsIamClient.AttachRolePolicy(ec.ctx, &awsiam.AttachRolePolicyInput{
				RoleName:  resultCreateRole.Role.RoleName,
				PolicyArn: &nodeRolePolicyArn,
			})
			if cerr != nil {
				return nil, cerr
			}
		}

		return nil, err
	}

	return result, nil
}

func (ec *eks) CreateEbsCSIRole(ctx context.Context) (*awsiam.CreateRoleOutput, error) {
	oidcProvider := ec.dp.Status.CloudInfraStatus.AwsCloudInfraConfigStatus.EksStatus.OIDCProviderArn

	roleName := MakeEBSCSIRoleName(ec.dp.Spec.CloudInfra.Region, ec.dp.Spec.CloudInfra.Eks.Name)

	_, err := ec.awsIamClient.GetRole(ec.ctx, &awsiam.GetRoleInput{
		RoleName: aws.String(roleName),
	})

	if err != nil {
		_, oidcProviderURL, found := strings.Cut(oidcProvider, "oidc-provider/")
		if !found {
			return nil, errors.New("invalid oidc provider arn")
		}
		accountID, err := ec.getAccountID()
		if err != nil {
			return nil, err
		}

		tmpl, err := template.New("ebs-template").Parse(ebsCSIRoleTrustJsonTemplate)
		if err != nil {
			return nil, err
		}
		var tmplOutput bytes.Buffer

		if err := tmpl.Execute(&tmplOutput, genericRoleTemplateInput{
			AccountID:    accountID,
			OIDCProvider: oidcProviderURL,
		}); err != nil {
			return nil, err
		}

		trustPolicy := tmplOutput.String()

		roleInput := awsiam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(strings.TrimSpace(trustPolicy)),
			RoleName:                 &roleName,
		}

		roleOutput, err := ec.awsIamClient.CreateRole(ctx, &roleInput)
		if err != nil {
			return nil, err
		}

		attachPolicyInput := awsiam.AttachRolePolicyInput{
			PolicyArn: &ebsCSIPolicyARN,
			RoleName:  roleOutput.Role.RoleName,
		}

		if _, err := ec.awsIamClient.AttachRolePolicy(ctx, &attachPolicyInput); err != nil {
			return nil, err
		}
		return roleOutput, nil
	}

	return &awsiam.CreateRoleOutput{}, nil
}

func (ec *eks) CreateVpcCniRole(ctx context.Context) (roleOutput *awsiam.CreateRoleOutput, arn string, err error) {
	oidcProvider := ec.dp.Status.CloudInfraStatus.AwsCloudInfraConfigStatus.EksStatus.OIDCProviderArn

	roleName := MakeVpcCniRoleName(ec.dp.Spec.CloudInfra.Region, ec.dp.Spec.CloudInfra.Eks.Name)

	_, err = ec.awsIamClient.GetRole(ec.ctx, &awsiam.GetRoleInput{
		RoleName: aws.String(roleName),
	})

	if err != nil {
		_, oidcProviderURL, found := strings.Cut(oidcProvider, "oidc-provider/")
		if !found {
			return nil, "", errors.New("invalid oidc provider arn")
		}
		accountID, err := ec.getAccountID()
		if err != nil {
			return nil, "", err
		}

		tmpl, err := template.New("vpcni-template").Parse(vpcCniTrustPolicy)
		if err != nil {
			return nil, "", err
		}
		var tmplOutput bytes.Buffer

		if err := tmpl.Execute(&tmplOutput, genericRoleTemplateInput{
			AccountID:    accountID,
			OIDCProvider: oidcProviderURL,
		}); err != nil {
			return nil, "", err
		}

		trustPolicy := tmplOutput.String()

		roleInput := awsiam.CreateRoleInput{
			AssumeRolePolicyDocument: aws.String(strings.TrimSpace(trustPolicy)),
			RoleName:                 &roleName,
		}

		roleOutput, err := ec.awsIamClient.CreateRole(ctx, &roleInput)
		if err != nil {
			return nil, "", err
		}

		attachPolicyInput := awsiam.AttachRolePolicyInput{
			PolicyArn: &vpcCNIPolicyARN,
			RoleName:  roleOutput.Role.RoleName,
		}

		if _, err := ec.awsIamClient.AttachRolePolicy(ctx, &attachPolicyInput); err != nil {
			return nil, "", err
		}

		return roleOutput, fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName), nil
	}

	return &awsiam.CreateRoleOutput{}, "", nil
}

func (ec *eks) CreateIAMPolicy(ctx context.Context, input *iam.CreatePolicyInput) (*iam.CreatePolicyOutput, error) {
	return ec.awsIamClient.CreatePolicy(ctx, input)
}

func (ec *eks) AttachRolePolicy(ctx context.Context, input *iam.AttachRolePolicyInput) (*iam.AttachRolePolicyOutput, error) {
	return ec.awsIamClient.AttachRolePolicy(ctx, input)
}
