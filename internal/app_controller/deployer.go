package app_controller

import (
	"context"
	"fmt"

	"datainfra.io/ballastdata/pkg/application/helm"
	"datainfra.io/ballastdata/pkg/aws/eks"

	v1 "datainfra.io/ballastdata/api/v1/types"
	"k8s.io/client-go/kubernetes"
)

type Application struct {
	Context      context.Context
	App          *v1.Application
	Env          *v1.Environment
	K8sClientSet *kubernetes.Clientset
	EksIC        eks.Eks
}

func NewApplication(
	ctx context.Context,
	app *v1.Application,
	env *v1.Environment,
	k8sClientSet *kubernetes.Clientset,
) *Application {
	return &Application{
		Context:      ctx,
		App:          app,
		Env:          env,
		EksIC:        eks.NewEks(ctx, env),
		K8sClientSet: k8sClientSet,
	}
}

// Deployer is responsible for deploying apps
func (a *Application) ReconcileApplicationDeployer() error {

	for _, app := range a.App.Spec.Applications {
		var namespace string
		if app.Scope == v1.EnvironmentScope {
			namespace = app.Name
		} else if app.Scope == v1.TenantScope {
			namespace = app.Tenant
		}

		fmt.Println(app)

		helm := helm.NewHelm(app.Name, namespace, app.Spec.ChartName, app.Spec.RepoName, app.Spec.RepoUrl, app.Spec.Values)

		restConfig, err := a.EksIC.GetRestConfig()
		if err != nil {
			return err
		}

		exists := helm.List(restConfig)
		if exists == false {
			go func() error {
				err = helm.Apply(restConfig)
				if err != nil {
					return err
				}
				return nil
			}()

			return nil
		}

	}

	return nil
}
