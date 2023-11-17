package v1

type TenantDeploymentType string

const (
	Siloed TenantDeploymentType = "siloed"
	Pool   TenantDeploymentType = "pool"
)

type HTTPTenantApplication struct {
	Name string `json:"name"`
	Size string `json:"app_size"`
}

type HTTPTenantSizes struct {
	Name        string        `json:"name"`
	MachineSpec []MachineSpec `json:"machine_pool"`
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

type HTTPTenant struct {
	Type            TenantDeploymentType  `json:"type"`
	Application     HTTPTenantApplication `json:"application"`
	NetworkSecurity NetworkSecurity       `json:"network_security,omitempty"`
}
