package v1

type HTTPTenantSizes struct {
	Name        string        `json:"name"`
	MachineSpec []MachineSpec `json:"machine_pool"`
}
