package controller

import (
	"context"
	"errors"
	"fmt"
	mrand "math/rand"
	"os"
	"strings"
	"time"

	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/baazhq/baaz/api/v1/types"
	"github.com/baazhq/baaz/pkg/aws/eks"
	"github.com/baazhq/baaz/pkg/aws/network"
	"github.com/baazhq/baaz/pkg/helm"
	"github.com/baazhq/baaz/pkg/store"
	"github.com/baazhq/baaz/pkg/utils"
)

const (
	awsEbsCsiDriver string = "aws-ebs-csi-driver"
	vpcCni          string = "vpc-cni"
)

func (r *DataPlaneReconciler) reconcileAwsEnvironment(ctx context.Context, dp *v1.DataPlanes) error {

	eksClient := eks.NewEks(ctx, dp)
	network, err := network.NewProvisioner(ctx, dp.Spec.CloudInfra.Region)
	if err != nil {
		return err
	}

	awsEnv := &awsEnv{
		ctx:     ctx,
		dp:      dp,
		eksIC:   eksClient,
		client:  r.Client,
		store:   r.NgStore,
		network: network,
	}

	if err := awsEnv.reconcileNetwork(ctx); err != nil {
		return fmt.Errorf("error in reconciling network: %s", err.Error())
	}

	if err := awsEnv.reconcileAwsEks(); err != nil {
		return fmt.Errorf("error in reconciling aws eks cluster: %s", err.Error())
	}

	if err := awsEnv.reconcileClusterAutoscaler(); err != nil {
		return fmt.Errorf("error in reconciling cluster autoscaler: %s", err.Error())
	}

	// bootstrap dataplane with apps
	if err := awsEnv.reconcileAwsApplications(); err != nil {
		return fmt.Errorf("error in reconciling applications: %s", err.Error())
	}

	return nil
}

type awsEnv struct {
	ctx     context.Context
	dp      *v1.DataPlanes
	eksIC   eks.Eks
	client  client.Client
	store   store.Store
	network network.Network
}

