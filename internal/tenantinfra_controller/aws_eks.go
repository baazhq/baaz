package tenantinfra_controller

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go/aws"
	v1 "github.com/baazhq/baaz/api/v1/types"
	"github.com/baazhq/baaz/pkg/aws/eks"
	"github.com/baazhq/baaz/pkg/store"
	"github.com/baazhq/baaz/pkg/utils"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
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

func getRandomSubnet(dp *v1.DataPlanes) string {
	subnets := dp.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds
	if dp.Spec.CloudInfra.ProvisionNetwork {
		subnets = dp.Status.CloudInfraStatus.SubnetIds
	}

	random := rand.Intn(100)
	return subnets[random%len(subnets)]
}

func getNodeGroupSubnet(tenants *v1.TenantsInfra, dp *v1.DataPlanes) string {
	// bytebeam-medium:
	// machinePool:
	// - name: bytebeam-app1
	//   #size: t2.small
	//   size: t2.medium
	// new name: bytebeam-medium-bytebeam-app1-t2-medium ()
	// old name: bytebeam-medium-bytebeam-app1-t2-small (status)
	// if len(tenants.Spec.TenantSizes) == len(tenants.Status.NodegroupStatus) {
	// 	return ""
	// }
	specFlags := make(map[string]bool)
	for tenantName, machineSpecs := range tenants.Spec.TenantSizes {
		for _, machineSpec := range machineSpecs.MachineSpec {
			node := getNodeName(tenantName, machineSpec)
			specFlags[node] = true
		}
	}

	for k, v := range tenants.Status.NodegroupStatus {
		if _, found := specFlags[k]; !found {
			return v.Subnet
		}
	}
	return getRandomSubnet(dp)
}

func getNodeName(tenantName string, machineSpec v1.MachineSpec) string {
	nodeName := fmt.Sprintf("%s-%s-%s", tenantName, machineSpec.Name, machineSpec.Size)
	nodeName = strings.ReplaceAll(nodeName, ".", "-")
	return nodeName
}

func (ae *awsEnv) ReconcileInfraTenants() error {
	klog.Info("Reconciling tenant infra node groups")

	for tenantName, machineSpecs := range ae.tenantsInfra.Spec.TenantSizes {

		for _, machineSpec := range machineSpecs.MachineSpec {

			if ae.dp.Status.NodegroupStatus[tenantName] != "DELETING" {
				nodeName := getNodeName(tenantName, machineSpec)
				describeNodegroupOutput, found, err := ae.eksIC.DescribeNodegroup(nodeName)
				if err != nil {
					return err
				}
				if !found {
					nodeRole, err := ae.eksIC.CreateNodeIamRole(nodeName)
					if err != nil {
						return err
					}
					if nodeRole.Role == nil {
						return errors.New("node role is nil")
					}

					subnet := getNodeGroupSubnet(ae.tenantsInfra, ae.dp)

					createNodeGroupOutput, err := ae.eksIC.CreateNodegroup(ae.getNodegroupInput(nodeName, *nodeRole.Role.Arn, subnet, &machineSpec))
					if err != nil {
						return err
					}
					if createNodeGroupOutput != nil && createNodeGroupOutput.Nodegroup != nil {
						klog.Infof("Initated NodeGroup Launch [%s]", *createNodeGroupOutput.Nodegroup.NodegroupName)
						if err := ae.patchStatus(*createNodeGroupOutput.Nodegroup.NodegroupName, &v1.NodegroupStatus{
							Status: string(createNodeGroupOutput.Nodegroup.Status),
							Subnet: subnet,
						}); err != nil {
							return err
						}
					}
				}

				if machineSpec.StrictScheduling == v1.StrictSchedulingStatusEnable &&
					machineSpec.Type == v1.MachineTypeLowPriority {
					dedicatedNodeName := fmt.Sprintf("%s-dedicated", nodeName)
					describeNodegroupOutput, found, err := ae.eksIC.DescribeNodegroup(dedicatedNodeName)
					if err != nil {
						return err
					}

					if !found {
						nodeRole, err := ae.eksIC.CreateNodeIamRole(dedicatedNodeName)
						if err != nil {
							return err
						}
						if nodeRole.Role == nil {
							return errors.New("node role is nil")
						}

						subnet := getNodeGroupSubnet(ae.tenantsInfra, ae.dp)
						input := ae.getNodegroupInput(dedicatedNodeName, *nodeRole.Role.Arn, subnet, &machineSpec)
						input.CapacityType = ""
						input.ScalingConfig.MinSize = aws.Int32(0)
						input.ScalingConfig.MaxSize = aws.Int32(1)
						input.ScalingConfig.DesiredSize = aws.Int32(0)

						createNodeGroupOutput, err := ae.eksIC.CreateNodegroup(input)
						if err != nil {
							return err
						}
						if createNodeGroupOutput != nil && createNodeGroupOutput.Nodegroup != nil {
							klog.Infof("Initated NodeGroup Launch [%s]", *createNodeGroupOutput.Nodegroup.NodegroupName)
							if err := ae.patchStatus(*createNodeGroupOutput.Nodegroup.NodegroupName, &v1.NodegroupStatus{
								Status: string(createNodeGroupOutput.Nodegroup.Status),
								Subnet: subnet,
							}); err != nil {
								return err
							}
						}
					}
					if describeNodegroupOutput != nil &&
						describeNodegroupOutput.Nodegroup != nil &&
						len(describeNodegroupOutput.Nodegroup.Subnets) > 0 {
						if err := ae.patchStatus(*describeNodegroupOutput.Nodegroup.NodegroupName, &v1.NodegroupStatus{
							Status: string(describeNodegroupOutput.Nodegroup.Status),
							Subnet: describeNodegroupOutput.Nodegroup.Subnets[0],
						}); err != nil {
							return err
						}
					}
				}

				if describeNodegroupOutput != nil &&
					describeNodegroupOutput.Nodegroup != nil &&
					len(describeNodegroupOutput.Nodegroup.Subnets) > 0 {
					if err := ae.patchStatus(*describeNodegroupOutput.Nodegroup.NodegroupName, &v1.NodegroupStatus{
						Status: string(describeNodegroupOutput.Nodegroup.Status),
						Subnet: describeNodegroupOutput.Nodegroup.Subnets[0],
					}); err != nil {
						return err
					}
				}
			}
		}

	}

	return ae.cleanUpUnusedNodeGroup()
}

