package tenantinfra_controller

import (
	"context"
	"errors"
	"fmt"

	v1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/pkg/aws/eks"
	"datainfra.io/baaz/pkg/store"
	"datainfra.io/baaz/pkg/utils"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go/aws"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type nodeGroupType string

const (
	app    nodeGroupType = "application"
	system nodeGroupType = "system"
)

type awsEnv struct {
	ctx          context.Context
	dp           *v1.DataPlanes
	tenantsInfra *v1.TenantsInfra
	eksIC        eks.Eks
	client       client.Client
	store        store.Store
}

func (ae *awsEnv) ReconcileInfraTenants() error {
	klog.Info("Reconciling tenant infra node groups")

	for _, tenantSize := range ae.tenantsInfra.Spec.TenantSizes {

		for _, machineSpec := range tenantSize.MachineSpec {
			if ae.dp.Status.NodegroupStatus[tenantSize.Name] != "DELETING" {
				nodeName := tenantSize.Name + "-" + machineSpec.Name

				describeNodegroupOutput, found, _ := ae.eksIC.DescribeNodegroup(nodeName)
				if !found {
					nodeRole, err := ae.eksIC.CreateNodeIamRole(nodeName)
					if err != nil {
						return err
					}
					if nodeRole.Role == nil {
						return errors.New("node role is nil")
					}

					createNodeGroupOutput, err := ae.eksIC.CreateNodegroup(ae.getNodegroupInput(nodeName, *nodeRole.Role.Arn, &machineSpec))
					if err != nil {
						return err
					}
					if createNodeGroupOutput != nil && createNodeGroupOutput.Nodegroup != nil {
						klog.Infof("Initated NodeGroup Launch [%s]", *createNodeGroupOutput.Nodegroup.NodegroupName)
						if err := ae.patchStatus(*createNodeGroupOutput.Nodegroup.NodegroupName, string(createNodeGroupOutput.Nodegroup.Status)); err != nil {
							return err
						}
					}

				}

				if describeNodegroupOutput != nil && describeNodegroupOutput.Nodegroup != nil {
					if err := ae.patchStatus(*describeNodegroupOutput.Nodegroup.NodegroupName, string(describeNodegroupOutput.Nodegroup.Status)); err != nil {
						return err
					}
				}
			}
		}

	}

	return nil
}

func makeNodeName(nodeName, appType, size string) string {
	return nodeName + "-" + appType + "-" + size
}

// func (ae *awsEnv) getNodeSpecForTenantSize(tenantConfig v1.TenantApplicationConfig) (*[]v1.MachineSpec, error) {

// 	// cm := corev1.ConfigMap{}
// 	// if err := ae.client.Get(
// 	// 	ae.ctx,
// 	// 	k8stypes.NamespacedName{Name: "tenant-sizes", Namespace: "kube-system"},
// 	// 	&cm,
// 	// ); err != nil {
// 	// 	return nil, err
// 	// }
// 	// sizeJson := cm.Data["size.json"]

// 	// var tenantInfraAppSize v1.TenantInfraAppSize

// 	// err := json.Unmarshal([]byte(sizeJson), &tenantInfraAppSize)

// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	for _, size := range tenantInfraAppSize.TenantSizes {
// 		if size.Name == tenantConfig.Size {
// 			return &size.MachineSpec, nil
// 		}
// 	}

// 	return nil, fmt.Errorf("no NodegroupSpec for app %s & size %s", tenantConfig.AppType, tenantConfig.Size)
// }

func (ae *awsEnv) getNodegroupInput(nodeName, roleArn string, machineSpec *v1.MachineSpec) (input *awseks.CreateNodegroupInput) {

	var taints = &[]types.Taint{}

	taints = makeTaints(nodeName)

	return &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(ae.dp.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String(roleArn),
		NodegroupName:      aws.String(nodeName),
		Subnets:            ae.dp.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds,
		AmiType:            "",
		CapacityType:       "",
		ClientRequestToken: nil,
		DiskSize:           nil,
		InstanceTypes:      []string{machineSpec.Size},
		Labels:             machineSpec.NodeLabels,
		LaunchTemplate:     nil,
		ReleaseVersion:     nil,
		RemoteAccess:       nil,
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(machineSpec.Min),
			MaxSize:     aws.Int32(machineSpec.Max),
			MinSize:     aws.Int32(machineSpec.Min),
		},
		Tags: map[string]string{
			fmt.Sprintf("kubernetes.io/cluster/%s", ae.dp.Spec.CloudInfra.Eks.Name): "owned",
		},
		Taints:       *taints,
		UpdateConfig: nil,
		Version:      nil,
	}

}

func makeTaints(value string) *[]types.Taint {
	return &[]types.Taint{
		{
			Effect: types.TaintEffectNoSchedule,
			Key:    aws.String("application"),
			Value:  aws.String(value),
		},
	}
}

func (ae *awsEnv) patchStatus(name, status string) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.tenantsInfra, func(obj client.Object) client.Object {
		in := obj.(*v1.TenantsInfra)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[name] = status
		in.Status.Phase = v1.TenantPhase(status)
		return in
	})
	return err
}
