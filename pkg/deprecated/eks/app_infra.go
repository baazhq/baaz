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

func makeTenantNodeGroupName(tenantConfigName string, appType v1.ApplicationType, appSize string, ngName string) string {
	return tenantConfigName + "-" + string(appType) + "-" + appSize + "-" + string(ngName)
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
