package applications

import (
	v1 "datainfra.io/ballastdata/api/v1"
	"datainfra.io/ballastdata/pkg/helm"
	"k8s.io/client-go/rest"
)

// zookeeper operator
const (
	zkOperatorReleaseName string = "zookeeper-operator"
	zkOperatorNamespace   string = "zookeeper-operator"
	zkOperatorChartName   string = "zookeeper-operator"
	zkOperatorRepoName    string = "pravega"
	zkOperatorRepoUrl     string = "https://charts.pravega.io"
)

// zookeeper
const (
	zkReleaseName string = "zookeeper"
	zkNamespace   string = "zookeeper"
	zkChartName   string = "zookeeper"
	zkRepoName    string = "pravega"
	zkRepoUrl     string = "https://charts.pravega.io"
)

type Zookeeper interface {
	ReconcileZookeeper() error
}

type Zk struct {
	RestConfig   *rest.Config
	TenantConfig v1.TenantConfig
	Namespace    string
	Cloud        v1.CloudType
}

func NewZookeeper(
	restConfig *rest.Config,
	tenantConfig v1.TenantConfig,
	namespace string,
	cloud v1.CloudType,
) Zookeeper {
	return &Zk{
		RestConfig:   restConfig,
		TenantConfig: tenantConfig,
		Namespace:    namespace,
		Cloud:        cloud,
	}

}

func (zk *Zk) ReconcileZookeeper() error {
	// deploy zk operator
	err := zk.deployZkOperator()
	if err != nil {
		return err
	}

	err = zk.deployZk()
	if err != nil {
		return err
	}

	return nil
}

func (zk *Zk) deployZkOperator() error {
	// deploy zk operator
	zkOperatorHelm := helm.NewHelm(
		zkOperatorReleaseName,
		zkOperatorNamespace,
		zkOperatorChartName,
		zkRepoName,
		zkRepoUrl,
		nil,
		nil,
		nil)

	exists, err := zkOperatorHelm.HelmList(zk.RestConfig)
	if !exists && err == nil {
		err = zkOperatorHelm.HelmInstall(zk.RestConfig)
		if err != nil {
			return err
		}
	}

	return nil

}

func (zk *Zk) deployZk() error {
	// deploy zk operator

	zkCrNameNodeName := makeZkCrNameSelectorName(zk.TenantConfig.Name, zk.TenantConfig.AppType)

	zkOperatorHelm := helm.NewHelm(
		zkCrNameNodeName,
		zk.Namespace,
		zkChartName,
		zkRepoName,
		zkRepoUrl,
		nil,
		[]string{
			"pod.nodeSelector." + getNodeSelector(zk.Cloud) + "=" + zkCrNameNodeName,
			"pod.tolerations[0].key=application",
			"pod.tolerations[0].operator=Equal",
			"pod.tolerations[0].value=" + zkCrNameNodeName,
			"pod.tolerations[0].effect=NoSchedule",
		},
		nil,
	)

	exists, err := zkOperatorHelm.HelmList(zk.RestConfig)

	if !exists && err == nil {
		err = zkOperatorHelm.HelmInstall(zk.RestConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeZkCrNameSelectorName(tenantConfigName string, appType v1.ApplicationType) string {
	return tenantConfigName + "-" + string(appType) + "-" + "zk"
}
