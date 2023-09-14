package tenant_controller

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	v1 "datainfra.io/ballastdata/api/v1/types"
	"datainfra.io/ballastdata/pkg/aws/eks"
	"datainfra.io/ballastdata/pkg/resources"
	"datainfra.io/ballastdata/pkg/store"
	"datainfra.io/ballastdata/pkg/utils"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go/aws"
	k8stypes "k8s.io/apimachinery/pkg/types"
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
	dp     *v1.DataPlanes
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

			if ae.dp.Status.NodegroupStatus[string(nodeSpec.Name)] != "DELETING" {

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

					if describeNodegroupOutput.Nodegroup.Status == types.NodegroupStatusActive {
						clientset, err := ae.eksIC.GetEksClientSet()
						if err != nil {
							return err
						}

						if err := ae.createNamespace(clientset); err != nil {
							return err
						}

						if err := ae.createOrUpdateNetworkPolicy(clientset); err != nil {
							return err
						}

						if err := ae.tenantExpansion(nodeSpec.Size); err != nil {
							return err
						}
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

func (ae *awsEnv) tenantExpansion(desiredSize string) error {
	currentTenantObj := &v1.Tenants{}
	err := ae.client.Get(ae.ctx, k8stypes.NamespacedName{
		Namespace: ae.tenant.Namespace,
		Name:      ae.tenant.Name,
	}, currentTenantObj)
	if err != nil {
		return err
	}

	for _, tenantConfig := range currentTenantObj.Spec.TenantConfig {
		nodeSpecs, err := ae.getNodeSpecForTenantSize(tenantConfig)
		if err != nil {
			return err
		}

		for _, nodeSpec := range *nodeSpecs {
			fmt.Println(nodeSpec.Size)
			fmt.Println(desiredSize)
			if nodeSpec.Size != desiredSize {
				klog.Infof(
					"Tenant Expansion: Tenant [%s], Node Name [%s], Current Size [%s], Desired Size [%s]",
					currentTenantObj.Name,
					currentTenantObj.Name,
					nodeSpec.Size,
					desiredSize,
				)
			}
		}
	}

	return nil
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

	var taints = &[]types.Taint{}

	if ae.tenant.Spec.Isolation.Machine.Enabled {
		taints = makeTaints(nodeName)
	}

	return &awseks.CreateNodegroupInput{
		ClusterName:        aws.String(ae.dp.Spec.CloudInfra.Eks.Name),
		NodeRole:           aws.String(roleArn),
		NodegroupName:      aws.String(nodeName),
		Subnets:            ae.dp.Spec.CloudInfra.AwsCloudInfraConfig.Eks.SubnetIds,
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
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.tenant, func(obj client.Object) client.Object {
		in := obj.(*v1.Tenants)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[name] = status
		in.Status.Phase = v1.DataPlanePhase(status)
		return in
	})
	return err
}

func (ae *awsEnv) createNamespace(clientset *kubernetes.Clientset) error {

	_, err := clientset.CoreV1().Namespaces().Get(ae.ctx, ae.tenant.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		ns, err := clientset.CoreV1().Namespaces().Create(ae.ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ae.tenant.Name,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		klog.Infof("Namespace [%s] created for environment [%s]", ns.Name, ae.dp.Name)

	}
	return nil

}

func (ae *awsEnv) createOrUpdateNetworkPolicy(clientset *kubernetes.Clientset) error {

	networkPolicyName := ae.tenant.Name + "-network-policy"

	_, err := clientset.NetworkingV1().NetworkPolicies(ae.tenant.Name).Get(ae.ctx, networkPolicyName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		ns, err := clientset.NetworkingV1().NetworkPolicies(ae.tenant.Name).Create(ae.ctx,
			resources.MakeNetworkPolicy(
				networkPolicyName,
				ae.tenant.Name,
				ae.tenant.Spec.Isolation.Network.AllowNamespaces,
				resources.MakeOwnerRef(ae.tenant.APIVersion, ae.tenant.Kind, ae.tenant.Name, ae.tenant.UID),
			), metav1.CreateOptions{},
		)
		if err != nil {
			return err
		}
		klog.Infof("Network Policy [%s] created for tenant [%s]", ns.Name, ae.tenant.Name)
	}

	// TODO
	// update logic
	return nil
}
