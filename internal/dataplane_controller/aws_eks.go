package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	v1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/pkg/aws/eks"
	"datainfra.io/baaz/pkg/helm"
	"datainfra.io/baaz/pkg/store"
	"datainfra.io/baaz/pkg/utils"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	awsEbsCsiDriver string = "aws-ebs-csi-driver"
	vpcCni          string = "vpc-cni"
)

func (r *DataPlaneReconciler) reconcileAwsEnvironment(ctx context.Context, dp *v1.DataPlanes) error {

	eksClient := eks.NewEks(ctx, dp)

	awsEnv := awsEnv{
		ctx:    ctx,
		dp:     dp,
		eksIC:  eksClient,
		client: r.Client,
		store:  r.NgStore,
	}

	if err := awsEnv.reconcileAwsEks(); err != nil {
		return err
	}

	// bootstrap dataplane with apps
	if err := awsEnv.reconcileAwsApplications(); err != nil {
		return err
	}

	return nil
}

type awsEnv struct {
	ctx    context.Context
	dp     *v1.DataPlanes
	eksIC  eks.Eks
	client client.Client
	store  store.Store
}

func (ae *awsEnv) reconcileAwsEks() error {

	eksDescribeClusterOutput, err := ae.eksIC.DescribeEks()
	if err != nil {
		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {
			klog.Infof("Creating EKS Control plane: %s for Environment: %s/%s", ae.dp.Spec.CloudInfra.Eks.Name, ae.dp.Namespace, ae.dp.Name)
			klog.Info("Updating Environment status to creating")

			clusterRoleOutput, err := ae.eksIC.CreateClusterIamRole()
			if err != nil {
				return err
			}

			klog.Infof("Cluster Role [%s] Created", *clusterRoleOutput.Role.RoleName)

			createEksResult := ae.eksIC.CreateEks()
			if createEksResult.Success {
				if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
					in := obj.(*v1.DataPlanes)
					in.Status.Phase = v1.CreatingD
					in.Status.Conditions = in.AddCondition(v1.DataPlaneCondition{
						Type:               v1.ControlPlaneCreateInitiated,
						Status:             corev1.ConditionTrue,
						LastUpdateTime:     metav1.Time{Time: time.Now()},
						LastTransitionTime: metav1.Time{Time: time.Now()},
						Reason:             string(eks.EksControlPlaneCreationInitatedReason),
						Message:            string(eks.EksControlPlaneCreationInitatedMsg),
					})
					return in
				}); err != nil {
					return err
				}

				klog.Info("Successfully initiated kubernetes control plane")
				return nil

			} else {
				klog.Errorf(createEksResult.Result)
			}

		} else {
			klog.Errorf("Dataplane Describe Fail [%s]", err.Error())
		}

	}
	if eksDescribeClusterOutput != nil {
		if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusActive {
			// checking for version upgrade
			statusVersion := ae.dp.Status.Version
			specVersion := ae.dp.Spec.CloudInfra.Eks.Version
			if statusVersion != "" && statusVersion != specVersion && *eksDescribeClusterOutput.Cluster.Version != specVersion {
				klog.Info("Updating Kubernetes version to: ", ae.dp.Spec.CloudInfra.Eks.Version)
				if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
					in := obj.(*v1.DataPlanes)
					in.Status.Phase = v1.UpdatingD
					in.Status.Conditions = in.AddCondition(v1.DataPlaneCondition{
						Type:               v1.VersionUpgradeInitiated,
						Status:             corev1.ConditionTrue,
						LastUpdateTime:     metav1.Time{Time: time.Now()},
						LastTransitionTime: metav1.Time{Time: time.Now()},
						Reason:             string(eks.EksControlPlaneUpgradedReason),
						Message:            string(eks.EksControlPlaneUpgradedIntiatedMsg),
					})
					return in
				}); err != nil {
					return err
				}
				result := ae.eksIC.UpdateEks()
				if !result.Success {
					return errors.New(result.Result)
				}
				klog.Info("Successfully initiated version update")
			}

			klog.Info("Sync Cluster status and version")

			if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				in.Status.Phase = v1.ActiveD
				in.Status.Version = in.Spec.CloudInfra.Eks.Version
				in.Status.Conditions = in.AddCondition(v1.DataPlaneCondition{
					Type:               v1.ControlPlaneCreated,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     metav1.Time{Time: time.Now()},
					LastTransitionTime: metav1.Time{Time: time.Now()},
					Reason:             string(eks.EksControlPlaneCreatedReason),
					Message:            string(eks.EksControlPlaneCreatedMsg),
				})
				return in
			}); err != nil {
				return err
			}

		} else if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusCreating {
			klog.Infof("EKS Cluster Control Plane [%s] in creating state", ae.dp.Spec.CloudInfra.Eks.Name)
		} else if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusUpdating {
			klog.Infof("EKS Cluster Control Plane [%s] in updated state", ae.dp.Spec.CloudInfra.Eks.Name)
		} else if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusDeleting {
			klog.Infof("EKS Cluster Control Plane [%s] in deleting state", ae.dp.Spec.CloudInfra.Eks.Name)
		}

	}

	if eksDescribeClusterOutput != nil && eksDescribeClusterOutput.Cluster != nil && eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusActive {

		oidcOutput, err := ae.reconcileOIDCProvider(eksDescribeClusterOutput)
		if err != nil {
			return err
		}

		if oidcOutput != nil && oidcOutput.OpenIDConnectProviderArn != nil {
			_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				in.Status.CloudInfraStatus.EksStatus.OIDCProviderArn = *oidcOutput.OpenIDConnectProviderArn
				return in
			})
		}

		if err := ae.reconcileSystemNodeGroup(); err != nil {
			return err
		}

		if err := ae.ReconcileDefaultAddons(); err != nil {
			return err
		}

		if err := ae.reconcilePhase(); err != nil {
			return err
		}
	}

	return nil

}

