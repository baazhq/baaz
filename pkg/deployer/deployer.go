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

<<<<<<< HEAD
func NewDeployer(restConfig *rest.Config, env *v1.Environment) Deploy {
=======
func NewDeployer(
	restConfig *rest.Config,
	env *v1.Environment,
	app *v1.Application,
) Deploy {
>>>>>>> e3e18e0 (Merge pull request #4 from datainfrahq/implementation)
	return &Deployer{
		RestConfig: restConfig,
		Env:        env,
		App:        app,
	}
}

// Deployer is responsible for deploying apps
func (deploy *Deployer) ReconcileDeployer() error {

	for appName, app := range deploy.App.Spec.Applications {

		application := applications.NewApplication(
			deploy.RestConfig,
			app,
			appName,
		)

<<<<<<< HEAD
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
			zk := applications.NewZookeeper(
				deploy.RestConfig,
				tenant,
				makeNamespace(tenant.Name, tenant.AppType),
				v1.CloudType(deploy.Env.Spec.CloudInfra.Type),
			)

			if err := zk.ReconcileZookeeper(); err != nil {
				return err
			}

			druid := applications.NewDruidz(
				deploy.RestConfig,
				tenant,
				makeNamespace(tenant.Name, tenant.AppType),
				v1.CloudType(deploy.Env.Spec.CloudInfra.Type),
			)

			if err := druid.ReconcileDruid(); err != nil {
				return err
			}
		}

		err := createNetworkPolicyPerTenant(*deploy.RestConfig, deploy.Env, makeNamespace(tenant.Name, tenant.AppType))
		if err != nil {
=======
		if err := application.ReconcileApplication(); err != nil {
>>>>>>> e3e18e0 (Merge pull request #4 from datainfrahq/implementation)
			return err
		}

		// err := createNetworkPolicyPerTenant(*deploy.RestConfig, deploy.Env, makeNamespace(tenantConfig.Name, tenantConfig.AppType))
		// if err != nil {
		// 	return err

		// }

	}

	return nil
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

const (
	eksNodeGroupSelector = "eks\\.amazonaws\\.com/nodegroup"
)

func getNodeSelector(cloudType v1.CloudType) string {
	switch cloudType {
	case v1.CloudType(v1.AWS):
		return eksNodeGroupSelector
	default:
		return ""
	}
}

func makeZkCrNameSelectorName(tenantConfigName string, appType v1.ApplicationType) string {
	return tenantConfigName + "-" + string(appType) + "-" + "zk"
}
