package applications

import v1 "datainfra.io/ballastdata/api/v1"

const (
	eksNodeGroupSelector = "eks\\.amazonaws\\.com/nodegroup"
)

func getNodeSelector(cloudType v1.CloudType) string {
	switch cloudType {
	case v1.CloudType(v1.AWS):
		return eksNodeGroupSelector
	default:
		return ""
	}
}
