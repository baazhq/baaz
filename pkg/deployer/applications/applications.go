package applications

import (
	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/helm"
	"k8s.io/client-go/rest"
)

type Application struct {
	RestConfig *rest.Config
	App        v1.AppSpec
	AppName    string
}

func NewApplication(
	restConfig *rest.Config,
	app v1.AppSpec,
	appName string,
) App {
	return &Application{
		RestConfig: restConfig,
		App:        app,
		AppName:    appName,
	}
}

func (app *Application) ReconcileApplication() error {

	err := app.deployApplication()
	if err != nil {
		return err
	}

	return nil

}

func (app *Application) deployApplication() error {

	var appNamespace string

	if app.App.Scope == v1.EnvironmentScope {
		appNamespace = app.AppName
	} else if app.App.Scope == v1.TenantScope {
		appNamespace = app.App.Tenant
	}

	application := helm.NewHelm(
		app.AppName,
		appNamespace,
		app.App.Spec.ChartName,
		app.App.Spec.RepoName,
		app.App.Spec.RepoUrl,
		nil,
		app.App.Spec.Values,
		nil)

	exists, err := application.HelmList(app.RestConfig)
	if !exists && err == nil {
		go func() error {
			err = application.HelmInstall(app.RestConfig)
			if err != nil {
				return err
			}
			return nil
		}()

	}

	return nil

}

func makeNamespace(tenantName string, appType v1.ApplicationType) string {
	return tenantName + "-" + string(appType)
}
