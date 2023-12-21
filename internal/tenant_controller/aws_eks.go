package tenant_controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"

	v1 "datainfra.io/baaz/api/v1/types"
	"datainfra.io/baaz/pkg/aws/eks"
	"datainfra.io/baaz/pkg/resources"
	"datainfra.io/baaz/pkg/store"
	"datainfra.io/baaz/pkg/utils"
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
	klog.Info("Reconciling tenants")

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

	return nil
}

func (ae *awsEnv) patchStatus(name, status string) error {
	// update status with current nodegroup status
	_, _, err := utils.PatchStatus(ae.ctx, ae.client, ae.tenant, func(obj client.Object) client.Object {
		in := obj.(*v1.Tenants)
		if in.Status.NodegroupStatus == nil {
			in.Status.NodegroupStatus = make(map[string]string)
		}
		in.Status.NodegroupStatus[name] = status
		in.Status.Phase = v1.TenantPhase(status)
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
		klog.Infof("Namespace [%s] created for dataplane [%s]", ns.Name, ae.dp.Name)

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
				ae.tenant.Spec.Isolation.Network.AllowedNamespaces,
				resources.MakeOwnerRef(ae.tenant.APIVersion, ae.tenant.Kind, ae.tenant.Name, ae.tenant.UID),
			), metav1.CreateOptions{},
		)
		if err != nil {
			return err
		}
		klog.Infof("Network Policy [%s] created for tenant [%s]", ns.Name, ae.tenant.Name)
	}

	return nil
}
