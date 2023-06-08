package controller

import (
	"context"
	"errors"
	"time"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
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

	eksEnv := eks.NewEksEnvironment(ctx, r.Client, env, *eks.NewConfig(env.Spec.CloudInfra.AwsRegion))

	result, err := eksEnv.DescribeEks()

	if err != nil {

		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {

			klog.Infof("Creating EKS Control plane: %s for Environment: %s/%s", env.Spec.CloudInfra.Eks.Name, env.Namespace, env.Name)
			klog.Info("Updating Environment status to creating")

			createEksResult := eksEnv.CreateEks()
			if createEksResult.Success {
				if _, _, err := utils.PatchStatus(ctx, r.Client, env, func(obj client.Object) client.Object {
					in := obj.(*v1.Environment)
					in.Status.Phase = v1.Creating
					in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
						Type:               v1.ControlPlaneCreateInitiated,
						Status:             corev1.ConditionTrue,
						LastUpdateTime:     metav1.Time{Time: time.Now()},
						LastTransitionTime: metav1.Time{Time: time.Now()},
						Reason:             string(eks.EksControlPlaneInitatedReason),
						Message:            string(eks.EksControlPlaneInitatedMsg),
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

	if result != nil {
		if err := eksEnv.UpdateAwsEksEnvironment(result); err != nil {
			return err
		}
	}

	if result != nil && result.Cluster != nil && result.Cluster.Status == eks.EKSStatusACTIVE {
		if err := eksEnv.ReconcileNodeGroup(r.NgStore); err != nil {
			return err
		}
		if err := eksEnv.ReconcileOIDCProvider(result); err != nil {
			return err
		}
		if err := eksEnv.ReconcileDefaultAddons(); err != nil {
			return err
		}

		if env.Status.AddonStatus[awsEbsCsiDriver] == string(types.AddonStatusDegraded) || env.Status.AddonStatus[awsEbsCsiDriver] == string(types.AddonStatusActive) {
			if err := eksEnv.ReconcileApplicationDeployer(); err != nil {
				return err
			}
		}

		return r.calculatePhase(ctx, env)
	}

	return nil
}

func (r *EnvironmentReconciler) calculatePhase(ctx context.Context, env *v1.Environment) error {
	klog.Info("Calculating Environment Status")

	for status, node := range env.Status.NodegroupStatus {
		if status != string(types.NodegroupStatusActive) {
			klog.Infof("Node %s not active yet", node)
			return nil
		}
	}

	for status, addon := range env.Status.AddonStatus {
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
