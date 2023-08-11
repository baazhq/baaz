package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TenantsSpec defines the desired state of Tenants
type TenantsSpec struct {
	// Environment ref
	EnvRef string `json:"envRef"`
	// Tenant Config consists of AppType
	TenantConfig []TenantConfig `json:"config"`
	// Define Size consists of AppType
	TenantSizes []TenantSizes `json:"sizes"`
	// Isolation
	Isolation IsolationConfig `json:"isolation"`
}

type IsolationConfig struct {
	Machine MachineConfig `json:"machine"`
	Network NetworkConfig `json:"network"`
}

type MachineConfig struct {
	Enabled bool `json:"enabled"`
}

type NetworkConfig struct {
	Enabled         bool     `json:"enabled"`
	AllowNamespaces []string `json:"allowNamespaces"`
}

type TenantSizes struct {
	Name string     `json:"name"`
	Spec []NodeSpec `json:"nodes"`
}

type TenantConfig struct {
	AppType ApplicationType `json:"appType"`
	Size    string          `json:"size"`
}

// TenantsStatus defines the observed state of Tenants
type TenantsStatus struct {
	Phase           DataPlanePhase    `json:"phase,omitempty"`
	NodegroupStatus map[string]string `json:"nodegroupStatus,omitempty"`
	Namespace       map[string]string `json:"namespace,omitempty"`
}

type NodeSpec struct {
	Name       string            `json:"name"`
	NodeLabels map[string]string `json:"labels"`
	Size       string            `json:"size"`
	// +kubebuilder:validation:Minimum:=1
	Min int32 `json:"min"`
	// +kubebuilder:validation:Minimum:=1
	Max int32 `json:"max"`
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
