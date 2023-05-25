package eks

import (
	v1 "datainfra.io/ballastdata/api/v1"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

const (
	application    string = "application"
	systemNgPrefix string = "-system"
)

func makeSystemNodeGroupName(appConfigName string) string { return appConfigName + systemNgPrefix }

func makeDruidNodeGroupName(appConfigName string, ngName v1.NodeGroupName) string {
	return appConfigName + "-" + "druid" + "-" + string(ngName)
}

func makeChiNodeGroupName(appConfigName string, ngName v1.NodeGroupName) string {
	return appConfigName + "-" + "clickhouse" + "-" + string(ngName)
}

func makePinotNodeGroupName(appConfigName string, ngName v1.NodeGroupName) string {
	return appConfigName + "-" + "pinot" + "-" + string(ngName)
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
