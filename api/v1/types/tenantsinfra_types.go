package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type TenantsInfraSpec struct {
	Dataplane   string                 `json:"dataplane"`
	TenantSizes map[string]TenantSizes `json:"tenantSizes"`
}

type TenantSizes struct {
	MachineSpec []MachineSpec `json:"machinePool"`
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

// TenantsStatus defines the observed state of Tenants
type TenantsInfraStatus struct {
	Phase           TenantPhase       `json:"phase,omitempty"`
	NodegroupStatus map[string]string `json:"machinePoolStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TenantsInfra is the Schema for the tenants API
type TenantsInfra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantsInfraSpec   `json:"spec,omitempty"`
	Status TenantsInfraStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantsInfraList contains a list of Tenants
type TenantsInfraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TenantsInfra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TenantsInfra{}, &TenantsInfraList{})
}
