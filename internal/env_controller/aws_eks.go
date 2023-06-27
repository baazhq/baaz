package controller

import (
	"context"
	"errors"
	"time"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"datainfra.io/ballastdata/pkg/deployer/applications"
	"datainfra.io/ballastdata/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var envFinalizer = "environment.datainfra.io/finalizer"

const (
	awsEbsCsiDriver string = "aws-ebs-csi-driver"
)

func (r *EnvironmentReconciler) createOrUpdateAwsEksEnvironment(ctx context.Context, env *v1.Environment) error {

	eksClient := eks.NewEks(ctx, env)
	eksDescribeClusterOutput, err := eksClient.DescribeEks()

	if err != nil {

		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {

			klog.Infof("Creating EKS Control plane: %s for Environment: %s/%s", env.Spec.CloudInfra.Eks.Name, env.Namespace, env.Name)
			klog.Info("Updating Environment status to creating")

			createEksResult := eksClient.CreateEks()
			if createEksResult.Success {
				if _, _, err := utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
					in := obj.(*v1.Environment)
					in.Status.Phase = v1.Creating
					in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
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

		}
	}

	if eksDescribeClusterOutput != nil {
		if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusActive {
			// checking for version upgrade
			statusVersion := env.Status.Version
			specVersion := env.Spec.CloudInfra.Eks.Version
			if statusVersion != "" && statusVersion != specVersion && *eksDescribeClusterOutput.Cluster.Version != specVersion {
				klog.Info("Updating Kubernetes version to: ", env.Spec.CloudInfra.Eks.Version)
				if _, _, err := utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
					in := obj.(*v1.Environment)
					in.Status.Phase = v1.Updating
					in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
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
				result := eksClient.UpdateEks()
				if !result.Success {
					return errors.New(result.Result)
				}
				klog.Info("Successfully initiated version update")
			}

			klog.Info("Sync Cluster status and version")

			if _, _, err := utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
				in := obj.(*v1.Environment)
				in.Status.Phase = v1.Success
				in.Status.Version = in.Spec.CloudInfra.Eks.Version
				in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
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
			klog.Infof("EKS Cluster Control Plane [%s] in creating state", env.Spec.CloudInfra.Eks.Name)
		} else if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusUpdating {
			klog.Infof("EKS Cluster Control Plane [%s] in updated state", env.Spec.CloudInfra.Eks.Name)
		} else if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusDeleting {
			klog.Infof("EKS Cluster Control Plane [%s] in deleting state", env.Spec.CloudInfra.Eks.Name)
		}

	}

	if eksDescribeClusterOutput != nil && eksDescribeClusterOutput.Cluster != nil && eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusActive {

		calico := applications.NewCali(
			env.Spec.CloudInfra.Eks.Name,
			v1.CloudType(env.Spec.CloudInfra.Type),
		)

		if err := calico.ReconcileCalico(); err != nil {
			return err
		}

		oidcOutput, err := eksClient.ReconcileOIDCProvider(eksDescribeClusterOutput)
		if err != nil {
			return err
		}

		if oidcOutput != nil && oidcOutput.OpenIDConnectProviderArn != nil {
			_, _, err = utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
				in := obj.(*v1.Environment)
				in.Status.CloudInfraStatus.EksStatus.OIDCProviderArn = *oidcOutput.OpenIDConnectProviderArn
				return in
			})
		}

		// if err := eksEnv.ReconcileDefaultAddons(); err != nil {
		// 	return err
		// }

		return r.calculatePhase(ctx, env)
	}

	return nil
}

func (r *EnvironmentReconciler) calculatePhase(ctx context.Context, env *v1.Environment) error {
	klog.Info("Calculating Environment Status")

	for node, status := range env.Status.NodegroupStatus {
		if status != string(types.NodegroupStatusActive) {
			klog.Infof("Node %s not active yet", node)
			return nil
		}
	}

	for addon, status := range env.Status.AddonStatus {
		if status != string(types.AddonStatusActive) {
			klog.Infof("Addon %s not active yet", addon)
			return nil
		}
	}

	if _, _, err := utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		in.Status.Phase = v1.Success
		return in
	}); err != nil {
		return err
	}
	return nil
}
