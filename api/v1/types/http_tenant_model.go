package v1

type TenantDeploymentType string

const (
	Siloed TenantDeploymentType = "siloed"
	Pool   TenantDeploymentType = "pool"
)

type HTTPTenantApplication struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

type HTTPTenantSizes struct {
	Name  string     `json:"name"`
	Nodes []NodeSpec `json:"nodes"`
}

type NetworkRules string

const (
	Allow NetworkRules = "Allow"
	Deny  NetworkRules = "Deny"
)

type NetworkSecurity struct {
	InterNamespaceTraffic NetworkRules `json:"inter_namespace_traffic"`
	AllowedNamespaces     []string     `json:"allowed_namespaces"`
}

type Tenant struct {
	TenantName      string                `json:"name"`
	Type            TenantDeploymentType  `json:"type"`
	Application     HTTPTenantApplication `json:"application"`
	Sizes           HTTPTenantSizes       `json:"sizes"`
	NetworkSecurity NetworkSecurity       `json:"network_security,omitempty"`
}
