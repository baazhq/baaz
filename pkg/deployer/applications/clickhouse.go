package applications

import (
	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/helm"
	"k8s.io/client-go/rest"
)

const (
	chOperatorReleaseName string = "clickhouse-operator"
	chOperatorNamespace   string = "kube-system"
	chOperatorChartName   string = "clickhouse-operator"
	chOperatorRepoName    string = "datainfra"
	chOperatorRepoUrl     string = "https://charts.datainfra.io"
)

type Clickhouse interface {
	ReconcileClickhouse() error
}

type Ch struct {
	RestConfig   *rest.Config
	TenantConfig v1.TenantConfig
	Namespace    string
	Cloud        v1.CloudType
}

func NewCh(
	restConfig *rest.Config,
	tenantConfig v1.TenantConfig,
	namespace string,
	cloud v1.CloudType,
) Clickhouse {
	return &Ch{
		RestConfig:   restConfig,
		TenantConfig: tenantConfig,
		Namespace:    namespace,
		Cloud:        cloud,
	}

}

func (ch *Ch) ReconcileClickhouse() error {
	// deploy ch operator

	err := ch.deployChOperator()
	if err != nil {
		return err
	}

	// deploy ch
	err = ch.deployChi()
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
		[]string{
			"nameOverride=" + chOperatorReleaseName,
		},
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

func (ch *Ch) deployChi() error {

	chCrNameNodeName := makeChCrNameSelectorName(ch.TenantConfig.Name, ch.TenantConfig.AppType)

	chHelm := helm.NewHelm(
		chCrNameNodeName,
		ch.Namespace,
		"clickhouse",
		"apps",
		"",
		nil,
		[]string{
			"ZkHost=" + ch.Namespace + "-zk-zookeeper-headless",
			"pod.nodeSelector." + getNodeSelector(ch.Cloud) + "=" + chCrNameNodeName,
			"pod.tolerations[0].key=application",
			"pod.tolerations[0].operator=Equal",
			"pod.tolerations[0].value=" + chCrNameNodeName,
			"pod.tolerations[0].effect=NoSchedule",
		},
		nil,
	)

	exists, err := chHelm.HelmList(ch.RestConfig)

	if !exists && err == nil {
		err = chHelm.HelmInstall(ch.RestConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeChCrNameSelectorName(tenantConfigName string, appType v1.ApplicationType) string {
	return tenantConfigName + "-" + string(appType) + "-" + "ch"
}
