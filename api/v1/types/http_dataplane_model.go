package v1

type SaaSTypes string

const (
	SharedSaaS    SaaSTypes = "SHARED"
	DedicatedSaaS SaaSTypes = "dedicated"
	PrivateSaaS   SaaSTypes = "private"
)

type Customer struct {
	SaaSType  SaaSTypes         `json:"saas_type"`
	CloudType CloudType         `json:"cloud_type"`
	Labels    map[string]string `json:"labels"`
}

type DataPlane struct {
	CloudType   CloudType        `json:"cloud_type"`
	SaaSType    SaaSTypes        `json:"saas_type"`
	CloudRegion string           `json:"cloud_region"`
	CloudAuth   CloudAuth        `json:"cloud_auth"`
	KubeConfig  KubernetesConfig `json:"kubernetes_config"`
}

type CloudAuth struct {
	AwsAuth AwsAuth `json:"aws_auth"`
}

type AwsAuth struct {
	AwsAccessKey string `json:"aws_access_key"`
	AwsSecretKey string `json:"aws_secret_key"`
}

type KubernetesConfig struct {
	EKS EKSConfig `json:"eks"`
}

type EKSConfig struct {
	Name             string   `json:"name"`
	SubnetIds        []string `json:"subnet_ids"`
	SecurityGroupIds []string `json:"security_group_ids"`
	Version          string   `json:"version"`
}
