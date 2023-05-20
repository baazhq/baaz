package eks

import (
	"context"
	"errors"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (eksEnv *EksEnvironment) ReconcileDefaultAddons() error {
	oidcProvider := eksEnv.Env.Status.CloudInfraStatus.AwsCloudInfraConfigStatus.EksStatus.OIDCProviderArn
	if oidcProvider == "" {
		klog.Info("ebs-csi-driver creation: waiting for oidcProvider to be created")
		return nil
	}
	clusterName := eksEnv.Env.Spec.CloudInfra.Eks.Name
	ebsAddon, err := eksEnv.describeAddon(eksEnv.Context, "aws-ebs-csi-driver", eksEnv.Env.Spec.CloudInfra.Eks.Name)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			klog.Info("Creating aws-ebs-csi-driver addon")
			_, cErr := eksEnv.createAddon(eksEnv.Context, &CreateAddonInput{
				Name:        "aws-ebs-csi-driver",
				ClusterName: clusterName,
			})
			if cErr != nil {
				return cErr
			}
			klog.Info("aws-ebs-csi-driver addon creation is initiated")
		} else {
			return err
		}
		return nil
	}
	if ebsAddon != nil && ebsAddon.Addon != nil {
		addonRes := ebsAddon.Addon
		klog.Info("aws-ebs-csi-driver addon status: ", addonRes.Status)
		if err := eksEnv.patchAddonStatus(*addonRes.AddonName, string(addonRes.Status)); err != nil {
			return err
		}
	}

	return nil
}

func (eksEnv *EksEnvironment) patchAddonStatus(addonName, status string) error {
	// update status with current addon status
	_, _, err := utils.PatchStatus(eksEnv.Context, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		if in.Status.AddonStatus == nil {
			in.Status.AddonStatus = make(map[string]string)
		}
		in.Status.AddonStatus[addonName] = status
		return in
	})
	return err
}

func (eksEnv *EksEnvironment) describeAddon(ctx context.Context, addonName, clusterName string) (*eks.DescribeAddonOutput, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)

	input := &awseks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}
	result, err := eksClient.DescribeAddon(ctx, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type CreateAddonInput struct {
	Name        string `json:"name"`
	ClusterName string `json:"clusterName"`
}

func (eksEnv *EksEnvironment) createAddon(ctx context.Context, params *CreateAddonInput) (*eks.CreateAddonOutput, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)

	role, err := eksEnv.createEbsCSIRole(ctx)
	if err != nil {
		return nil, err
	}

	input := &awseks.CreateAddonInput{
		AddonName:             aws.String(params.Name),
		ClusterName:           aws.String(params.ClusterName),
		ResolveConflicts:      types.ResolveConflictsOverwrite,
		ServiceAccountRoleArn: role.Role.Arn,
	}
	result, err := eksClient.CreateAddon(ctx, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}
