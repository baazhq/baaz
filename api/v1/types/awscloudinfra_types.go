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
	Vpc                string    `json:"vpc,omitempty"`
	SubnetIds          []string  `json:"subnetIds,omitempty"`
	SecurityGroupIds   []string  `json:"securityGroupIds,omitempty"`
	NATGatewayId       string    `json:"natGatewayId,omitempty"`
	NATAttachedWithRT  bool      `json:"natAttchedWithRT,omitempty"`
	SGInboundRuleAdded bool      `json:"sgInboundRuleAdded,omitempty"`
	InternetGatewayId  string    `json:"internetGatewayId,omitempty"`
	PublicRTId         string    `json:"publicRTId,omitempty"`
	EksStatus          EksStatus `json:"eksStatus,omitempty"`
}

type EksStatus struct {
	ClusterId       string `json:"clusterId,omitempty"`
	OIDCProviderArn string `json:"OIDCProviderArn,omitempty"`
}
