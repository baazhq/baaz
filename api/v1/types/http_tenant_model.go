package v1

type TenantDeploymentType string

const (
	Siloed TenantDeploymentType = "SILOED"
	Pool   TenantDeploymentType = "POOL"
)

type HTTPTenantApplication struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

type HTTPTenantSizes struct {
	Name  string     `json:"name"`
	Nodes []NodeSpec `json:"nodes"`
}

type NetworkSecurity struct {
	Network NetworkConfig `json:"network,omitempty"`
}

type Tenant struct {
	TenantName      string                `json:"name"`
	Type            TenantDeploymentType  `json:"tenant_type"`
	DataplaneName   string                `json:"dataplane_name"`
	Application     HTTPTenantApplication `json:"application"`
	Sizes           HTTPTenantSizes       `json:"sizes"`
	NetworkSecurity NetworkSecurity       `json:"network_security,omitempty"`
}
