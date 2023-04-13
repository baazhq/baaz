package eks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

func (ke *EksEnvironment) CreateEks() EksOutput {

	errChannel := make(chan error)

	go ke.createEks(errChannel)

	for err := range errChannel {
		if err != nil {
			return EksOutput{Result: err.Error()}
		}
	}
	return EksOutput{Result: ClusterLaunchInitated}

	//	updatedStatus := v1.EnvironmentStatus{}

	// updatedStatus.CloudInfraStatus = v1.CloudInfraStatus{
	// 	Type: "aws",
	// 	AwsCloudInfraConfigStatus: v1.AwsCloudInfraConfigStatus{
	// 		EksStatus: v1.EksStatus{
	// 			ClusterId: *clusterOutput.Cluster.Id,
	// 		},
	// 	},
	// }

	// patchBytes, err := json.Marshal(map[string]v1.EnvironmentStatus{"status": updatedStatus})
	// if err != nil {
	// 	return KubernetesOutput{Result: "patch fail"}, err
	// }

	// if err := ke.Client.Status().Patch(
	// 	context.Background(),
	// 	ke.Env,
	// 	client.RawPatch(k8stypes.MergePatchType,
	// 		patchBytes,
	// 	)); err != nil {
	// 	return KubernetesOutput{Result: "patch fail"}, err
	// }

}

func (ke *EksEnvironment) createEks(errorChan chan<- error) {
	eksClient := eks.NewFromConfig(ke.Config)

	_, err := eksClient.CreateCluster(context.TODO(), &eks.CreateClusterInput{
		Name: &ke.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds:        ke.Env.Spec.CloudInfra.Eks.SubnetIds,
			SecurityGroupIds: ke.Env.Spec.CloudInfra.Eks.SecurityGroupIds,
		},
		RoleArn: aws.String(ke.Env.Spec.CloudInfra.Eks.RoleArn),
		Version: aws.String(ke.Env.Spec.CloudInfra.Eks.Version),
	})
	errorChan <- err
}

// func (ke *EksEnvironment) createEksNodegroup(errorChan chan<- error) {
// 	eksClient := eks.NewFromConfig(ke.Config)

// 	_, err := eksClient.CreateNodegroup(context.TODO(), &eks.CreateNodegroupInput{
// 		ClusterName: &ke.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.Name,
// 		ResourcesVpcConfig: &types.VpcConfigRequest{
// 			SubnetIds:        ke.Env.Spec.CloudInfra.Eks.SubnetIds,
// 			SecurityGroupIds: ke.Env.Spec.CloudInfra.Eks.SecurityGroupIds,
// 		},
// 		RoleArn: aws.String(ke.Env.Spec.CloudInfra.Eks.RoleArn),
// 		Version: aws.String(ke.Env.Spec.CloudInfra.Eks.Version),
// 	})
// 	errorChan <- err
// }
