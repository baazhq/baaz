package http_handlers

import (
	v1 "datainfra.io/ballastdata/api/v1/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getAwsEksConfig(dataPlaneName string, dataplane v1.DataPlane) *unstructured.Unstructured {
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
						"secretName":    dataplane.CloudAuth.SecretRef.SecretName,
						"accessKeyName": dataplane.CloudAuth.SecretRef.AccessKeyName,
						"secretKeyName": dataplane.CloudAuth.SecretRef.SecretKeyName,
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
