package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TenantPhase string

const (
	PendingT     TenantPhase = "Pending"
	CreatingT    TenantPhase = "Creating"
	ActiveT      TenantPhase = "Active"
	FailedT      TenantPhase = "Failed"
	UpdatingT    TenantPhase = "Updating"
	TerminatingT TenantPhase = "Terminating"
)

// TenantsSpec defines the desired state of Tenants
type TenantsSpec struct {
	// Environment ref
	DataplaneName string `json:"dataplaneName"`
	// Tenant Config consists of AppType
	TenantConfig []TenantApplicationConfig `json:"config"`
	// Isolation
	Isolation IsolationConfig `json:"isolation,omitempty"`
}

type IsolationConfig struct {
	Machine MachineConfig `json:"machine,omitempty"`
	Network NetworkConfig `json:"network,omitempty"`
}

type MachineConfig struct {
	Enabled bool `json:"enabled,omitempty"`
}

type NetworkConfig struct {
	Enabled           bool     `json:"enabled,omitempty"`
	AllowedNamespaces []string `json:"allowedNamespaces,omitempty"`
}

type TenantApplicationConfig struct {
	AppType ApplicationType `json:"appType"`
	Size    string          `json:"appSize"`
}

// TenantsStatus defines the observed state of Tenants
type TenantsStatus struct {
	Phase           TenantPhase       `json:"phase,omitempty"`
	NodegroupStatus map[string]string `json:"machinePoolStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tenants is the Schema for the tenants API
type Tenants struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantsSpec   `json:"spec,omitempty"`
	Status TenantsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantsList contains a list of Tenants
type TenantsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenants `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenants{}, &TenantsList{})
}
