package eks

import (
	"errors"
	"fmt"

	"datainfra.io/ballastdata/pkg/store"
	"datainfra.io/ballastdata/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "datainfra.io/ballastdata/api/v1"
	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"k8s.io/klog/v2"
)

type nodeGroupType string

const (
	app    nodeGroupType = "application"
	system nodeGroupType = "system"
)

type nodeGroup struct {
	EksEnv        *EksEnvironment
	TenantConfig  *v1.TenantConfig
	TenantName    string
	NodeGroupName v1.NodeGroupName
	NodeGroupSpec *v1.NodeGroupSpec
}

func newNodeGroup(
	eksEnv *EksEnvironment,
	tenantConfig *v1.TenantConfig,
	nodeGroupSpec *v1.NodeGroupSpec,
	nodeGroupName v1.NodeGroupName,
	tenantName string,
) *nodeGroup {
	ngs := &nodeGroup{
		EksEnv:        eksEnv,
		TenantConfig:  tenantConfig,
		NodeGroupSpec: nodeGroupSpec,
		NodeGroupName: nodeGroupName,
		TenantName:    tenantName,
	}
	return ngs
}

func (eksEnv *EksEnvironment) ReconcileNodeGroup(store store.Store) error {
	klog.Info("Reconciling node groups")

	for tenantName, tenantConfigs := range eksEnv.Env.Spec.Tenant {

		for _, tenantConfig := range tenantConfigs {
			ngNameNgSpec, err := getNodegroupSpecForTenantSize(eksEnv.Env, tenantName, tenantConfig)
			if err != nil {
				return err
			}

			for ngName, ngNodeSpec := range ngNameNgSpec {

				if eksEnv.Env.Status.NodegroupStatus[string(ngName)] != "DELETING" {

					ngs := newNodeGroup(eksEnv, &tenantConfig, ngNodeSpec, ngName, tenantName)

					_, err = ngs.createNodeGroupForTenant(store)
					if err != nil {
						return err
					}
				}
			}
		}

	}

	return nil
}

func (ng *nodeGroup) createNodeGroupForTenant(store store.Store) (*awseks.CreateNodegroupOutput, error) {

	switch ng.TenantConfig.AppType {

	case v1.Druid:

		systemNgName := *aws.String(makeSystemNodeGroupName(ng.TenantName))
		druidNodeNgName := *aws.String(makeTenantNodeGroupName(ng.TenantName, v1.Druid, ng.TenantConfig.Size, ng.NodeGroupName))

		// system nodepool
		_, err := ng.createOrUpdateNodeGroup(systemNgName, system, store)
		if err != nil {
			return nil, err
		}

		// druid nodepool
		_, err = ng.createOrUpdateNodeGroup(druidNodeNgName, app, store)
		if err != nil {
			return nil, err
		}

	case v1.Zookeeper:

		zkNodeNgName := *aws.String(makeTenantNodeGroupName(ng.TenantName, v1.Zookeeper, ng.TenantConfig.Size, ng.NodeGroupName))
		_, err := ng.createOrUpdateNodeGroup(zkNodeNgName, app, store)
		if err != nil {
			return nil, err
		}

	}

	return &awseks.CreateNodegroupOutput{}, nil
}

func (ng *nodeGroup) getNodeGroupInput(name string, ngType nodeGroupType) (*awseks.CreateNodegroupInput, error) {

	var taints = []types.Taint{}

	if ngType == app {
		taints = *makeTaints(name)
	}

	nodeRole, err := ng.EksEnv.createNodeIamRole(name)
	if err != nil {
		return nil, err
	}

	if nodeRole.Role == nil {
		return nil, errors.New("node role is nil")
	}

	return &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(ng.EksEnv.Env.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String(*nodeRole.Role.Arn),
		NodegroupName:      aws.String(name),
		Subnets:            ng.EksEnv.Env.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds,
		AmiType:            "",
		CapacityType:       "",
		ClientRequestToken: nil,
		DiskSize:           nil,
		InstanceTypes:      []string{ng.NodeGroupSpec.NodeSize},
		Labels:             ng.NodeGroupSpec.NodeLabels,
		LaunchTemplate:     nil,
		ReleaseVersion:     nil,
		RemoteAccess:       nil,
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(ng.NodeGroupSpec.Min),
			MaxSize:     aws.Int32(ng.NodeGroupSpec.Max),
			MinSize:     aws.Int32(ng.NodeGroupSpec.Min),
		},
		Tags: map[string]string{
			fmt.Sprintf("kubernetes.io/cluster/%s", ng.EksEnv.Env.Spec.CloudInfra.Eks.Name): "owned",
		},
		Taints:       taints,
		UpdateConfig: nil,
		Version:      nil,
	}, nil
}

