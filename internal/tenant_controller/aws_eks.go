package tenant_controller

import (
	"context"
	"errors"
	"fmt"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"datainfra.io/ballastdata/pkg/store"
	"datainfra.io/ballastdata/pkg/utils"
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
	ctx    context.Context
	env    *v1.Environment
	tenant *v1.Tenants
	eksIC  eks.Eks
	client client.Client
	store  store.Store
}

func (ae *awsEnv) ReconcileTenants() error {
	klog.Info("Reconciling node groups")

	var nodeName string
	for _, tenantConfig := range ae.tenant.Spec.TenantConfig {

		ngNameNgSpec, err := ae.getNodeSpecForTenantSize(tenantConfig)
		if err != nil {
			return err
		}

		for _, nodeSpec := range *ngNameNgSpec {
			nodeName = makeNodeName(nodeSpec.Name, string(tenantConfig.AppType), tenantConfig.Size)

			if ae.env.Status.NodegroupStatus[string(nodeSpec.Name)] != "DELETING" {

				describeNodegroupOutput, found, _ := ae.eksIC.DescribeNodegroup(nodeName)
				if !found {
					nodeRole, err := ae.eksIC.CreateNodeIamRole(nodeName)
					if err != nil {
						return err
					}
					if nodeRole.Role == nil {
						return errors.New("node role is nil")
					}

					createNodeGroupOutput, err := ae.eksIC.CreateNodegroup(ae.getNodegroupInput(nodeName, *nodeRole.Role.Arn, &nodeSpec))
					if err != nil {
						return err
					}
					if createNodeGroupOutput != nil && createNodeGroupOutput.Nodegroup != nil {
						klog.Infof("Initated NodeGroup Launch [%s]", *createNodeGroupOutput.Nodegroup.ClusterName)
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
		if ae.env != nil {
			ae.store.Add(ae.env.Spec.CloudInfra.Eks.Name, nodeName)
		}
	}

	return nil
}

func makeNodeName(nodeName, appType, size string) string {
	return nodeName + "-" + appType + "-" + size
}

func (ae *awsEnv) getNodeSpecForTenantSize(tenantConfig v1.TenantConfig) (*[]v1.NodeSpec, error) {
	for _, size := range ae.tenant.Spec.TenantSizes {
		if size.Name == tenantConfig.Size {
			return &size.Spec, nil
		}
	}
	return nil, fmt.Errorf("no NodegroupSpec for app %s & size %s", tenantConfig.AppType, tenantConfig.Size)
}

func (ae *awsEnv) getNodegroupInput(nodeName, roleArn string, nodeSpec *v1.NodeSpec) (input *awseks.CreateNodegroupInput) {

	return &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(ae.env.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String(roleArn),
		NodegroupName:      aws.String(nodeName),
		Subnets:            ae.env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds,
		AmiType:            "",
		CapacityType:       "",
		ClientRequestToken: nil,
		DiskSize:           nil,
		InstanceTypes:      []string{nodeSpec.Size},
		Labels:             nodeSpec.NodeLabels,
		LaunchTemplate:     nil,
		ReleaseVersion:     nil,
		RemoteAccess:       nil,
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(nodeSpec.Min),
			MaxSize:     aws.Int32(nodeSpec.Max),
			MinSize:     aws.Int32(nodeSpec.Min),
		},
		Tags: map[string]string{
			fmt.Sprintf("kubernetes.io/cluster/%s", ae.env.Spec.CloudInfra.Eks.Name): "owned",
		},
		Taints:       []types.Taint{},
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
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.tenant, func(obj client.Object) client.Object {
		in := obj.(*v1.Tenants)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[name] = status
		return in
	})
	return err
}
