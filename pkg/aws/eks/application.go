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

func makeSystemNodeGroupName(tenantConfigName string) string {
	return tenantConfigName + systemNgPrefix
}

func makeTenantNodeGroupName(tenantConfigName string, appType v1.ApplicationType, ngName v1.NodeGroupName) string {
	return tenantConfigName + "-" + string(appType) + "-" + string(ngName)
}

func makeZkTenantNodeGroupName(tenantConfigName string, appType v1.ApplicationType) string {
	return tenantConfigName + "-" + string(appType) + "-" + "zk"
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
