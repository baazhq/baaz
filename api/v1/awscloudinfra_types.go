package v1

type AwsCloudInfraConfig struct {
	Auth      AwsAuthentication `json:"auth"`
	AwsRegion string            `json:"awsRegion"`
	Eks       EksConfig         `json:"eks"`
}

type AwsAuthentication struct {
	AwsAccessKey       string `json:"awsAccessKey"`
	AwsSecretAccessKey string `json:"awsSecretAccessKey"`
}

type EksConfig struct {
	Name             string   `json:"name"`
	SubnetIds        []string `json:"subnetIds"`
	SecurityGroupIds []string `json:"securityGroupIds"`
	RoleArn          string   `json:"roleArn"`
	Version          string   `json:"version"`
}

type AwsCloudInfraConfigStatus struct {
	EksStatus EksStatus `json:"eksStatus"`
}

type EksStatus struct {
	ClusterId string `json:"clusterId"`
}
