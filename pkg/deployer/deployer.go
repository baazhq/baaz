package deployer

import (
	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/deployer/applications"
	"k8s.io/client-go/rest"
)

// versions

const (
	zkControlPlaneVersion string = "0.2.15"
)

// druid
const (
	druidOperatorReleaseName string = "druid-operator"
	druidOperatorNamespace   string = "druid-operator"
	druidOperatorChartName   string = "druid-operator"
	druidRepoName            string = "datainfra"
	druidRepoUrl             string = "https://charts.datainfra.io"
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
			)

			if err := zk.ReconcileZookeeper(); err != nil {
				return err
			}

		case v1.Druid:

		}
	}
	return nil
}

func makeNamespace(tenantConfigName string, appType v1.ApplicationType) string {
	return tenantConfigName + "-" + string(appType)
}
