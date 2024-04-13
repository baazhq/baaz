package app_controller

import (
	"context"
	"errors"

	v1 "github.com/baazhq/baaz/api/v1/types"
	"github.com/baazhq/baaz/pkg/aws/eks"
	"github.com/baazhq/baaz/pkg/helm"
	"github.com/baazhq/baaz/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Application struct {
	Context      context.Context
	App          *v1.Applications
	DataPlanes   *v1.DataPlanes
	K8sClientSet *kubernetes.Clientset
	Client       client.Client
	EksIC        eks.Eks
}

func NewApplication(
	ctx context.Context,
	app *v1.Applications,
	dp *v1.DataPlanes,
	client client.Client,
	k8sClientSet *kubernetes.Clientset,
) *Application {
	return &Application{
		Context:      ctx,
		App:          app,
		DataPlanes:   dp,
		EksIC:        eks.NewEks(ctx, dp),
		Client:       client,
		K8sClientSet: k8sClientSet,
	}
}

// Deployer is responsible for deploying apps
func (a *Application) ReconcileApplicationDeployer() error {

	ch := make(chan error, len(a.App.Spec.Applications))
	count := 0

	for _, app := range a.App.Spec.Applications {

		helm := helm.NewHelm(app.Name, a.App.Spec.Tenant, app.Spec.ChartName, app.Spec.RepoName, app.Spec.RepoUrl, app.Spec.Values)

		restConfig, err := a.EksIC.GetRestConfig()
		if err != nil {
			return err
		}

		result, exists := helm.List(restConfig)
		if exists {
			for _, current := range a.App.Status.ApplicationCurrentSpec.Applications {
				if current.Spec.Version != app.Spec.Version {
					klog.Infof("Initating upgrade for application [%s], current version [%s], desired version [%s]", app.Spec.ChartName, current.Spec.Version, app.Spec.Version)
					err = helm.Upgrade(restConfig)
					if err != nil {
						return err
					}
					if _, _, err := utils.PatchStatus(a.Context, a.Client, a.App, func(obj client.Object) client.Object {
						in := obj.(*v1.Applications)
						in.Status.Phase = v1.ApplicationPhase(result)
						in.Status.ApplicationCurrentSpec = a.App.Spec
						return in
					}); err != nil {
						return err
					}
				}
			}
		}

		if !exists {
			klog.Infof("installing application: %s", app.Name)
			count += 1
			go func(ch chan error) {
				if err := helm.Apply(restConfig); err != nil {
					ch <- err
				}
				ch <- nil
			}(ch)
			// if _, _, err := utils.PatchStatus(a.Context, a.Client, a.App, func(obj client.Object) client.Object {
			// 	in := obj.(*v1.Applications)
			// 	in.Status.Phase = v1.ApplicationPhase(result)
			// 	in.Status.ApplicationCurrentSpec = a.App.Spec
			// 	return in
			// }); err != nil {
			// 	return err
			// }
		}

	}
	var errs []error
	for i := 0; i < count; i += 1 {
		if err := <-ch; err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (a *Application) UninstallApplications() error {

	for _, app := range a.App.Spec.Applications {

		helm := helm.NewHelm(app.Name, a.App.Spec.Tenant, app.Spec.ChartName, app.Spec.RepoName, app.Spec.RepoUrl, app.Spec.Values)

		restConfig, err := a.EksIC.GetRestConfig()
		if err != nil {
			return err
		}

		err = helm.Uninstall(restConfig)
		if _, _, err := utils.PatchStatus(a.Context, a.Client, a.App, func(obj client.Object) client.Object {
			in := obj.(*v1.Applications)
			in.Status.Phase = v1.ApplicationPhase(v1.UninstallingA)
			in.Status.ApplicationCurrentSpec = a.App.Spec
			return in
		}); err != nil {
			return err
		}
	}

	return nil
}