// bootstrap applications for aws eks dataplanes
// check if system nodepool is active
// once its active install applications
func (ae *awsEnv) reconcileAwsApplications() error {

	if ae.dp.Status.NodegroupStatus[ae.dp.Spec.CloudInfra.Eks.Name+"-system"] == string(types.NodegroupStatusActive) {

		for _, app := range ae.dp.Spec.Applications {

			helm := helm.NewHelm(
				app.Name,
				app.Namespace,
				app.Spec.ChartName,
				app.Spec.RepoName,
				app.Spec.RepoUrl,
				app.Spec.Values,
			)

			restConfig, err := ae.eksIC.GetRestConfig()
			if err != nil {
				return err
			}

			_, exists := helm.List(restConfig)

			// if _, _, err := utils.PatchStatus(
			// 	ae.ctx, ae.client, ae.App, func(obj client.Object) client.Object {
			// 	in := obj.(*v1.Applications)
			// 	in.Status.Phase = v1.ApplicationPhase(result)
			// 	in.Status.ApplicationCurrentSpec = a.App.Spec
			// 	return in
			// }); err != nil {
			// 	return err
			// }

			if exists == false {
				go func() error {
					err = helm.Apply(restConfig)
					if err != nil {
						fmt.Println(err)
						return err
					}
					return nil
				}()
				return err
			}

		}
	}
	return nil
}

func (ae *awsEnv) reconcileOIDCProvider(clusterOutput *awseks.DescribeClusterOutput) (*awsiam.CreateOpenIDConnectProviderOutput, error) {
	if clusterOutput == nil || clusterOutput.Cluster == nil ||
		clusterOutput.Cluster.Identity == nil || clusterOutput.Cluster.Identity.Oidc == nil {
		return nil, errors.New("oidc provider url not found in cluster output")
	}
	oidcProviderUrl := *clusterOutput.Cluster.Identity.Oidc.Issuer

	// Compute the SHA-1 thumbprint of the OIDC provider certificate
	thumbprint, err := getIssuerCAThumbprint(oidcProviderUrl)
	if err != nil {
		return nil, err
	}

	input := &eks.CreateOIDCProviderInput{
		URL:            oidcProviderUrl,
		ThumbPrintList: []string{thumbprint},
	}

	oidcProviderArn := ae.dp.Status.CloudInfraStatus.EksStatus.OIDCProviderArn

	if oidcProviderArn != "" {
		// oidc provider is previously created
		// looking for it
		providers, err := ae.eksIC.ListOIDCProvider()
		if err != nil {
			return nil, err
		}

		for _, oidc := range providers.OpenIDConnectProviderList {
			if *oidc.Arn == ae.dp.Status.CloudInfraStatus.EksStatus.OIDCProviderArn {
				// oidc provider is already created and existed
				return nil, nil
			}
		}
	}

	result, err := ae.eksIC.CreateOIDCProvider(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ae *awsEnv) reconcilePhase() error {
	klog.Info("Calculating Environment Status")

	for node, status := range ae.dp.Status.NodegroupStatus {
		if status != string(types.NodegroupStatusActive) {
			klog.Infof("Node %s not active yet", node)
			return nil
		}
	}

	for addon, status := range ae.dp.Status.AddonStatus {
		if status != string(types.AddonStatusActive) {
			klog.Infof("Addon %s not active yet", addon)
			return nil
		}
	}

	if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
		in := obj.(*v1.DataPlanes)
		in.Status.Phase = v1.ActiveD
		return in
	}); err != nil {
		return err
	}
	return nil
}

