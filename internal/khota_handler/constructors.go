package khota_handler

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"

	v1 "github.com/baazhq/baaz/api/v1/types"
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

	var allApplications []map[string]interface{}

	for _, app := range dataplane.ApplicationConfig {
		allApplications = append(allApplications, map[string]interface{}{
			"name":      app.ApplicationName,
			"namespace": app.Namespace,
			"spec": map[string]interface{}{
				"chartName": app.ChartName,
				"repoName":  app.RepoName,
				"repoUrl":   app.RepoURL,
				"version":   app.Version,
				"values":    app.Values,
			},
		})
	}

	if labels[v1.PrivateObjectLabelKey] != "true" {
		dataplane.CloudAuth.AwsAuthRef.SecretName = dataPlaneName + "-aws-secret"
		dataplane.CloudAuth.AwsAuthRef.AccessKeyName = access_key
		dataplane.CloudAuth.AwsAuthRef.SecretKeyName = secret_key
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "baaz.dev/v1",
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
						"secretName":    dataplane.CloudAuth.AwsAuthRef.SecretName,
						"accessKeyName": dataplane.CloudAuth.AwsAuthRef.AccessKeyName,
						"secretKeyName": dataplane.CloudAuth.AwsAuthRef.SecretKeyName,
					},
					"provisionNetwork": dataplane.ProvisionNetwork,
					"eks": map[string]interface{}{
						"name":             dataPlaneName,
						"subnetIds":        dataplane.KubeConfig.EKS.SubnetIds,
						"securityGroupIds": dataplane.KubeConfig.EKS.SecurityGroupIds,
						"version":          dataplane.KubeConfig.EKS.Version,
					},
				},
				"applications": allApplications,
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
			"apiVersion": "baaz.dev/v1",
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

func makeApplicationConfig(apps []v1.HTTPApplication, dataplaneName, tenantName, appCRName string, labels map[string]string) *unstructured.Unstructured {

	var allApplications []map[string]interface{}
	for _, app := range apps {
		allApplications = append(allApplications, map[string]interface{}{
			"name": app.ApplicationName,
			"spec": map[string]interface{}{
				"chartName": app.ChartName,
				"repoName":  app.RepoName,
				"repoUrl":   app.RepoURL,
				"version":   app.Version,
				"values":    app.Values,
			},
		})
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "baaz.dev/v1",
			"kind":       "Applications",
			"metadata": map[string]interface{}{
				"name":   appCRName,
				"labels": labels,
			},
			"spec": map[string]interface{}{
				"dataplane":    dataplaneName,
				"tenant":       tenantName,
				"applications": allApplications,
			},
		},
	}
}

// type TenantsInfraSpec struct {
// 	Dataplane   string                 `json:"dataplane"`
// 	TenantSizes map[string]TenantSizes `json:"tenantSizes"`
// }

// type TenantSizes struct {
// 	MachineSpec []MachineSpec `json:"machinePool"`
// }

func makeTenantsInfra(dataplaneName string, tenantSizes map[string]v1.HTTPTenantSizes, labels map[string]string) *unstructured.Unstructured {

	allTenantSizes := make(map[string]interface{})
	for tName, tenantSize := range tenantSizes {
		allTenantSizes[tName] = map[string][]v1.MachineSpec{
			"machinePool": tenantSize.MachineSpec,
		}
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "baaz.dev/v1",
			"kind":       "TenantsInfra",
			"metadata": map[string]interface{}{
				"name":   dataplaneName + "-" + "tenantinfra",
				"labels": labels,
			},
			"spec": map[string]interface{}{
				"dataplane":   dataplaneName,
				"tenantSizes": allTenantSizes,
			},
		},
	}
}

func createSaToken(
	clientSet *kubernetes.Clientset,
	customerName string,
) error {
	sa, err := clientSet.CoreV1().ServiceAccounts(customerName).Create(context.TODO(), &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customerName,
			Namespace: customerName,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	_, err = clientSet.CoreV1().Secrets(customerName).Create(
		context.TODO(),
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      sa.GetName(),
				Namespace: sa.GetNamespace(),
				Annotations: map[string]string{
					"kubernetes.io/service-account.name": sa.GetName(),
				},
			},
			Type: corev1.SecretTypeServiceAccountToken,
		}, metav1.CreateOptions{})

	if err != nil {
		return err
	}

	return nil
}