func (ae *awsEnv) cleanUpUnusedNodeGroup() error {
	cleanupNodes := make(map[string]bool)
	for node, status := range ae.tenantsInfra.Status.NodegroupStatus {
		if status.Status != "ACTIVE" {
			return errors.New("nodegroups are not in ready state, clean up will happen later")
		}
		cleanupNodes[node] = true
	}

	for tenantName, machineSpecs := range ae.tenantsInfra.Spec.TenantSizes {
		for _, machineSpec := range machineSpecs.MachineSpec {
			nodeName := fmt.Sprintf("%s-%s-%s", tenantName, machineSpec.Name, machineSpec.Size)
			nodeName = strings.ReplaceAll(nodeName, ".", "-")
			dedicatedNodeName := fmt.Sprintf("%s-dedicated", nodeName)
			cleanupNodes[nodeName] = false
			cleanupNodes[dedicatedNodeName] = false
		}
	}

	for node, cleanup := range cleanupNodes {
		if cleanup {
			klog.Infof("going to cleanup & delete nodegroup: %s", node)
			ngOutput, found, err := ae.eksIC.DescribeNodegroup(node)
			if err != nil {
				return err
			}

			if !found {
				if err := ae.patchStatus(node, nil); err != nil {
					return err
				}
				continue
			}

			taintFound := false

			for _, tn := range ngOutput.Nodegroup.Taints {
				if tn.Effect == types.TaintEffectNoSchedule {
					taintFound = true
					break
				}
			}

			if !taintFound {
				taintKey := "key"
				taintValue := "value"
				taintEffect := types.TaintEffectNoSchedule

				newTaints := append(ngOutput.Nodegroup.Taints,
					types.Taint{Key: aws.String(taintKey), Value: aws.String(taintValue), Effect: taintEffect},
				)

				updateNodegroupInput := &awseks.UpdateNodegroupConfigInput{
					ClusterName:   aws.String(ae.dp.Spec.CloudInfra.Eks.Name),
					NodegroupName: aws.String(node),
					ScalingConfig: ngOutput.Nodegroup.ScalingConfig,
					Taints: &types.UpdateTaintsPayload{
						AddOrUpdateTaints: newTaints,
					},
				}

				_, err = ae.eksIC.UpdateNodegroup(updateNodegroupInput)
				if err != nil {
					return err
				}
			}

			clusterOutput, err := ae.eksIC.DescribeEks()
			if err != nil {
				return err
			}

			// Create a Kubernetes client
			clientset, err := newClientset(clusterOutput.Cluster)
			if err != nil {
				return err
			}

			nodeList, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			k8sNodeNames := make([]string, 0)

			for _, n := range nodeList.Items {
				if n.Labels["eks.amazonaws.com/nodegroup"] == node {
					k8sNodeNames = append(k8sNodeNames, n.Name)
				}
			}

			// Now you can use 'clientset' to interact with the Kubernetes API
			// For example, you can list pods in the cluster
			pods, err := clientset.CoreV1().Pods(core.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return err
			}

			for _, pod := range pods.Items {
				if slices.Contains(k8sNodeNames, pod.Spec.NodeName) {
					if err := clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{}); err != nil {
						return err
					}
				}
			}

			_, err = ae.eksIC.DeleteNodeGroup(node)
			if err != nil {
				return err
			}
			if err := ae.patchStatus(node, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func newClientset(cluster *types.Cluster) (*kubernetes.Clientset, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: aws.StringValue(cluster.Name),
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(
		&rest.Config{
			Host:        aws.StringValue(cluster.Endpoint),
			BearerToken: tok.Token,
			TLSClientConfig: rest.TLSClientConfig{
				CAData: ca,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return clientset, nil
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

func (ae *awsEnv) getNodegroupInput(nodeName, roleArn, subnet string, machineSpec *v1.MachineSpec) (input *awseks.CreateNodegroupInput) {

	var taints = &[]types.Taint{}

	taints = makeTaints(nodeName)

	var capacityType types.CapacityTypes
	if machineSpec.Type == v1.MachineTypeLowPriority {
		capacityType = types.CapacityTypesSpot
	}

	return &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(ae.dp.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String(roleArn),
		NodegroupName:      aws.String(nodeName),
		Subnets:            []string{subnet},
		AmiType:            "",
		CapacityType:       capacityType,
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

func (ae *awsEnv) patchStatus(name string, status *v1.NodegroupStatus) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.tenantsInfra, func(obj client.Object) client.Object {
		in := obj.(*v1.TenantsInfra)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]v1.NodegroupStatus)
		}
		if status == nil {
			delete(in.Status.NodegroupStatus, name)
			return in
		}
		in.Status.NodegroupStatus[name] = *status
		return in
	})
	return err
}
