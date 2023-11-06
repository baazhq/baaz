package v1

type TenantAppSize struct {
	AppSizes []TenantSizes `json:"app_sizes"`
}
type TenantSizes struct {
	Name        string        `json:"name"`
	MachineSpec []MachineSpec `json:"machine_pool"`
}

type MachineSpec struct {
	Name       string            `json:"name"`
	NodeLabels map[string]string `json:"labels"`
	Size       string            `json:"size"`
	// +kubebuilder:validation:Minimum:=1
	Min int32 `json:"min"`
	// +kubebuilder:validation:Minimum:=1
	Max int32 `json:"max"`
}
