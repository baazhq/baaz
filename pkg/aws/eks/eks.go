package eks

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"k8s.io/klog/v2"
)

type EnvironmentControllerReason string

const (
	EksControlPlaneCreationInitatedReason EnvironmentControllerReason = "EksControlPlaneCreationInitated"
	EksControlPlaneCreatedReason          EnvironmentControllerReason = "EksControlPlaneCreated	"
	EksControlPlaneCreationUpgradedReason EnvironmentControllerReason = "EksControlUpgradeInitatedReason"
	EksControlPlaneUpgradedReason         EnvironmentControllerReason = "EksControlPlaneUpgradeReason"
)

type EnvironmentControllerMsg string

const (
	EksControlPlaneCreationInitatedMsg EnvironmentControllerMsg = "Initiated creation eks kubernetes control plane"
	EksControlPlaneCreatedMsg          EnvironmentControllerMsg = "Created eks kubernetes control plane"
	EksControlPlaneUpgradedIntiatedMsg EnvironmentControllerMsg = "Initaled upgrade eks kubernetes control plane"
	EksControlPlaneUpgradedMsg         EnvironmentControllerMsg = "Upgraded eks kubernetes control plane"
)

type EksInternalOutput struct {
	Result     string
	Properties map[string]string
	Success    bool
}

func (ec *eks) DescribeEks() (*awseks.DescribeClusterOutput, error) {

	result, err := ec.awsClient.DescribeCluster(
		ec.ctx,
		&awseks.DescribeClusterInput{
			Name: aws.String(ec.environment.Spec.CloudInfra.Eks.Name),
		},
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ec *eks) CreateEks() *EksInternalOutput {
	if err := ec.createEks(); err != nil {
		return &EksInternalOutput{Result: err.Error()}
	}

	return &EksInternalOutput{Result: string(EksControlPlaneCreationInitatedMsg), Success: true}
}

func (ec *eks) createEks() error {

	roleName, err := ec.CreateClusterIamRole()
	if err != nil {
		return err
	}
	_, err = ec.awsClient.CreateCluster(ec.ctx, &awseks.CreateClusterInput{
		Name: &ec.environment.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: ec.environment.Spec.CloudInfra.Eks.SubnetIds,
		},
		RoleArn: roleName.Role.Arn,
		Version: aws.String(ec.environment.Spec.CloudInfra.Eks.Version),
	})

	return err
}

func (ec *eks) UpdateEks() *EksInternalOutput {

	err := ec.updateEks()
	if err != nil {
		return &EksInternalOutput{Result: err.Error()}
	}

	return &EksInternalOutput{
		Result:  string(EksControlPlaneCreationInitatedReason),
		Success: true,
	}
}

func (ec *eks) updateEks() error {

	_, err := ec.awsClient.UpdateClusterVersion(ec.ctx, &awseks.UpdateClusterVersionInput{
		Name:    &ec.environment.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
		Version: aws.String(ec.environment.Spec.CloudInfra.Eks.Version),
	})

	if err != nil {
		return err
	}

	return nil
}

func (ec *eks) UpdateAwsEksEnvironment(clusterResult *awseks.DescribeClusterOutput) types.ClusterStatus {

	klog.Infof("Syncing Environment: %s/%s", ec.environment.Namespace, ec.environment.Name)

	switch clusterResult.Cluster.Status {

	case types.ClusterStatusCreating:
		return types.ClusterStatusCreating
	case types.ClusterStatusUpdating:
		return types.ClusterStatusUpdating
	case types.ClusterStatusActive:
		return types.ClusterStatusActive
	}
	return clusterResult.Cluster.Status
}

func (ec *eks) DeleteEKS() (*awseks.DeleteClusterOutput, error) {
	klog.Infof("Deleting EKS Control Plane [%s]", ec.environment.Spec.CloudInfra.Eks.Name)
	return ec.awsClient.DeleteCluster(ec.ctx, &awseks.DeleteClusterInput{
		Name: &ec.environment.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
	})
}
