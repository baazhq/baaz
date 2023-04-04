package v1

type AwsCloudInfraConfig struct {
	Auth       Authentication `json:"auth"`
	AwsRegion  string         `json:"awsRegion"`
	Network    NetworkConfig  `json:"network"`
	Storage    StorageConfig  `json:"storage"`
	Kubernetes Kubernetes     `json:"kubernetes"`
}

type Authentication struct {
	AwsAccessKey       string `json:"awsAccessKey"`
	AwsSecretAccessKey string `json:"awsSecretAccessKey"`
}

type NetworkConfig struct {
	VpcName   string            `json:"vpcName"`
	CidrBlock string            `json:"cidrBlock"`
	Subnets   []Subnet          `json:"subnets"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type Subnet struct {
	Name      string            `json:"name"`
	CidrBlock string            `json:"cidrBlock"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type StorageConfig struct {
}

type Kubernetes struct {
	Eks EksConfig `json:"eks,omitempty"`
}

type EksConfig struct {
	Name    string `json:"name"`
	RoleArn string `json:"roleArn"`
	Version string `json:"version"`
}
