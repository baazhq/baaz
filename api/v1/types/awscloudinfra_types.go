package v1

type AwsCloudInfraConfig struct {
	// AuthSecretRef holds the secret info which contains aws secret key & access key info
	// Secret must be in the same namespace as dataplane
	AuthSecretRef    AWSAuthSecretRef `json:"authSecretRef"`
	ProvisionNetwork bool             `json:"provisionNetwork"`
	Eks              EksConfig        `json:"eks"`
}

type AWSAuthSecretRef struct {
	SecretName    string `json:"secretName"`
	AccessKeyName string `json:"accessKeyName"`
	SecretKeyName string `json:"secretKeyName"`
}

type EksConfig struct {
	Name             string   `json:"name"`
	SubnetIds        []string `json:"subnetIds"`
	SecurityGroupIds []string `json:"securityGroupIds"`
	Version          string   `json:"version"`
}

type AwsCloudInfraConfigStatus struct {
	EksStatus EksStatus `json:"eksStatus,omitempty"`
}

type EksStatus struct {
	ClusterId       string `json:"clusterId,omitempty"`
	OIDCProviderArn string `json:"OIDCProviderArn,omitempty"`
}
