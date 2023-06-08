package deployer

import (
	"context"
	"fmt"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/deployer/applications"
	"datainfra.io/ballastdata/pkg/resources"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// versions

const (
	zkControlPlaneVersion string = "0.2.15"
)

type Deploy interface {
	ReconcileDeployer() error
}

type Deployer struct {
	RestConfig *rest.Config
	Env        *v1.Environment
}

func NewDeployer(restConfig *rest.Config, env *v1.Environment) Deploy {
	return &Deployer{
		RestConfig: restConfig,
		Env:        env,
	}
}

// Deployer is responsible for deploying apps
func (deploy *Deployer) ReconcileDeployer() error {

	for _, tenant := range deploy.Env.Spec.Tenant {

		switch tenant.AppType {

		case v1.ClickHouse:

			zk := applications.NewZookeeper(
				deploy.RestConfig,
				tenant,
				makeNamespace(tenant.Name, tenant.AppType),
				v1.CloudType(deploy.Env.Spec.CloudInfra.Type),
			)

			if err := zk.ReconcileZookeeper(); err != nil {
				fmt.Println(err)
				return err
			}

			ch := applications.NewCh(
				deploy.RestConfig,
				tenant,
				makeNamespace(tenant.Name, tenant.AppType),
				v1.CloudType(deploy.Env.Spec.CloudInfra.Type),
			)

			if err := ch.ReconcileClickhouse(); err != nil {
				return err
			}

		case v1.Druid:

		}

		err := createNetworkPolicyPerTenant(*deploy.RestConfig, deploy.Env, makeNamespace(tenant.Name, tenant.AppType))
		if err != nil {
			return err
		}

	}

	return nil
}

func makeNamespace(tenantConfigName string, appType v1.ApplicationType) string {
	return tenantConfigName + "-" + string(appType)
}

func createNetworkPolicyPerTenant(restConfig rest.Config, env *v1.Environment, namespace string) error {
	getOwnerRef := resources.MakeOwnerRef(env.APIVersion, env.Kind, env.Name, env.UID)

	np := resources.MakeNetworkPolicy(namespace+"-network-policy", namespace, getOwnerRef)

	clientset, err := kubernetes.NewForConfig(&restConfig)
	if err != nil {
		return nil
	}

	_, err = clientset.NetworkingV1().NetworkPolicies(namespace).Get(context.TODO(), namespace+"-network-policy", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		np, err := clientset.NetworkingV1().NetworkPolicies(namespace).Create(context.TODO(), np, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		klog.Infof("Created Network Policy [%s] in namespace [%s]", np.GetName(), np.GetNamespace())
	}

	return nil
}
