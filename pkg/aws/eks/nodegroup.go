package eks

import (
	"context"
	"errors"
	"fmt"

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

type CreateNodeGroupOutput struct {
	Result *awseks.CreateNodegroupOutput `json:"result"`
}

type DescribeNodegroupOutput struct {
	Result *awseks.DescribeNodegroupOutput `json:"result"`
}

type NodeGroups interface {
	CreateNodeGroupForApp() (*CreateNodeGroupOutput, error)
}

type NodeGroup struct {
	Ctx       context.Context
	EksEnv    *EksEnvironment
	AppConfig *v1.ApplicationConfig
	NodeSpec  *v1.NodeGroupSpec
}

func NewNodeGroup(
	ctx context.Context,
	eksEnv *EksEnvironment,
	appConfig *v1.ApplicationConfig,
	nodeSpec *v1.NodeGroupSpec,
) NodeGroups {
	ngs := &NodeGroup{
		Ctx:       ctx,
		EksEnv:    eksEnv,
		AppConfig: appConfig,
		NodeSpec:  nodeSpec,
	}
	return ngs
}

func (ng *NodeGroup) CreateNodeGroupForApp() (*CreateNodeGroupOutput, error) {

	switch ng.AppConfig.AppType {

	case v1.ClickHouse:

		systemNgName := *aws.String(makeSystemNodeGroupName(ng.AppConfig.Name))
		chiNgName := *aws.String(makeChiNodeGroupName(ng.AppConfig.Name))
		zkChiNgName := *aws.String(makeZkChiNodeGroupName(ng.AppConfig.Name))

		// system nodepool
		_, err := ng.createOrUpdateNodeGroup(systemNgName, system)
		if err != nil {
			return &CreateNodeGroupOutput{}, nil
		}

		// clickhouse nodepool
		_, err = ng.createOrUpdateNodeGroup(chiNgName, app)
		if err != nil {
			return &CreateNodeGroupOutput{}, nil
		}

		// zookeeper nodepool
		_, err = ng.createOrUpdateNodeGroup(zkChiNgName, app)
		if err != nil {
			return &CreateNodeGroupOutput{}, nil
		}

	case v1.Druid:

		systemNgName := *aws.String(makeSystemNodeGroupName(ng.AppConfig.Name))
		druidDataNodeNgName := *aws.String(makeDruidDataNodeGroupName(ng.AppConfig.Name))
		druidQueryNodeNgName := *aws.String(makeDruidQueryNodeNodeGroupName(ng.AppConfig.Name))
		druidMasterNodeNgName := *aws.String(makeDruidMasterNodeNodeGroupName(ng.AppConfig.Name))

		// system nodepool
		_, err := ng.createOrUpdateNodeGroup(systemNgName, system)
		if err != nil {
			return &CreateNodeGroupOutput{}, nil
		}

		// druid datanodes nodepool
		_, err = ng.createOrUpdateNodeGroup(druidDataNodeNgName, app)
		if err != nil {
			return &CreateNodeGroupOutput{}, nil
		}
		// druid querynode nodepool
		_, err = ng.createOrUpdateNodeGroup(druidQueryNodeNgName, app)
		if err != nil {
			return &CreateNodeGroupOutput{}, nil
		}

		// druid masternode nodepool
		_, err = ng.createOrUpdateNodeGroup(druidMasterNodeNgName, app)
		if err != nil {
			return &CreateNodeGroupOutput{}, nil
		}

	}

	return &CreateNodeGroupOutput{}, nil
}

func (ng *NodeGroup) getNodeGroup(name string, ngType nodeGroupType) *awseks.CreateNodegroupInput {

	var taints = []types.Taint{}

	if ngType == app {
		taints = *makeTaints(name)
	}

	nodeRole, err := ng.EksEnv.createNodeIamRole(name)
	if err != nil {
		return &awseks.CreateNodegroupInput{}
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
		InstanceTypes:      []string{ng.NodeSpec.NodeSize},
		Labels:             ng.NodeSpec.NodeLabels,
		LaunchTemplate:     nil,
		ReleaseVersion:     nil,
		RemoteAccess:       nil,
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(ng.NodeSpec.Min),
			MaxSize:     aws.Int32(ng.NodeSpec.Max),
			MinSize:     aws.Int32(ng.NodeSpec.Min),
		},
		Tags: map[string]string{
			fmt.Sprintf("kubernetes.io/cluster/%s", ng.EksEnv.Env.Spec.CloudInfra.Eks.Name): "owned",
		},
		Taints:       taints,
		UpdateConfig: nil,
		Version:      nil,
	}
}

func (ng *NodeGroup) describeNodegroup(name string) (*DescribeNodegroupOutput, error) {
	eksClient := awseks.NewFromConfig(ng.EksEnv.Config)
	input := &awseks.DescribeNodegroupInput{
		ClusterName:   aws.String(ng.EksEnv.Env.Spec.CloudInfra.Eks.Name),
		NodegroupName: aws.String(name),
	}

	result, err := eksClient.DescribeNodegroup(ng.Ctx, input)
	if err != nil {
		return nil, err
	}
	return &DescribeNodegroupOutput{Result: result}, nil
}

func (ng *NodeGroup) patchStatus(name, status string) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ng.Ctx, ng.EksEnv.Client, ng.EksEnv.Env, func(obj client.Object) client.Object {
		in := obj.(*v1.Environment)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[name] = status
		return in
	})
	return err
}

func (ng *NodeGroup) createOrUpdateNodeGroup(nodeGroupName string, ngType nodeGroupType) (*CreateNodeGroupOutput, error) {
	eksClient := awseks.NewFromConfig(ng.EksEnv.Config)

	describeRes, err := ng.describeNodegroup(nodeGroupName)
	if err != nil {
		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {
			nodeGroup := ng.getNodeGroup(nodeGroupName, ngType)
			result, err := eksClient.CreateNodegroup(ng.Ctx, nodeGroup)
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
		return nil, err
	}
	if describeRes != nil && describeRes.Result != nil && describeRes.Result.Nodegroup != nil {
		if err := ng.patchStatus(*describeRes.Result.Nodegroup.NodegroupName, string(describeRes.Result.Nodegroup.Status)); err != nil {
			return nil, err
		}
	}
	return &CreateNodeGroupOutput{}, nil
}