var (
	casIamPolicy = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"autoscaling:DescribeAutoScalingGroups",
					"autoscaling:DescribeAutoScalingInstances",
					"autoscaling:DescribeLaunchConfigurations",
					"autoscaling:DescribeScalingActivities",
					"ec2:DescribeInstanceTypes",
					"ec2:DescribeLaunchTemplateVersions"
				],
				"Resource": ["*"]
			},
			{
				"Effect": "Allow",
				"Action": [
					"autoscaling:SetDesiredCapacity",
					"autoscaling:TerminateInstanceInAutoScalingGroup"
				],
				"Resource": ["*"]
			}
		]
	}`
)

const (
	CASPolicyName = "cas-policy"
)

/* Error faced

E0519 09:29:10.091183       1 aws_manager.go:128] Failed to regenerate ASG cache: AccessDenied: User: arn:aws:sts::437639712640:assumed-role/aws-us-east-1-owkb-system-node-role/i-06bb159e4a93c9753 is not authorized to perform: autoscaling:DescribeAutoScalingGroups because no identity-based policy allows the autoscaling:DescribeAutoScalingGroups action
	status code: 403, request id: 6a1b2526-8012-4f05-a5f5-4fb783a352b3
F0519 09:29:10.091232       1 aws_cloud_provider.go:460] Failed to create AWS Manager: AccessDenied: User: arn:aws:sts::437639712640:assumed-role/aws-us-east-1-owkb-system-node-role/i-06bb159e4a93c9753 is not authorized to perform: autoscaling:DescribeAutoScalingGroups because no identity-based policy allows the autoscaling:DescribeAutoScalingGroups action
	status code: 403, request id: 6a1b2526-8012-4f05-a5f5-4fb783a352b3

*/

func (ae *awsEnv) reconcileClusterAutoscaler() error {
	klog.Info("reconciling cluster autoscaler")

	if ae.dp.Status.NodegroupStatus[ae.dp.Spec.CloudInfra.Eks.Name+"-system"] != string(types.NodegroupStatusActive) {
		return nil
	}

	if ae.dp.Status.ClusterAutoScalerPolicyArn == "" {
		policyInput := &iam.CreatePolicyInput{
			PolicyDocument: aws.String(casIamPolicy),
			PolicyName:     aws.String(ae.dp.Spec.CloudInfra.Eks.Name + "-cas-policy"),
		}

		policyOutput, err := ae.eksIC.CreateIAMPolicy(ae.ctx, policyInput)
		if err != nil {
			return err
		}

		_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			in.Status.ClusterAutoScalerPolicyArn = *policyOutput.Policy.Arn
			return in
		})
		if err != nil {
			return err
		}
	}

	if ae.dp.Status.ClusterAutoScalerPolicyArn != "" {
		roles, err := ae.eksIC.GetClusterNodeRoles()
		if err != nil {
			return err
		}

		for _, r := range roles {
			attachRolePolicyInput := &iam.AttachRolePolicyInput{
				PolicyArn: &ae.dp.Status.ClusterAutoScalerPolicyArn,
				RoleName:  &r,
			}

			_, err = ae.eksIC.AttachRolePolicy(ae.ctx, attachRolePolicyInput)
			if err != nil {
				return err
			}
		}
	}

	restConfig, err := ae.eksIC.GetRestConfig()
	if err != nil {
		return err
	}

	ch := make(chan ChartCh)

	chartValues := []string{fmt.Sprintf("autoDiscovery.clusterName=%s", ae.dp.Spec.CloudInfra.Eks.Name)}

	if ae.dp.Status.ClusterAutoScalerStatus == v1.DeployedA || ae.dp.Status.ClusterAutoScalerStatus == v1.InstallingA {
		return nil
	}

	if ae.dp.Status.ClusterAutoScalerStatus != v1.DeployedA {
		helm := helm.NewHelm(
			"cas",
			"kube-system",
			"cluster-autoscaler",
			"autoscaler",
			"https://kubernetes.github.io/autoscaler",
			"9.37.0",
			restConfig,
			chartValues,
		)

		_, exists := helm.List(restConfig)

		if !exists {
			go func(ch chan ChartCh) {
				c := ChartCh{
					Name: "cas",
					Err:  nil,
				}
				if err := helm.Apply(restConfig); err != nil {
					c.Err = err
				}
				ch <- c
			}(ch)

			_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				in.Status.ClusterAutoScalerStatus = v1.InstallingA
				return in
			})
			if err != nil {
				return err
			}

			chartCh := <-ch
			var latestState v1.ApplicationPhase
			if chartCh.Err != nil {
				klog.Errorf("installing chart %s failed, reason: %s", chartCh.Name, chartCh.Err.Error())
				latestState = v1.FailedA
			} else {
				latestState = v1.DeployedA
			}

			_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				in.Status.ClusterAutoScalerStatus = latestState
				return in
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (ae *awsEnv) reconcileAwsEks() error {

	eksDescribeClusterOutput, err := ae.eksIC.DescribeEks()
	if err != nil {
		var ngNotFound *types.ResourceNotFoundException
		if errors.As(err, &ngNotFound) {
			klog.Infof("Creating EKS Control plane: %s for Environment: %s/%s", ae.dp.Spec.CloudInfra.Eks.Name, ae.dp.Namespace, ae.dp.Name)
			klog.Info("Updating Environment status to creating")

			clusterRoleOutput, err := ae.eksIC.CreateClusterIamRole()
			if err != nil {
				return fmt.Errorf("failed to create cluster iam role: %s", err.Error())
			}

			klog.Infof("Cluster Role [%s] Created", *clusterRoleOutput.Role.RoleName)

			createEksResult := ae.eksIC.CreateEks()
			if createEksResult.Success {
				fmt.Println(createEksResult.Success)
				if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
					in := obj.(*v1.DataPlanes)
					in.Status.Phase = v1.CreatingD
					in.Status.Conditions = in.AddCondition(v1.DataPlaneCondition{
						Type:               v1.ControlPlaneCreateInitiated,
						Status:             corev1.ConditionTrue,
						LastUpdateTime:     metav1.Time{Time: time.Now()},
						LastTransitionTime: metav1.Time{Time: time.Now()},
						Reason:             string(eks.EksControlPlaneCreationInitatedReason),
						Message:            string(eks.EksControlPlaneCreationInitatedMsg),
					})
					return in
				}); err != nil {
					fmt.Println(err)
					return err
				}

				klog.Info("Successfully initiated kubernetes control plane")
				return nil

			} else {
				klog.Errorf(createEksResult.Result)
			}

		} else {
			klog.Errorf("Dataplane Describe Fail [%s]", err.Error())
		}

	}
	if eksDescribeClusterOutput != nil {
		if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusActive {
			// checking for version upgrade
			statusVersion := ae.dp.Status.Version
			specVersion := ae.dp.Spec.CloudInfra.Eks.Version
			if statusVersion != "" && statusVersion != specVersion && *eksDescribeClusterOutput.Cluster.Version != specVersion {
				klog.Info("Updating Kubernetes version to: ", ae.dp.Spec.CloudInfra.Eks.Version)
				if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
					in := obj.(*v1.DataPlanes)
					in.Status.Phase = v1.UpdatingD
					in.Status.Conditions = in.AddCondition(v1.DataPlaneCondition{
						Type:               v1.VersionUpgradeInitiated,
						Status:             corev1.ConditionTrue,
						LastUpdateTime:     metav1.Time{Time: time.Now()},
						LastTransitionTime: metav1.Time{Time: time.Now()},
						Reason:             string(eks.EksControlPlaneUpgradedReason),
						Message:            string(eks.EksControlPlaneUpgradedIntiatedMsg),
					})
					return in
				}); err != nil {
					return err
				}
				result := ae.eksIC.UpdateEks()
				if !result.Success {
					return errors.New(result.Result)
				}
				klog.Info("Successfully initiated version update")
			}

			klog.Info("Sync Cluster status and version")

			if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				in.Status.Phase = v1.ActiveD
				in.Status.Version = in.Spec.CloudInfra.Eks.Version
				in.Status.Conditions = in.AddCondition(v1.DataPlaneCondition{
					Type:               v1.ControlPlaneCreated,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     metav1.Time{Time: time.Now()},
					LastTransitionTime: metav1.Time{Time: time.Now()},
					Reason:             string(eks.EksControlPlaneCreatedReason),
					Message:            string(eks.EksControlPlaneCreatedMsg),
				})
				return in
			}); err != nil {
				return err
			}

		} else if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusCreating {
			klog.Infof("EKS Cluster Control Plane [%s] in creating state", ae.dp.Spec.CloudInfra.Eks.Name)
			if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				in.Status.Phase = v1.ActiveD
				in.Status.Version = in.Spec.CloudInfra.Eks.Version
				in.Status.Conditions = in.AddCondition(v1.DataPlaneCondition{
					Type:               v1.DataPlaneConditionType(v1.ActiveD),
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     metav1.Time{Time: time.Now()},
					LastTransitionTime: metav1.Time{Time: time.Now()},
					Reason:             string(eks.EksControlPlaneProvisioningMsg),
					Message:            string(eks.EksControlPlaneProvisioningReason),
				})
				return in
			}); err != nil {
				return err
			}
		} else if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusUpdating {
			klog.Infof("EKS Cluster Control Plane [%s] in updated state", ae.dp.Spec.CloudInfra.Eks.Name)
		} else if eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusDeleting {
			klog.Infof("EKS Cluster Control Plane [%s] in deleting state", ae.dp.Spec.CloudInfra.Eks.Name)
		}

	}

	if eksDescribeClusterOutput != nil && eksDescribeClusterOutput.Cluster != nil && eksDescribeClusterOutput.Cluster.Status == types.ClusterStatusActive {

		oidcOutput, err := ae.reconcileOIDCProvider(eksDescribeClusterOutput)
		if err != nil {
			return err
		}

		if oidcOutput != nil && oidcOutput.OpenIDConnectProviderArn != nil {
			_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				in.Status.CloudInfraStatus.EksStatus.OIDCProviderArn = *oidcOutput.OpenIDConnectProviderArn
				return in
			})
			if err != nil {
				return err
			}
		}

		if err := ae.reconcileSystemNodeGroup(); err != nil {
			return err
		}

		if err := ae.ReconcileDefaultAddons(); err != nil {
			return err
		}

		if err := ae.reconcilePhase(); err != nil {
			return err
		}

		if err := ae.reconcileLBPhase(); err != nil {
			return err
		}
	}

	return nil

}

func (ae *awsEnv) reconcileLBPhase() error {
	eksClient, err := ae.eksIC.GetEksClientSet()
	if err != nil {
		return err
	}

	services, err := eksClient.CoreV1().Services(corev1.NamespaceAll).List(ae.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	lbArns := []string{}
	for _, svc := range services.Items {
		if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			for _, in := range svc.Status.LoadBalancer.Ingress {
				if strings.Contains(in.Hostname, ".amazonaws.com") {
					data := strings.Split(in.Hostname, "-")
					if len(data) >= 2 {
						lbArns = append(lbArns, data[0])
					}
				}
			}
		}
	}

	_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
		in := obj.(*v1.DataPlanes)
		in.Status.CloudInfraStatus.LBArns = lbArns
		return in
	})
	return err
}

// reconcileNetwork ensures that the network infrastructure for the data plane is correctly set up.
// The function performs the following steps:
//
// 1. Checks if network provisioning is enabled in the specification.
// 2. Generates a random CIDR block for the VPC.
// 3. Creates a VPC if it doesn't exist and updates the status with the VPC ID.
// 4. Creates and attaches an Internet Gateway to the VPC if it doesn't exist and updates the status with the Internet Gateway ID.
// 5. Creates a Route Table if it doesn't exist and updates the status with the Route Table ID.
// 6. Creates a default route in the Route Table to the Internet Gateway.
// 7. Defines CIDR blocks for subnets and creates them if they don't exist, updating the status with the Subnet IDs. Public subnets are auto-assigned public IPs and associated with the Route Table.
// 8. Creates a Security Group if it doesn't exist and updates the status with the Security Group IDs.
// 9. Adds an inbound rule to the Security Group if it hasn't been added and updates the status.
// 10. Creates a NAT Gateway if it doesn't exist and updates the status with the NAT Gateway ID.
// 11. Associates the NAT Gateway with the Route Table if it hasn't been done and updates the status.
//
// Flow Chart:
//
// +--------------------------------+
// | Start                          |
// +--------------------------------+
//
//	|
//	v
//
// +--------------------------------+
// | Check if network provisioning  |
// | is enabled                     |
// +--------------------------------+
//
//	|
//	v
//
// +-------------------+  No   +----------------+
// | ProvisionNetwork? |------>| Return nil     |
// |                   |       +----------------+
// |        Yes        |
// +-------------------+
//
//	|
//	v
//
// +--------------------------------+
// | Generate random CIDR block for |
// | VPC if user doesn't specify    |
// | vpc cidr range                 |
// +--------------------------------+
//
//	|
//	v
//
// +---------------------+  No   +----------------------+
// | Check if VPC exists |------>| Create VPC, update   |
// |                     |       | status with VPC ID   |
// |         Yes         |       +----------------------+
// +---------------------+
//
//	|
//	v
//
// +---------------------------+  No   +----------------------+
// | Check if Internet Gateway |------>| Create Internet      |
// | exists                    |       | Gateway, update      |
// |                           |       | status with IGW ID   |
// |            Yes            |       +----------------------+
// +---------------------------+
//
//	|
//	v
//
// +-------------------------+  No   +--------------------------+
// | Check if Route Table    |------>| Create Route Table,      |
// | exists                  |       | update status with RT ID |
// |                         |       +--------------------------+
// |            Yes          |
// +-------------------------+
//
//	|
//	v
//
// +----------------------------+
// | Create default route to    |
// | Internet Gateway           |
// +----------------------------+
//
//	|
//	v
//
// +----------------------------+
// | Define CIDR blocks for     |
// | subnets                    |
// +----------------------------+
//
//	|
//	v
//
// +-------------------------+  No   +-----------------------+
// | Check if subnets exist  |------>| Create subnets,       |
// |                         |       | update status with    |
// |            Yes          |       | Subnet IDs            |
// +-------------------------+       +-----------------------+
//
//	|
//	v
//
// +-------------------------+  No   +-----------------------+
// | Check if Security Group |------>| Create Security Group,|
// | exists                  |       | update status with SG |
// |                         |       | IDs                   |
// |            Yes          |       +-----------------------+
// +-------------------------+
//
//	|
//	v
//
// +--------------------------+  No   +-----------------------+
// | Check if inbound rule    |------>| Add inbound rule,     |
// | is added to Security     |       | update status         |
// | Group                    |       +-----------------------+
// |            Yes           |
// +--------------------------+
//
//	|
//	v
//
// +-------------------------+  No   +-----------------------+
// | Check if NAT Gateway    |------>| Create NAT Gateway,   |
// | exists                  |       | update status with NAT|
// |                         |       | Gateway ID            |
// |            Yes          |       +-----------------------+
// +-------------------------+
//
//	|
//	v
//
// +-------------------------+  No   +-----------------------+
// | Check if NAT Gateway    |------>| Associate NAT Gateway,|
// | is associated with Route|       | update status         |
// | Table                   |       +-----------------------+
// |            Yes          |
// +-------------------------+
//
//	|
//	v
//
// +--------------------------------+
// | End                            |
// +--------------------------------+
func (ae *awsEnv) reconcileNetwork(ctx context.Context) error {
	// Check if network provisioning is enabled in the specification
	if !ae.dp.Spec.CloudInfra.ProvisionNetwork {
		return nil
	}

	// Generate a random number between 0 and 253 for CIDR block allocation
	cidrRandom := mrand.Intn(254)

	vpcId := ae.dp.Status.CloudInfraStatus.Vpc
	vpcName := fmt.Sprintf("%s-%s", ae.dp.Name, ae.dp.Namespace)

	// Create VPC if not already created
	if vpcId == "" {
		vpcCidr := ae.dp.Spec.CloudInfra.VpcCidr
		if vpcCidr == "" {
			vpcCidr = fmt.Sprintf("10.%d.0.0/16", cidrRandom)
		}

		vpc, err := ae.network.CreateVPC(ctx, &awsec2.CreateVpcInput{
			CidrBlock: &vpcCidr,
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeVpc,
					Tags: []ec2types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(vpcName),
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}
		// Update VPC ID in status
		upObj, _, err := utils.PatchStatus(ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			in.Status.CloudInfraStatus.Vpc = *vpc.Vpc.VpcId

			return in
		})
		if err != nil {
			return err
		}
		vpcId = *vpc.Vpc.VpcId
		ae.dp = upObj.(*v1.DataPlanes)
	}

	// Create and attach Internet Gateway if not already created
	if ae.dp.Status.CloudInfraStatus.InternetGatewayId == "" {
		ig, err := ae.network.CreateInternetGateway(ctx, &awsec2.CreateInternetGatewayInput{
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeInternetGateway,
					Tags: []ec2types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(fmt.Sprintf("%s-%s-ig", ae.dp.Name, ae.dp.Namespace)),
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}
		// Update Internet Gateway ID in status
		upObj, _, err := utils.PatchStatus(ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			in.Status.CloudInfraStatus.InternetGatewayId = *ig.InternetGateway.InternetGatewayId
			return in
		})
		if err != nil {
			return err
		}
		ae.dp = upObj.(*v1.DataPlanes)

		_, err = ae.network.AttachInternetGateway(ctx, *ig.InternetGateway.InternetGatewayId, vpcId)
		if err != nil {
			return err
		}
	}

	// Create Route Table if not already created
	if ae.dp.Status.CloudInfraStatus.PublicRTId == "" {
		rt, err := ae.network.CreateRouteTable(ctx, ae.dp.Status.CloudInfraStatus.Vpc, &awsec2.CreateRouteTableInput{
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeRouteTable,
					Tags: []ec2types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(fmt.Sprintf("%s-%s-rt", ae.dp.Name, ae.dp.Namespace)),
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}
		// Update Route Table ID in status
		upObj, _, err := utils.PatchStatus(ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			in.Status.CloudInfraStatus.PublicRTId = *rt.RouteTable.RouteTableId
			return in
		})
		if err != nil {
			return err
		}
		ae.dp = upObj.(*v1.DataPlanes)

		// Create default route to Internet Gateway
		if _, err := ae.network.CreateRoute(ctx, &awsec2.CreateRouteInput{
			RouteTableId:         rt.RouteTable.RouteTableId,
			GatewayId:            &ae.dp.Status.CloudInfraStatus.InternetGatewayId,
			DestinationCidrBlock: aws.String("0.0.0.0/0"),
		}); err != nil {
			return err
		}
	}

	// Define CIDR blocks for subnets
	subnetsCidr := []string{
		fmt.Sprintf("10.%d.16.0/20", cidrRandom),
		fmt.Sprintf("10.%d.32.0/20", cidrRandom),
		fmt.Sprintf("10.%d.0.0/20", cidrRandom),
		fmt.Sprintf("10.%d.80.0/20", cidrRandom),
	}

	// Create subnets
	for i := range subnetsCidr {
		if len(ae.dp.Status.CloudInfraStatus.SubnetIds) >= 4 {
			break
		}

		az := fmt.Sprintf("%s%c", ae.dp.Spec.CloudInfra.Region, 'a'+(i%3))
		subnetName := fmt.Sprintf("%s-%s", vpcName, az)
		subnetInput := &awsec2.CreateSubnetInput{
			VpcId:            &vpcId,
			CidrBlock:        &subnetsCidr[i],
			AvailabilityZone: &az,
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeSubnet,
					Tags: []ec2types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(subnetName),
						},
					},
				},
			},
		}

		subnet, err := ae.network.CreateSubnet(ctx, subnetInput)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// Auto-assign public IP for public subnets
		if i%2 == 0 {
			if _, err := ae.network.SubnetAutoAssignPublicIP(ctx, *subnet.Subnet.SubnetId); err != nil {
				return err
			}
			if err := ae.network.AssociateRTWithSubnet(ctx, ae.dp.Status.CloudInfraStatus.PublicRTId, *subnet.Subnet.SubnetId); err != nil {
				return err
			}
		}

		// Update Subnet IDs in status
		newObj, _, err := utils.PatchStatus(ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			if in.Status.CloudInfraStatus.SubnetIds == nil {
				in.Status.CloudInfraStatus.SubnetIds = make([]string, 0)
			}
			in.Status.CloudInfraStatus.SubnetIds = append(in.Status.CloudInfraStatus.SubnetIds, *subnet.Subnet.SubnetId)

			return in
		})
		if err != nil {
			return err
		}

		ae.dp = newObj.(*v1.DataPlanes)
	}

	// Create Security Group if not already created
	if len(ae.dp.Status.CloudInfraStatus.SecurityGroupIds) == 0 {
		sgName := fmt.Sprintf("%s-%s", ae.dp.Name, ae.dp.Namespace)

		sgDescription := fmt.Sprintf("sg for %s", ae.dp.Name)
		sgInput := &awsec2.CreateSecurityGroupInput{
			Description: &sgDescription,
			GroupName:   &sgName,
			VpcId:       &vpcId,
			TagSpecifications: []ec2types.TagSpecification{
				{
					ResourceType: ec2types.ResourceTypeSecurityGroup,
					Tags: []ec2types.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(sgName),
						},
					},
				},
			},
		}

		sg, err := ae.network.CreateSG(ctx, sgInput)
		if err != nil {
			return err
		}

		// Update Security Group IDs in status
		newObj, _, err := utils.PatchStatus(ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			if in.Status.CloudInfraStatus.SecurityGroupIds == nil {
				in.Status.CloudInfraStatus.SecurityGroupIds = make([]string, 0)
			}
			in.Status.CloudInfraStatus.SecurityGroupIds = append(in.Status.CloudInfraStatus.SecurityGroupIds, *sg.GroupId)

			return in
		})
		if err != nil {
			return err
		}
		ae.dp = newObj.(*v1.DataPlanes)
	}

	// Add inbound rule to Security Group if not already added
	if !ae.dp.Status.CloudInfraStatus.SGInboundRuleAdded && len(ae.dp.Status.CloudInfraStatus.SecurityGroupIds) > 0 {
		if _, err := ae.network.AddSGInboundRule(ctx, ae.dp.Status.CloudInfraStatus.SecurityGroupIds[0], ae.dp.Status.CloudInfraStatus.Vpc); err != nil {
			return err
		}

		// Update Security Group Inbound Rule status
		newObj, _, err := utils.PatchStatus(ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			in.Status.CloudInfraStatus.SGInboundRuleAdded = true

			return in
		})
		if err != nil {
			return err
		}

		ae.dp = newObj.(*v1.DataPlanes)
	}

	// Create NAT Gateway if not already created
	if ae.dp.Status.CloudInfraStatus.NATGatewayId == "" {
		nat, err := ae.network.CreateNAT(ctx, ae.dp)
		if err != nil {
			return err
		}

		// Update NAT Gateway ID in status
		upObj, _, err := utils.PatchStatus(ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			in.Status.CloudInfraStatus.NATGatewayId = *nat.NatGateway.NatGatewayId
			return in
		})
		if err != nil {
			return err
		}

		ae.dp = upObj.(*v1.DataPlanes)
	}

	// Associate NAT Gateway with Route Table if not already done
	if ae.dp.Status.CloudInfraStatus.NATGatewayId != "" && !ae.dp.Status.CloudInfraStatus.NATAttachedWithRT {
		if err := ae.network.AssociateNATWithRT(ctx, ae.dp); err != nil {
			return err
		}

		// Update NAT Gateway association status
		upObj, _, err := utils.PatchStatus(ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			in.Status.CloudInfraStatus.NATAttachedWithRT = true
			return in
		})
		if err != nil {
			return err
		}

		ae.dp = upObj.(*v1.DataPlanes)
	}

	return nil
}

func getChartName(app v1.AppSpec) string {
	return app.Spec.ChartName
}

type ChartCh struct {
	Name string
	Err  error
}

// bootstrap applications for aws eks dataplanes
// check if system nodepool is active
// once its active install applications
func (ae *awsEnv) reconcileAwsApplications() error {
	klog.Info("reconciling dataplane applications")

	if ae.dp.Status.NodegroupStatus[ae.dp.Spec.CloudInfra.Eks.Name+"-system"] != string(types.NodegroupStatusActive) {
		return nil
	}

	count := 0
	ch := make(chan ChartCh, len(ae.dp.Spec.Applications))

	for _, app := range ae.dp.Spec.Applications {

		chartStatus := ae.dp.Status.AppStatus[getChartName(app)]

		if chartStatus == v1.DeployedA {
			continue
		}

		restConfig, err := ae.eksIC.GetRestConfig()
		if err != nil {
			return err
		}

		helm := helm.NewHelm(
			app.Name,
			app.Namespace,
			app.Spec.ChartName,
			app.Spec.RepoName,
			app.Spec.RepoUrl,
			app.Spec.Version,
			restConfig,
			app.Spec.Values,
		)

		_, exists := helm.List(restConfig)

		if !exists {
			klog.Infof("installing chart: %s", app.Name)

			count += 1
			go func(ch chan ChartCh, app v1.AppSpec) {
				c := ChartCh{
					Name: getChartName(app),
					Err:  nil,
				}
				if err := helm.Apply(restConfig); err != nil {
					c.Err = err
				}
				ch <- c
			}(ch, app)

			_, _, err = utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
				in := obj.(*v1.DataPlanes)
				if in.Status.AppStatus == nil {
					in.Status.AppStatus = make(map[string]v1.ApplicationPhase)
				}
				in.Status.AppStatus[getChartName(app)] = v1.InstallingA
				return in
			})
			if err != nil {
				return err
			}
		}
	}

	for i := 0; i < count; i += 1 {
		chartCh := <-ch
		var latestState v1.ApplicationPhase
		if chartCh.Err != nil {
			klog.Errorf("installing chart %s failed, reason: %s", chartCh.Name, chartCh.Err.Error())
			latestState = v1.FailedA
		} else {
			latestState = v1.DeployedA
		}

		_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
			in := obj.(*v1.DataPlanes)
			if in.Status.AppStatus == nil {
				in.Status.AppStatus = make(map[string]v1.ApplicationPhase)
			}
			in.Status.AppStatus[chartCh.Name] = latestState
			return in
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (ae *awsEnv) reconcileOIDCProvider(clusterOutput *awseks.DescribeClusterOutput) (*awsiam.CreateOpenIDConnectProviderOutput, error) {
	if clusterOutput == nil || clusterOutput.Cluster == nil ||
		clusterOutput.Cluster.Identity == nil || clusterOutput.Cluster.Identity.Oidc == nil {
		return nil, errors.New("oidc provider url not found in cluster output")
	}
	oidcProviderUrl := *clusterOutput.Cluster.Identity.Oidc.Issuer

	// Compute the SHA-1 thumbprint of the OIDC provider certificate
	thumbprint, err := getIssuerCAThumbprint(oidcProviderUrl)
	if err != nil {
		return nil, err
	}

	input := &eks.CreateOIDCProviderInput{
		URL:            oidcProviderUrl,
		ThumbPrintList: []string{thumbprint},
	}

	oidcProviderArn := ae.dp.Status.CloudInfraStatus.EksStatus.OIDCProviderArn

	if oidcProviderArn != "" {
		// oidc provider is previously created
		// looking for it
		providers, err := ae.eksIC.ListOIDCProvider()
		if err != nil {
			return nil, err
		}

		for _, oidc := range providers.OpenIDConnectProviderList {
			if *oidc.Arn == ae.dp.Status.CloudInfraStatus.EksStatus.OIDCProviderArn {
				// oidc provider is already created and existed
				return nil, nil
			}
		}
	}

	result, err := ae.eksIC.CreateOIDCProvider(input)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ae *awsEnv) reconcilePhase() error {
	klog.Info("Calculating Environment Status")

	for node, status := range ae.dp.Status.NodegroupStatus {
		if status != string(types.NodegroupStatusActive) {
			klog.Infof("Node %s not active yet", node)
			return nil
		}
	}

	for addon, status := range ae.dp.Status.AddonStatus {
		if status != string(types.AddonStatusActive) {
			klog.Infof("Addon %s not active yet", addon)
			return nil
		}
	}

	if _, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
		in := obj.(*v1.DataPlanes)
		in.Status.Phase = v1.ActiveD
		return in
	}); err != nil {
		return err
	}
	return nil
}

func (ae *awsEnv) reconcileSystemNodeGroup() error {
	systemNodeGroupName := ae.dp.Spec.CloudInfra.Eks.Name + "-system"

	nodeRole, err := ae.eksIC.CreateNodeIamRole(systemNodeGroupName)
	if err != nil {
		return err
	}
	if nodeRole.Role == nil {
		return errors.New("node role is nil")
	}

	subnetIds := ae.dp.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds
	if ae.dp.Spec.CloudInfra.ProvisionNetwork {
		subnetIds = ae.dp.Status.CloudInfraStatus.SubnetIds
	}

	systemNodeGroupInput := &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(ae.dp.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String(*nodeRole.Role.Arn),
		NodegroupName:      aws.String(systemNodeGroupName),
		Subnets:            subnetIds,
		AmiType:            "",
		CapacityType:       "",
		ClientRequestToken: nil,
		DiskSize:           nil,
		InstanceTypes:      []string{os.Getenv("AWS_SYSTEM_NODEGROUP_SIZE")},
		Labels: map[string]string{
			"nodeType": "system",
			"name":     systemNodeGroupName,
		},
		LaunchTemplate: nil,
		ReleaseVersion: nil,
		RemoteAccess:   nil,
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(1),
			MaxSize:     aws.Int32(2),
			MinSize:     aws.Int32(1),
		},
		Tags: map[string]string{
			fmt.Sprintf("kubernetes.io/cluster/%s", ae.dp.Spec.CloudInfra.Eks.Name): "owned",
		},
		UpdateConfig: nil,
		Version:      nil,
	}

	describeNodeGroupOutput, found, err := ae.eksIC.DescribeNodegroup(systemNodeGroupName)
	if !found && err == nil {
		if ae.dp.DeletionTimestamp == nil {
			createSystemNodeGroupResult, err := ae.eksIC.CreateSystemNodeGroup(*systemNodeGroupInput)
			if err != nil {
				return err
			}

			if createSystemNodeGroupResult != nil && createSystemNodeGroupResult.Nodegroup != nil {
				klog.Infof("Initated NodeGroup Launch [%s]", *createSystemNodeGroupResult.Nodegroup.ClusterName)
				if err := ae.wrapNgPatchStatus(*createSystemNodeGroupResult.Nodegroup.NodegroupName, string(createSystemNodeGroupResult.Nodegroup.Status)); err != nil {
					return err
				}
			}
		}
	}

	if describeNodeGroupOutput != nil && describeNodeGroupOutput.Nodegroup != nil {
		if err := ae.wrapNgPatchStatus(*describeNodeGroupOutput.Nodegroup.NodegroupName, string(describeNodeGroupOutput.Nodegroup.Status)); err != nil {
			return err
		}
	}

	ae.store.Add(ae.dp.Spec.CloudInfra.Eks.Name, systemNodeGroupName)
	return nil
}

func (ae *awsEnv) wrapNgPatchStatus(name, status string) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
		in := obj.(*v1.DataPlanes)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[name] = status
		return in
	})
	return err
}

func (ae *awsEnv) wrapAddonPatchStatus(addonName, status string) error {
	// update status with current addon status
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.dp, func(obj client.Object) client.Object {
		in := obj.(*v1.DataPlanes)
		if in.Status.AddonStatus == nil {
			in.Status.AddonStatus = make(map[string]string)
		}
		in.Status.AddonStatus[addonName] = status
		return in
	})
	return err
}

func (ae *awsEnv) ReconcileDefaultAddons() error {
	oidcProvider := ae.dp.Status.CloudInfraStatus.AwsCloudInfraConfigStatus.EksStatus.OIDCProviderArn
	if oidcProvider == "" {
		klog.Info("ebs-csi-driver creation: waiting for oidcProvider to be created")
		return nil
	}
	clusterName := ae.dp.Spec.CloudInfra.Eks.Name
	ebsAddon, err := ae.eksIC.DescribeAddon(awsEbsCsiDriver)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			klog.Info("Creating aws-ebs-csi-driver addon")

			role, err := ae.eksIC.CreateEbsCSIRole(ae.ctx)
			if err != nil {
				return err
			}

			_, cErr := ae.eksIC.CreateAddon(ae.ctx, &awseks.CreateAddonInput{
				AddonName:             aws.String(awsEbsCsiDriver),
				ClusterName:           aws.String(clusterName),
				ResolveConflicts:      types.ResolveConflictsOverwrite,
				ServiceAccountRoleArn: role.Role.Arn,
			})
			if cErr != nil {
				return cErr
			}
			klog.Info("aws-ebs-csi-driver addon creation is initiated")
		} else {
			return err
		}
		return nil
	}
	if ebsAddon != nil && ebsAddon.Addon != nil {
		addonRes := ebsAddon.Addon
		klog.Info("aws-ebs-csi-driver addon status: ", addonRes.Status)
		if err := ae.wrapAddonPatchStatus(*addonRes.AddonName, string(addonRes.Status)); err != nil {
			return err
		}
	}

	vpcCniAddon, err := ae.eksIC.DescribeAddon(vpcCni)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			klog.Info("Creating vpc cni addon")
			_, arn, err := ae.eksIC.CreateVpcCniRole(ae.ctx)
			if err != nil {
				return err
			}

			v := `{"enableNetworkPolicy": "true"}`

			_, cErr := ae.eksIC.CreateAddon(ae.ctx, &awseks.CreateAddonInput{
				AddonName:             aws.String(vpcCni),
				ClusterName:           aws.String(clusterName),
				ResolveConflicts:      types.ResolveConflictsOverwrite,
				ServiceAccountRoleArn: aws.String(arn),
				AddonVersion:          aws.String("v1.15.0-eksbuild.2"),
				ConfigurationValues:   aws.String(v),
			})
			if cErr != nil {
				return cErr
			}
			klog.Info("vpc cni addon creation is initiated")
		} else {
			return err
		}
		return nil
	}
	if vpcCniAddon != nil && vpcCniAddon.Addon != nil {
		addonRes := vpcCniAddon.Addon
		klog.Info("vpc cni addon status: ", addonRes.Status)
		if err := ae.wrapAddonPatchStatus(*addonRes.AddonName, string(addonRes.Status)); err != nil {
			return err
		}
	}

	return nil
}
