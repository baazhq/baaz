package eks

import (
	"context"
	"errors"
	"fmt"

	"datainfra.io/ballastdata/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "datainfra.io/ballastdata/api/v1"
	app "datainfra.io/ballastdata/pkg/application"
	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"k8s.io/klog/v2"
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

	eksClient := awseks.NewFromConfig(ng.EksEnv.Config)

	switch ng.AppConfig.AppType {

	case v1.ClickHouse:

		systemNgName := *aws.String(app.MakeSystemNodeGroupName(ng.AppConfig.Name))
		chiNgName := *aws.String(app.MakeChiNodeGroupName(ng.AppConfig.Name))
		zkChiNgName := *aws.String(app.MakeZkChiNodeGroupName(ng.AppConfig.Name))

		// System Pool
		describeRes, err := ng.describeNodegroup(systemNgName)
		if err != nil {
			var ngNotFound *types.ResourceNotFoundException
			if errors.As(err, &ngNotFound) {
				systemNodeGroup := ng.getNodeGroup(systemNgName)
				result, err := eksClient.CreateNodegroup(ng.Ctx, systemNodeGroup)
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

		// Clickhouse Pool
		describeRes, err = ng.describeNodegroup(chiNgName)
		if err != nil {
			var ngNotFound *types.ResourceNotFoundException
			if errors.As(err, &ngNotFound) {
				chiNodeGroup := ng.getNodeGroup(chiNgName)
				result, err := eksClient.CreateNodegroup(ng.Ctx, chiNodeGroup)
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

		// Zk Pool
		describeRes, err = ng.describeNodegroup(zkChiNgName)
		if err != nil {
			var ngNotFound *types.ResourceNotFoundException
			if errors.As(err, &ngNotFound) {
				zkChiNodeGroup := ng.getNodeGroup(zkChiNgName)
				result, err := eksClient.CreateNodegroup(ng.Ctx, zkChiNodeGroup)
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
	}

	return &CreateNodeGroupOutput{}, nil
}

func (ng *NodeGroup) getNodeGroup(name string) *awseks.CreateNodegroupInput {

	return &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(ng.EksEnv.Env.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String("arn:aws:iam::437639712640:role/pulak-eks-node-role"),
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
		Taints:       *app.MakeTaints(name),
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