func (ae *awsEnv) reconcileSystemNodeGroup() error {
	systemNodeGroupName := ae.dp.Spec.CloudInfra.Eks.Name + "-system"

	nodeRole, err := ae.eksIC.CreateNodeIamRole(systemNodeGroupName)
	if err != nil {
		return err
	}
	if nodeRole.Role == nil {
		return errors.New("node role is nil")
	}

	systemNodeGroupInput := &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(ae.dp.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String(*nodeRole.Role.Arn),
		NodegroupName:      aws.String(systemNodeGroupName),
		Subnets:            ae.dp.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds,
		AmiType:            "",
		CapacityType:       "",
		ClientRequestToken: nil,
		DiskSize:           nil,
		InstanceTypes:      []string{os.Getenv("AWS_SYSTEM_NODEGROUP_SIZE")},
		Labels: map[string]string{
			"nodeType": "system",
			"name":     systemNodeGroupName,
		},
		LaunchTemplate: nil,
		ReleaseVersion: nil,
		RemoteAccess:   nil,
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(1),
			MaxSize:     aws.Int32(2),
			MinSize:     aws.Int32(1),
		},
		Tags: map[string]string{
			fmt.Sprintf("kubernetes.io/cluster/%s", ae.dp.Spec.CloudInfra.Eks.Name): "owned",
		},
		UpdateConfig: nil,
		Version:      nil,
	}

	describeNodeGroupOutput, found, err := ae.eksIC.DescribeNodegroup(systemNodeGroupName)
	if !found && err == nil {
		if ae.dp.DeletionTimestamp == nil {
			createSystemNodeGroupResult, err := ae.eksIC.CreateSystemNodeGroup(*systemNodeGroupInput)
			if err != nil {
				return err
			}

			if createSystemNodeGroupResult != nil && createSystemNodeGroupResult.Nodegroup != nil {
				klog.Infof("Initated NodeGroup Launch [%s]", *createSystemNodeGroupResult.Nodegroup.ClusterName)
				if err := ae.wrapNgPatchStatus(*createSystemNodeGroupResult.Nodegroup.NodegroupName, string(createSystemNodeGroupResult.Nodegroup.Status)); err != nil {
					return err
				}
			}
		}
	}

	if describeNodeGroupOutput != nil && describeNodeGroupOutput.Nodegroup != nil {
		if err := ae.wrapNgPatchStatus(*describeNodeGroupOutput.Nodegroup.NodegroupName, string(describeNodeGroupOutput.Nodegroup.Status)); err != nil {
			return err
		}
	}

	ae.store.Add(ae.dp.Spec.CloudInfra.Eks.Name, systemNodeGroupName)
	return nil
}

func (ae *awsEnv) wrapNgPatchStatus(name, status string) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
		in := obj.(*v1.DataPlanes)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[name] = status
		return in
	})
	return err
}

func (ae *awsEnv) wrapAddonPatchStatus(addonName, status string) error {
	// update status with current addon status
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
		in := obj.(*v1.DataPlanes)
		if in.Status.AddonStatus == nil {
			in.Status.AddonStatus = make(map[string]string)
		}
		in.Status.AddonStatus[addonName] = status
		return in
	})
	return err
}

func (ae *awsEnv) ReconcileDefaultAddons() error {
	oidcProvider := ae.dp.Status.CloudInfraStatus.AwsCloudInfraConfigStatus.EksStatus.OIDCProviderArn
	if oidcProvider == "" {
		klog.Info("ebs-csi-driver creation: waiting for oidcProvider to be created")
		return nil
	}
	clusterName := ae.dp.Spec.CloudInfra.Eks.Name
	ebsAddon, err := ae.eksIC.DescribeAddon(awsEbsCsiDriver)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			klog.Info("Creating aws-ebs-csi-driver addon")

			role, err := ae.eksIC.CreateEbsCSIRole(ae.ctx)
			if err != nil {
				return err
			}

			_, cErr := ae.eksIC.CreateAddon(ae.ctx, &awseks.CreateAddonInput{
				AddonName:             aws.String(awsEbsCsiDriver),
				ClusterName:           aws.String(clusterName),
				ResolveConflicts:      types.ResolveConflictsOverwrite,
				ServiceAccountRoleArn: role.Role.Arn,
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
		if err := ae.wrapAddonPatchStatus(*addonRes.AddonName, string(addonRes.Status)); err != nil {
			return err
		}
	}

	vpcCniAddon, err := ae.eksIC.DescribeAddon(vpcCni)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			klog.Info("Creating vpc cni addon")
			_, arn, err := ae.eksIC.CreateVpcCniRole(ae.ctx)
			if err != nil {
				return err
			}

			v := `{"enableNetworkPolicy": "true"}`

			_, cErr := ae.eksIC.CreateAddon(ae.ctx, &awseks.CreateAddonInput{
				AddonName:             aws.String(vpcCni),
				ClusterName:           aws.String(clusterName),
				ResolveConflicts:      types.ResolveConflictsOverwrite,
				ServiceAccountRoleArn: aws.String(arn),
				AddonVersion:          aws.String("v1.15.0-eksbuild.2"),
				ConfigurationValues:   aws.String(v),
			})
			if cErr != nil {
				return cErr
			}
			klog.Info("vpc cni addon creation is initiated")
		} else {
			return err
		}
		return nil
	}
	if vpcCniAddon != nil && vpcCniAddon.Addon != nil {
		addonRes := vpcCniAddon.Addon
		klog.Info("vpc cni addon status: ", addonRes.Status)
		if err := ae.wrapAddonPatchStatus(*addonRes.AddonName, string(addonRes.Status)); err != nil {
			return err
		}
	}

	return nil
}
