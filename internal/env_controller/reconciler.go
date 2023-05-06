package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks/types"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileEnvironment(ctx context.Context, client client.Client, env *v1.Environment, recorder record.EventRecorder) error {
	os.Setenv("AWS_ACCESS_KEY_ID", env.Spec.CloudInfra.Auth.AwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", env.Spec.CloudInfra.Auth.AwsSecretAccessKey)

	return createOrUpdateEnvironment(ctx, env, client, recorder)
}

func createOrUpdateEnvironment(ctx context.Context, env *v1.Environment, c client.Client, record record.EventRecorder) error {
	eksEnv := eks.EksEnvironment{
		Env:    env,
		Config: *eks.NewConfig(env.Spec.CloudInfra.AwsRegion),
		Client: c,
	}

	result, err := eks.DescribeCluster(ctx, eksEnv)
	if err != nil {
		// need to filter out others error except NOT FOUND error
		klog.Infof("Creating EKS Control plane: %s for Environment: %s/%s", eksEnv.Env.Spec.CloudInfra.Eks.Name, eksEnv.Env.Namespace, eksEnv.Env.Name)
		klog.Info("Updating Environment status to creating")
		if _, _, err := PatchStatus(ctx, c, env, func(obj client.Object) client.Object {
			in := obj.(*v1.Environment)
			in.Status.Phase = v1.Creating
			in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
				Type:               v1.ControlPlaneCreateInitiated,
				Status:             corev1.ConditionTrue,
				LastUpdateTime:     metav1.Time{Time: time.Now()},
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Initiated kubernetes control plane",
				Message:            "Initiated kubernetes control plane",
			})
			return in
		}); err != nil {
			return err
		}
		if err := createEnvironment(eksEnv); err != nil {
			return err
		}
		klog.Info("Successfully initiated kubernetes control plane")
		return nil
	}
	if err := updateEnvironment(ctx, eksEnv, result); err != nil {
		return err
	}
	if result.Result.Cluster.Status == eks.EKSStatusACTIVE {
		return reconcileNodeGroup(ctx, eksEnv)
	}
	return nil
}

func updateEnvironment(ctx context.Context, eksEnv eks.EksEnvironment, clusterResult *eks.DescribeClusterOutput) error {
	if clusterResult == nil || clusterResult.Result == nil || clusterResult.Result.Cluster == nil {
		return errors.New("describe cluster output is nil")
	}

	klog.Infof("Syncing Environment: %s/%s", eksEnv.Env.Namespace, eksEnv.Env.Name)

	switch clusterResult.Result.Cluster.Status {
	case eks.EKSStatusCreating:
		klog.Info("Waiting for eks control plane to be created")
		return nil
	case eks.EKSStatusUpdating:
		klog.Info("Waiting for eks control plane to be updated")
		return nil
	case eks.EKSStatusACTIVE:
		return syncControlPlane(ctx, eksEnv, clusterResult)
	}
	return nil
}

func reconcileNodeGroup(ctx context.Context, eksEnv eks.EksEnvironment) error {
	for _, app := range eksEnv.Env.Spec.Application {
		// Todo: Need to figure out how do we know this node group is already created or not
		// One possible option can be Name pattern:
		// like: {eksname}-{appname} in this case if user change the app name, we gonna create
		// another nodegroup for this one
		//
		// another option can be status, we can track created node group in the status in same order of application
		// if user change the application order in the spec, then how we handle that?
		nodeSpec, err := getNodegroupSpecForAppSize(eksEnv.Env, app)
		if err != nil {
			return err
		}

		nodegroup, err := eks.DescribeNodegroup(ctx, &eksEnv, &app)
		if err != nil {
			// Todo: Create only for 404 error
			result, err := eks.CreateNodeGroup(ctx, &eksEnv, nodeSpec, &app)
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println(result)
		}
		if nodegroup != nil && nodegroup.Result != nil && nodegroup.Result.Nodegroup != nil {
			nodegroupName := *nodegroup.Result.Nodegroup.NodegroupName
			switch nodegroup.Result.Nodegroup.Status {
			case types.NodegroupStatusCreating:
				klog.Infof("Waiting for nodegroup %s to be created", nodegroupName)
				return updateStatusWithNodegroup(ctx, &eksEnv, nodegroupName, string(types.NodegroupStatusCreating))
			case types.NodegroupStatusCreateFailed:
				klog.Errorf("Create failed for nodegroup %s", nodegroupName)
				return updateStatusWithNodegroup(ctx, &eksEnv, nodegroupName, string(types.NodegroupStatusCreateFailed))
			case types.NodegroupStatusActive:
				if err := updateStatusWithNodegroup(ctx, &eksEnv, nodegroupName, string(types.NodegroupStatusActive)); err != nil {
					return err
				}
				return syncNodegroup(ctx, &eksEnv)
			}
		} else {
			klog.Error("Bad formatted result for node group")
		}
	}
	return nil
}

func updateStatusWithNodegroup(ctx context.Context, eksEnv *eks.EksEnvironment, nodegroup, status string) error {
	// update status with current nodegroup status
	_, _, err := PatchStatus(ctx, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
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

func syncControlPlane(ctx context.Context, eksEnv eks.EksEnvironment, clusterResult *eks.DescribeClusterOutput) error {
	// checking for version upgrade
	statusVersion := eksEnv.Env.Status.Version
	specVersion := eksEnv.Env.Spec.CloudInfra.Eks.Version
	if statusVersion != "" && statusVersion != specVersion && *clusterResult.Result.Cluster.Version != specVersion {
		klog.Info("Updating Kubernetes version to: ", eksEnv.Env.Spec.CloudInfra.Eks.Version)
		if _, _, err := PatchStatus(ctx, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
			in := obj.(*v1.Environment)
			in.Status.Phase = v1.Updating
			in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
				Type:               v1.VersionUpgradeInitiated,
				Status:             corev1.ConditionTrue,
				LastUpdateTime:     metav1.Time{Time: time.Now()},
				LastTransitionTime: metav1.Time{Time: time.Now()},
				Reason:             "Kubernetes control plane version upgrade initiated",
				Message:            "Kubernetes control plane version upgrade initiated",
			})
			return in
		}); err != nil {
			return err
		}
		result := eksEnv.UpdateEks()
		if !result.Success {
			return errors.New(result.Result)
		}
		klog.Info("Successfully initiated version update")
	}

	klog.Info("Sync Cluster status and version")
	if _, _, err := PatchStatus(ctx, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		in.Status.Phase = v1.Success
		in.Status.Version = in.Spec.CloudInfra.Eks.Version
		in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
			Type:               v1.ControlPlaneCreated,
			Status:             corev1.ConditionTrue,
			LastUpdateTime:     metav1.Time{Time: time.Now()},
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Reason:             "Kubernetes control plane created",
			Message:            "Kubernetes control plane created",
		})
		return in
	}); err != nil {
		return err
	}
	return nil
}

func createEnvironment(eksEnv eks.EksEnvironment) error {
	output := eksEnv.CreateEks()
	if output.Success {
		return nil
	}
	return errors.New(output.Result)
}
