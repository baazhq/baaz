package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"datainfra.io/ballastdata/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func createOrUpdateAwsEksEnvironment(ctx context.Context, env *v1.Environment, c client.Client, record record.EventRecorder) error {

	// init eks environment
	eksEnv := eks.NewEksEnvironment(ctx, c, env, *eks.NewConfig(env.Spec.CloudInfra.AwsRegion))

	result, err := eksEnv.DescribeEks()

	if err != nil {

		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {

			klog.Infof("Creating EKS Control plane: %s for Environment: %s/%s", env.Spec.CloudInfra.Eks.Name, env.Namespace, env.Name)
			klog.Info("Updating Environment status to creating")

			createEksResult := eksEnv.CreateEks()
			if createEksResult.Success {
				if _, _, err := utils.PatchStatus(ctx, c, env, func(obj client.Object) client.Object {
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

			}

		}
	}

	if result != nil {
		if err := eksEnv.UpdateAwsEksEnvironment(result); err != nil {
			return err
		}
	}
	if result.Result.Cluster.Status == eks.EKSStatusACTIVE {
		if err := reconcileNodeGroup(ctx, eksEnv); err != nil {
			return err
		}
		return reconcileDefaultAddons(ctx, eksEnv)
	}
	return nil
}

func reconcileDefaultAddons(ctx context.Context, eksEnv *eks.EksEnvironment) error {
	ebsAddon, err := eksEnv.DescribeAddon(ctx, "aws-ebs-csi-driver", eksEnv.Env.Spec.CloudInfra.Eks.Name)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			// create addon
			_, cErr := eksEnv.CreateAddon(ctx, &eks.CreateAddonInput{
				Name:        "aws-ebs-csi-driver",
				ClusterName: eksEnv.Env.Spec.CloudInfra.Eks.Name,
				RoleArn:     "arn:aws:iam::437639712640:role/ebs-sa-role",
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
	if ebsAddon.Result != nil && ebsAddon.Result.Addon != nil {
		klog.Info("aws-ebs-csi-driver addon status: ", ebsAddon.Result.Addon.Status)
	}
	return nil
}

func reconcileNodeGroup(ctx context.Context, env *eks.EksEnvironment) error {

	for _, app := range env.Env.Spec.Application {

		nodeSpec, err := getNodegroupSpecForAppSize(env.Env, app)
		if err != nil {
			return err
		}

		ngs := eks.NewNodeGroup(ctx, env, &app, nodeSpec)

		_, err = ngs.CreateNodeGroupForApp()
		if err != nil {
			return err
		}

	}

	return nil
}

func updateStatusWithNodegroup(ctx context.Context, eksEnv *eks.EksEnvironment, nodegroup, status string) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ctx, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[nodegroup] = status
		return in
	})
	return err
}

func syncNodegroup(ctx context.Context, eksEnv *eks.EksEnvironment) error {
	// update node group if spec node group is updated
	return nil
}

func getNodegroupSpecForAppSize(env *v1.Environment, app v1.ApplicationConfig) (*v1.NodeGroupSpec, error) {
	for _, size := range env.Spec.Size {
		if size.Name == app.Size && size.Spec.AppType == app.AppType {
			return size.Spec.Nodes, nil
		}
	}
	return nil, fmt.Errorf("no NodegroupSpec for app %s & size %s", app.Name, app.Size)
}
