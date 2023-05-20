package eks

import (
	"context"
	"errors"
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/utils"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

type EnvironmentControllerReason string

const (
	EksControlPlaneInitatedReason EnvironmentControllerReason = "EksControlPlaneInitated"
	EksControlPlaneCreatedReason  EnvironmentControllerReason = "EksControlPlaneCreated"
)

type EnvironmentControllerMsg string

const (
	EksControlPlaneInitatedMsg EnvironmentControllerReason = "Initiated eks kubernetes control plane"
	EksControlPlaneCreatedMsg  EnvironmentControllerReason = "Created eks kubernetes control plane"
)

const (
	EKSStatusCreating = "CREATING"
	EKSStatusACTIVE   = "ACTIVE"
	EKSStatusUpdating = "UPDATING"
)

type EksOutput struct {
	Result     string
	Properties map[string]string
	Success    bool
}

type DescribeClusterOutput struct {
	Result *awseks.DescribeClusterOutput `json:"result"`
}

type EksEnvironment struct {
	Context context.Context
	Client  client.Client
	Env     *v1.Environment
	Config  aws.Config
}

func NewEksEnvironment(
	ctx context.Context,
	client client.Client,
	env *v1.Environment,
	config aws.Config,
) *EksEnvironment {
	return &EksEnvironment{
		Context: ctx,
		Client:  client,
		Env:     env,
		Config:  config,
	}
}

func NewConfig(awsRegion string) *aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return &cfg
}

func (eksEnv *EksEnvironment) CreateEks() EksOutput {
	if err := eksEnv.createEks(); err != nil {
		return EksOutput{Result: err.Error()}
	}

	return EksOutput{Result: string(EksControlPlaneInitatedMsg), Success: true}
}

func (eksEnv *EksEnvironment) createEks() error {
	eksClient := eks.NewFromConfig(eksEnv.Config)

	roleName, err := eksEnv.createClusterIamRole()
	if err != nil {
		return err
	}
	_, err = eksClient.CreateCluster(eksEnv.Context, &eks.CreateClusterInput{
		Name: &eksEnv.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: eksEnv.Env.Spec.CloudInfra.Eks.SubnetIds,
		},
		RoleArn: roleName.Role.Arn,
		Version: aws.String(eksEnv.Env.Spec.CloudInfra.Eks.Version),
	})

	return err
}

func (eksEnv *EksEnvironment) UpdateEks() EksOutput {

	errChannel := make(chan error)

	go eksEnv.updateEks(errChannel)

	for err := range errChannel {
		if err != nil {
			return EksOutput{Result: err.Error()}
		}
		break
	}
	return EksOutput{
		Result:  string(EksControlPlaneInitatedMsg),
		Success: true,
	}
}

func (eksEnv *EksEnvironment) updateEks(errorChan chan<- error) {
	eksClient := eks.NewFromConfig(eksEnv.Config)

	_, err := eksClient.UpdateClusterVersion(eksEnv.Context, &eks.UpdateClusterVersionInput{
		Name:    &eksEnv.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
		Version: aws.String(eksEnv.Env.Spec.CloudInfra.Eks.Version),
	})

	if err != nil {
		errorChan <- err
	}
}

func (eksEnv *EksEnvironment) DescribeEks() (*DescribeClusterOutput, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)

	result, err := eksClient.DescribeCluster(eksEnv.Context, &awseks.DescribeClusterInput{Name: aws.String(eksEnv.Env.Spec.CloudInfra.Eks.Name)})
	if err != nil {
		return nil, err
	}
	return &DescribeClusterOutput{Result: result}, nil
}

func (eksEnv *EksEnvironment) UpdateAwsEksEnvironment(clusterResult *DescribeClusterOutput) error {

	klog.Infof("Syncing Environment: %s/%s", eksEnv.Env.Namespace, eksEnv.Env.Name)

	switch clusterResult.Result.Cluster.Status {

	case EKSStatusCreating:
		klog.Info("Waiting for eks control plane to be created")
		return nil
	case EKSStatusUpdating:
		klog.Info("Waiting for eks control plane to be updated")
		return nil
	case EKSStatusACTIVE:
		return eksEnv.syncEksControlPlane(clusterResult)
	}

	return nil
}

func (eksEnv *EksEnvironment) syncEksControlPlane(clusterResult *DescribeClusterOutput) error {
	// checking for version upgrade
	statusVersion := eksEnv.Env.Status.Version
	specVersion := eksEnv.Env.Spec.CloudInfra.Eks.Version
	if statusVersion != "" && statusVersion != specVersion && *clusterResult.Result.Cluster.Version != specVersion {
		klog.Info("Updating Kubernetes version to: ", eksEnv.Env.Spec.CloudInfra.Eks.Version)
		if _, _, err := utils.PatchStatus(eksEnv.Context, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
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

	if _, _, err := utils.PatchStatus(eksEnv.Context, eksEnv.Client, eksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		in.Status.Phase = v1.Success
		in.Status.Version = in.Spec.CloudInfra.Eks.Version
		in.Status.Conditions = in.AddCondition(v1.EnvironmentCondition{
			Type:               v1.ControlPlaneCreated,
			Status:             corev1.ConditionTrue,
			LastUpdateTime:     metav1.Time{Time: time.Now()},
			LastTransitionTime: metav1.Time{Time: time.Now()},
			Reason:             string(EksControlPlaneCreatedReason),
			Message:            string(EksControlPlaneCreatedMsg),
		})
		return in
	}); err != nil {
		return err
	}
	return nil
}

type DescribeAddonOutput struct {
	Result *eks.DescribeAddonOutput `json:"result"`
}

func (eksEnv *EksEnvironment) DescribeAddon(ctx context.Context, addonName, clusterName string) (*DescribeAddonOutput, error) {
	eksClient := awseks.NewFromConfig(eksEnv.Config)

	input := &awseks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}
	result, err := eksClient.DescribeAddon(ctx, input)
	if err != nil {
		return nil, err
	}
	return &DescribeAddonOutput{Result: result}, nil
}

type CreateAddonOutput struct {
	Result *eks.CreateAddonOutput `json:"result"`
}

type CreateAddonInput struct {
	Name        string `json:"name"`
	ClusterName string `json:"clusterName"`
}

func (eksEnv *EksEnvironment) CreateAddon(ctx context.Context, params *CreateAddonInput) (*CreateAddonOutput, error) {
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
	return &CreateAddonOutput{Result: result}, nil
}
