package v1

type HTTPApplication struct {
	ApplicationName string   `json:"name"`
	TenantName      string   `json:"tenant_name,omitempty"`
	Namespace       string   `json:"namespace,omitempty"`
	ChartName       string   `json:"chart_name"`
	RepoName        string   `json:"repo_name"`
	RepoURL         string   `json:"repo_url"`
	Version         string   `json:"version"`
	Values          []string `json:"values,omitempty"`
}
