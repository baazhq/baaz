package deployer

import (
	"context"

	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/deployer/applications"
	"datainfra.io/ballastdata/pkg/resources"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Deploy interface {
	ReconcileDeployer() error
}

type Deployer struct {
	RestConfig *rest.Config
	Env        *v1.Environment
	App        *v1.Application
}

func NewDeployer(
	restConfig *rest.Config,
	env *v1.Environment,
	app *v1.Application,
) Deploy {
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

		if err := application.ReconcileApplication(); err != nil {
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
