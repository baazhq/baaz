package application

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

const (
	application    string = "application"
	systemNgPrefix string = "-system"
	chiNgPrefix    string = "-chi-ng"
	chiZkNgPrefix  string = "-chi-zk-ng"
)

func MakeSystemNodeGroupName(appConfigName string) string { return appConfigName + systemNgPrefix }
func MakeChiNodeGroupName(appConfigName string) string    { return appConfigName + chiNgPrefix }
func MakeZkChiNodeGroupName(appConfigName string) string  { return appConfigName + chiZkNgPrefix }

// NewTaints constructs taints for nodes specific to application type.
func MakeTaints(value string) *[]types.Taint {
	return &[]types.Taint{
		{
			Effect: types.TaintEffectNoSchedule,
			Key:    aws.String(application),
			Value:  aws.String(value),
		},
	}
}
