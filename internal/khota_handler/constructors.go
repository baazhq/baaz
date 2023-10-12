package khota_handler

import (
	v1 "datainfra.io/baaz/api/v1/types"
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

func makeAwsEksConfig(dataPlaneName string, dataplane v1.DataPlane, labels map[string]string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "datainfra.io/v1",
			"kind":       "DataPlanes",
			"metadata": map[string]interface{}{
				"name":   dataPlaneName,
				"labels": labels,
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

func makeTenantConfig(tenant v1.HTTPTenant, dataplaneName string) *unstructured.Unstructured {
	var isolationEnabled, networkSecurityEnabled bool
	var allowedNamespaces []string

	if tenant.Type == v1.Siloed {
		isolationEnabled = true
	} else if tenant.Type == v1.Pool {
		isolationEnabled = false
	}

	if tenant.NetworkSecurity.InterNamespaceTraffic == v1.Deny {
		networkSecurityEnabled = true
		if tenant.NetworkSecurity.AllowedNamespaces != nil {
			allowedNamespaces = tenant.NetworkSecurity.AllowedNamespaces
		}
	} else if tenant.NetworkSecurity.InterNamespaceTraffic == v1.Allow {
		networkSecurityEnabled = false
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "datainfra.io/v1",
			"kind":       "Tenants",
			"metadata": map[string]interface{}{
				"name": tenant.TenantName,
			},
			"spec": map[string]interface{}{
				"dataplaneName": dataplaneName,
				"isolation": map[string]interface{}{
					"machine": map[string]interface{}{
						"enabled": isolationEnabled,
					},
					"network": map[string]interface{}{
						"enabled":           networkSecurityEnabled,
						"allowedNamespaces": allowedNamespaces,
					},
				},
				"config": []map[string]interface{}{
					{
						"appType": tenant.Application.Name,
						"appSize": tenant.Application.Size,
					},
				},
				"appSizes": tenant.Sizes,
			},
		},
	}
}

func makeApplicationConfig(app v1.HTTPApplication, dataplaneName, applicationName string) *unstructured.Unstructured {

	var values []string

	if app.Values != nil {
		values = app.Values
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "datainfra.io/v1",
			"kind":       "Applications",
			"metadata": map[string]interface{}{
				"name": applicationName,
			},
			"spec": map[string]interface{}{
				"envRef": dataplaneName,
				"applications": []map[string]interface{}{
					{
						"name":  applicationName,
						"scope": app.Scope,
						"spec": map[string]interface{}{
							"chartName": app.ChartName,
							"repoName":  app.RepoName,
							"repoUrl":   app.RepoURL,
							"version":   app.Version,
							"values":    values,
						},
					},
				},
			},
		},
	}
}
