package controller

import (
	"context"
	"errors"
	"os"
	"time"

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
	return updateEnvironment(ctx, eksEnv, result)
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
