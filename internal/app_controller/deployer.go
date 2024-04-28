package app_controller

import (
	"context"
	"errors"
	"fmt"

	v1 "github.com/baazhq/baaz/api/v1/types"
	"github.com/baazhq/baaz/pkg/aws/eks"
	"github.com/baazhq/baaz/pkg/helm"
	"github.com/baazhq/baaz/pkg/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ChartCh struct {
	Name string
	Err  error
}

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

func getChartName(app v1.AppSpec) string {
	return fmt.Sprintf("%s-%s", app.Name, app.Namespace)
}

type InstallChart struct {
	Name string
	Err  error
}

// Deployer is responsible for deploying apps
func (a *Application) ReconcileApplicationDeployer() error {

	ch := make(chan InstallChart, len(a.App.Spec.Applications))
	count := 0

	for _, app := range a.App.Spec.Applications {

		restConfig, err := a.EksIC.GetRestConfig()
		if err != nil {
			return err
		}

		helm := helm.NewHelm(app.Name, a.App.Spec.Tenant, app.Spec.ChartName, app.Spec.RepoName,
			app.Spec.RepoUrl, app.Spec.Version, restConfig, app.Spec.Values)

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
			klog.Infof("installing chart: %s", app.Name)

			count += 1
			go func(ch chan InstallChart, app v1.AppSpec) {
				c := InstallChart{
					Name: getChartName(app),
					Err:  nil,
				}
				if err := helm.Apply(restConfig); err != nil {
					c.Err = err
				}
				ch <- c
			}(ch, app)

			_, _, err = utils.PatchStatus(a.Context, a.Client, a.App, func(obj client.Object) client.Object {
				in := obj.(*v1.Applications)
				if in.Status.AppStatus == nil {
					in.Status.AppStatus = make(map[string]v1.ApplicationPhase)
				}
				in.Status.AppStatus[getChartName(app)] = v1.InstallingA
				in.Status.ApplicationCurrentSpec = a.App.Spec
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

		_, _, err := utils.PatchStatus(a.Context, a.Client, a.App, func(obj client.Object) client.Object {
			in := obj.(*v1.Applications)
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

func (a *Application) UninstallApplications() error {

	restConfig, err := a.EksIC.GetRestConfig()
	if err != nil {
		return err
	}

	count := 0
	ch := make(chan ChartCh, len(a.App.Spec.Applications))

	if _, _, err := utils.PatchStatus(a.Context, a.Client, a.App, func(obj client.Object) client.Object {
		in := obj.(*v1.Applications)
		in.Status.Phase = v1.ApplicationPhase(v1.UninstallingA)
		return in
	}); err != nil {
		return err
	}

	for _, app := range a.App.Spec.Applications {

		helm := helm.NewHelm(app.Name, a.App.Spec.Tenant, app.Spec.ChartName, app.Spec.RepoName,
			app.Spec.RepoUrl, app.Spec.Version, restConfig, app.Spec.Values)

		restConfig, err := a.EksIC.GetRestConfig()
		if err != nil {
			return err
		}

		_, exists := helm.List(restConfig)

		if exists {
			count += 1
			go func(ch chan ChartCh, app v1.AppSpec) {
				c := ChartCh{
					Name: app.Name,
				}
				if err := helm.Uninstall(restConfig); err != nil {
					c.Err = err
				}
				ch <- c
			}(ch, app)
		}
	}

	var errs []error

	for i := 0; i < count; i += 1 {
		chartCh := <-ch
		var latestState v1.ApplicationPhase
		if chartCh.Err != nil {
			klog.Errorf("uninstalling chart %s failed, reason: %s", chartCh.Name, chartCh.Err.Error())
			errs = append(errs, chartCh.Err)
			latestState = v1.FailedA
		} else {
			latestState = v1.Uninstalled
		}

		_, _, err := utils.PatchStatus(a.Context, a.Client, a.App, func(obj client.Object) client.Object {
			in := obj.(*v1.Applications)
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

	return errors.Join(errs...)
}