func (ng *nodeGroup) describeNodegroup(name string) (*awseks.DescribeNodegroupOutput, error) {
	eksClient := awseks.NewFromConfig(ng.EksEnv.Config)
	input := &awseks.DescribeNodegroupInput{
		ClusterName:   aws.String(ng.EksEnv.Env.Spec.CloudInfra.Eks.Name),
		NodegroupName: aws.String(name),
	}

	result, err := eksClient.DescribeNodegroup(ng.EksEnv.Context, input)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ng *nodeGroup) patchStatus(name, status string) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ng.EksEnv.Context, ng.EksEnv.Client, ng.EksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[name] = status
		return in
	})
	return err
}

func (ng *nodeGroup) createOrUpdateNodeGroup(nodeGroupName string, ngType nodeGroupType, store store.Store) (*awseks.CreateNodegroupOutput, error) {
	eksClient := awseks.NewFromConfig(ng.EksEnv.Config)

	describeRes, err := ng.describeNodegroup(nodeGroupName)
	if err != nil {
		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {

			if ng.EksEnv.Env.DeletionTimestamp == nil {
				nodeGroup, err := ng.getNodeGroupInput(nodeGroupName, ngType)
				if err != nil {
					return nil, err
				}
				result, err := eksClient.CreateNodegroup(ng.EksEnv.Context, nodeGroup)
				if err != nil {
					return nil, err
				}

				if result != nil && result.Nodegroup != nil {
					klog.Infof("Initated NodeGroup Launch [%s]", *result.Nodegroup.ClusterName)
					if err := ng.patchStatus(*result.Nodegroup.NodegroupName, string(result.Nodegroup.Status)); err != nil {
						return nil, err
					}
				}
			}
		}
		return nil, err
	}

	if describeRes != nil && describeRes.Nodegroup != nil {
		if err := ng.patchStatus(*describeRes.Nodegroup.NodegroupName, string(describeRes.Nodegroup.Status)); err != nil {
			return nil, err
		}
	}

	store.Add(ng.EksEnv.Env.Spec.CloudInfra.Eks.Name, nodeGroupName)

	return &awseks.CreateNodegroupOutput{}, nil
}

func (eksEnv *EksEnvironment) syncNodegroup() error {
	// update node group if spec node group is updated
	return nil
}

func (eksEnv *EksEnvironment) NodeGroupExists(ngName string) bool {
	eksClient := awseks.NewFromConfig(eksEnv.Config)

	_, err := eksClient.DescribeNodegroup(eksEnv.Context, &awseks.DescribeNodegroupInput{
		ClusterName:   &eksEnv.Env.Spec.CloudInfra.Eks.Name,
		NodegroupName: &ngName,
	})
	if err != nil {
		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {
			return false
		}
	}

	return true
}

func getNodegroupSpecForTenantSize(env *v1.Environment, tenantName string, tenantConfig v1.TenantConfig) (map[v1.NodeGroupName]*v1.NodeGroupSpec, error) {
	for _, size := range env.Spec.Size {
		if size.Name == tenantConfig.Size && size.Spec.AppType == tenantConfig.AppType {
			return size.Spec.Nodes, nil
		}
	}
	return nil, fmt.Errorf("no NodegroupSpec for app %s & size %s", tenantName, tenantConfig.Size)
}
