package applications

import (
	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/helm"
	"k8s.io/client-go/rest"
)

const (
	chOperatorReleaseName string = "clickhouse-operator"
	chOperatorNamespace   string = "clickhouse-operator"
	chOperatorChartName   string = "clickhouse-operator"
	chOperatorRepoName    string = "datainfrao"
	chOperatorRepoUrl     string = "https://charts.datainfra.io"
)

type Clickhouse interface {
	ReconcileClickhouse() error
}

type Ch struct {
	RestConfig   *rest.Config
	TenantConfig v1.TenantConfig
	Namespace    string
}

func NewCh(
	restConfig *rest.Config,
	tenantConfig v1.TenantConfig,
	namespace string,
) Clickhouse {
	return &Ch{
		RestConfig:   restConfig,
		TenantConfig: tenantConfig,
		Namespace:    namespace,
	}

}

func (ch *Ch) ReconcileClickhouse() error {
	// deploy zk operator
	err := ch.deployChOperator()
	if err != nil {
		return err
	}

	return nil
}

func (ch *Ch) deployChOperator() error {

	chOperatorHelm := helm.NewHelm(
		chOperatorReleaseName,
		chOperatorNamespace,
		chOperatorChartName,
		chOperatorRepoName,
		chOperatorRepoUrl,
		nil,
		nil,
		nil,
	)

	exists, err := chOperatorHelm.HelmList(ch.RestConfig)
	if !exists && err == nil {
		err = chOperatorHelm.HelmInstall(ch.RestConfig)
		if err != nil {
			return err
		}
	}

	return nil

}

// func (zk *Zk) deployZk() error {
// 	// deploy zk operator

// 	zkCrNameNodeName := makeZkCrNameSelectorName(zk.TenantConfig.Name, zk.TenantConfig.AppType)

// 	zkOperatorHelm := helm.NewHelm(
// 		zkCrNameNodeName,
// 		zk.Namespace,
// 		zkChartName,
// 		zkRepoName,
// 		zkRepoUrl,
// 		nil,
// 		[]string{
// 			"pod.nodeSelector.eks\\.amazonaws\\.com/nodegroup=" + zkCrNameNodeName,
// 			"pod.tolerations[0].key=application",
// 			"pod.tolerations[0].operator=Equal",
// 			"pod.tolerations[0].value=" + zkCrNameNodeName,
// 			"pod.tolerations[0].effect=NoSchedule",
// 		},
// 		nil,
// 	)

// 	exists, err := zkOperatorHelm.HelmList(zk.RestConfig)

// 	if !exists && err == nil {
// 		err = zkOperatorHelm.HelmInstall(zk.RestConfig)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

func makeChCrNameSelectorName(tenantConfigName string, appType v1.ApplicationType) string {
	return tenantConfigName + "-" + string(appType) + "-" + "ch"
}
