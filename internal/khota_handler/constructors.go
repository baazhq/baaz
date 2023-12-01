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
//	accessKey: kjbdsfkjbsdf
//	secretKey: lknasflbnafslkbnaflkbadf
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

func makeTenantConfig(
	tenantName string,
	tenant v1.HTTPTenant,
	dataplaneName string,
	labels map[string]string) *unstructured.Unstructured {
	var networkSecurityEnabled bool
	var allowedNamespaces []string

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
				"name":   tenantName,
				"labels": labels,
			},
			"spec": map[string]interface{}{
				"dataplaneName": dataplaneName,
				"isolation": map[string]interface{}{
					"machine": map[string]interface{}{
						"enabled": false,
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

func makeTenantsInfra(dataplaneName string, tenantSizes *[]v1.HTTPTenantSizes) *unstructured.Unstructured {

	var allTenantSizes []map[string]interface{}
	for _, tenantSize := range *tenantSizes {
		allTenantSizes = append(allTenantSizes, map[string]interface{}{
			"name":        tenantSize.Name,
			"machinePool": tenantSize.MachineSpec,
		})
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "datainfra.io/v1",
			"kind":       "TenantsInfra",
			"metadata": map[string]interface{}{
				"name": dataplaneName + "-" + "tenantinfra",
				"labels": map[string]string{
					"dataplane_name": dataplaneName,
				},
			},
			"spec": map[string]interface{}{
				"dataplane":   dataplaneName,
				"tenantSizes": allTenantSizes,
			},
		},
	}
}
