package v1

type HTTPApplication struct {
	ApplicationName string   `json:"application_name"`
	TenantName      string   `json:"tenant_name"`
	ChartName       string   `json:"chart_name"`
	RepoName        string   `json:"repo_name"`
	RepoURL         string   `json:"repo_url"`
	Version         string   `json:"version"`
	Values          []string `json:"values,omitempty"`
}
