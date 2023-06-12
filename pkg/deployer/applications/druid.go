package applications

import (
	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/helm"
	"k8s.io/client-go/rest"
)

const (
	druidOperatorReleaseName string = "druid-operator"
	druidOperatorNamespace   string = "druid-operator"
	druidOperatorChartName   string = "druid-operator"
	druidOperatorRepoName    string = "datainfra"
	druidOperatorRepoUrl     string = "https://charts.datainfra.io"
)

type Druidz interface {
	ReconcileDruid() error
}

type Druid struct {
	RestConfig   *rest.Config
	TenantConfig v1.TenantConfig
	Namespace    string
	Cloud        v1.CloudType
}

func NewDruidz(
	restConfig *rest.Config,
	tenantConfig v1.TenantConfig,
	namespace string,
	cloud v1.CloudType,
) Druidz {
	return &Druid{
		RestConfig:   restConfig,
		TenantConfig: tenantConfig,
		Namespace:    namespace,
		Cloud:        cloud,
	}

}

func (d *Druid) ReconcileDruid() error {
	// deploy zk operator
	err := d.deployDruidOperator()
	if err != nil {
		return err
	}

	return nil

}

func (d *Druid) deployDruidOperator() error {

	druidOperatorHelm := helm.NewHelm(
		druidOperatorReleaseName,
		druidOperatorNamespace,
		druidOperatorChartName,
		druidOperatorRepoName,
		druidOperatorRepoUrl,
		nil,
		nil,
		nil)

	exists, err := druidOperatorHelm.HelmList(d.RestConfig)
	if !exists && err == nil {
		err = druidOperatorHelm.HelmInstall(d.RestConfig)
		if err != nil {
			return err
		}
	}

	return nil

}

// func (d *Druid) deployZk() error {
// 	// deploy zk operator

// 	zkCrNameNodeName := makeZkCrNameSelectorName(d.TenantConfig.Name, zk.TenantConfig.AppType)

// 	zkOperatorHelm := helm.NewHelm(
// 		zkCrNameNodeName,
// 		zk.Namespace,
// 		zkChartName,
// 		zkRepoName,
// 		zkRepoUrl,
// 		nil,
// 		[]string{
// 			"pod.nodeSelector." + getNodeSelector(d.Cloud) + "=" + zkCrNameNodeName,
// 			"pod.tolerations[0].key=application",
// 			"pod.tolerations[0].operator=Equal",
// 			"pod.tolerations[0].value=" + zkCrNameNodeName,
// 			"pod.tolerations[0].effect=NoSchedule",
// 		},
// 		nil,
// 	)

// 	exists, err := zkOperatorHelm.HelmList(d.RestConfig)

// 	if !exists && err == nil {
// 		err = zkOperatorHelm.HelmInstall(d.RestConfig)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func makeDruidCrNameSelectorName(tenantConfigName string, appType v1.ApplicationType) string {
	return tenantConfigName + "-" + string(appType) + "-" + "druid"
}
