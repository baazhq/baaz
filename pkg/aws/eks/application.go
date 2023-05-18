package eks

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

const (
	application           string = "application"
	systemNgPrefix        string = "-system"
	chiNgPrefix           string = "-chi-ng"
	chiZkNgPrefix         string = "-chi-zk-ng"
	druidDataNodePrefix   string = "-druid-datanode-ng"
	druidMasterNodePrefix string = "-druid-masternode-ng"
	druidQueryNodePrefix  string = "-druid-querynode-ng"
)

func makeSystemNodeGroupName(appConfigName string) string { return appConfigName + systemNgPrefix }

// Clickhouse
func makeChiNodeGroupName(appConfigName string) string {
	return appConfigName + chiNgPrefix
}
func makeZkChiNodeGroupName(appConfigName string) string {
	return appConfigName + chiZkNgPrefix
}

// Druid
func makeDruidDataNodeGroupName(appConfigName string) string {
	return appConfigName + druidDataNodePrefix
}
func makeDruidMasterNodeNodeGroupName(appConfigName string) string {
	return appConfigName + druidMasterNodePrefix
}
func makeDruidQueryNodeNodeGroupName(appConfigName string) string {
	return appConfigName + druidQueryNodePrefix
}

// NewTaints constructs taints for nodes specific to application type.
func makeTaints(value string) *[]types.Taint {
	return &[]types.Taint{
		{
			Effect: types.TaintEffectNoSchedule,
			Key:    aws.String(application),
			Value:  aws.String(value),
		},
	}
}
