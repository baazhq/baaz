package v1

type SaaSTypes string

const (
	SharedSaaS    SaaSTypes = "Shared-SaaS"
	DedicatedSaaS SaaSTypes = "Dedicated-SaaS"
	PrivateSaaS   SaaSTypes = "Private-SaaS"
)

type Customer struct {
	SaaSType    SaaSTypes         `json:"saas_type"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels"`
}

type DataPlane struct {
	CloudType   CloudType        `json:"cloud_type"`
	SaaSType    SaaSTypes        `json:"saas_type"`
	CloudRegion string           `json:"cloud_region"`
	CloudAuth   CloudAuth        `json:"cloud_auth"`
	KubeConfig  KubernetesConfig `json:"kubernetes_config"`
}

type CloudAuth struct {
	AwsAccessKey string `json:"awsAccessKey"`
	AwsSecretKey string `json:"awsSecretKey"`
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
