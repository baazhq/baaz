package khota_handler

import (
	v1 "datainfra.io/ballastdata/api/v1/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	access_key string = "accessKey"
	secret_key string = "secretKey"
)

// apiVersion: v1
// kind: Secret
// metadata:
//
//	name: aws-secret
//	namespace: shared
//
// stringData:
//
//	accessKey: AKIAWLZK4B6ACNA3H43S
//	secretKey: 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
func getAwsEksSecret(dataPlaneName string, dataplane v1.DataPlane) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name": dataPlaneName + "-aws-secret",
			},
			"stringData": map[string]interface{}{
				access_key: dataplane.CloudAuth.AwsAuth.AwsAccessKey,
				secret_key: dataplane.CloudAuth.AwsAuth.AwsSecretKey,
			},
		}}
}

func makeAwsEksConfig(dataPlaneName string, dataplane v1.DataPlane) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "datainfra.io/v1",
			"kind":       "DataPlanes",
			"metadata": map[string]interface{}{
				"name": dataPlaneName,
			},
			"spec": map[string]interface{}{
				"saasType": dataplane.SaaSType,
				"cloudInfra": map[string]interface{}{
					"cloudType": dataplane.CloudType,
					"region":    dataplane.CloudRegion,
					"authSecretRef": map[string]interface{}{
						"secretName":    dataPlaneName + "-aws-secret",
						"accessKeyName": access_key,
						"secretKeyName": secret_key,
					},
					"eks": map[string]interface{}{
						"name":             dataPlaneName,
						"subnetIds":        dataplane.KubeConfig.EKS.SubnetIds,
						"securityGroupIds": dataplane.KubeConfig.EKS.SecurityGroupIds,
						"version":          dataplane.KubeConfig.EKS.Version,
					},
				},
			},
		},
	}
}