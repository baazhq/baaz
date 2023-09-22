package v1

type HTTPApplication struct {
	Scope      AppScope `json:"scope"`
	TenantName string   `json:"tenant_name,omitempty"`
	ChartName  string   `json:"chart_name"`
	RepoName   string   `json:"repo_name"`
	RepoURL    string   `json:"repo_url"`
	Version    string   `json:"version"`
	Values     []string `json:"values,omitempty"`
}
