package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppScope string

const (
	EnvironmentScope AppScope = "environment"
	TenantScope      AppScope = "tenant"
)

type ApplicationType string

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	EnvRef       string             `json:"envRef"`
	Applications map[string]AppSpec `json:"applications"`
}

type AppSpec struct {
	Scope   AppScope        `json:"scope"`
	Tenant  string          `json:"tenant,omitempty"`
	AppType ApplicationType `json:"appType,omitempty"`
	Spec    ChartSpec       `json:"spec"`
}

type ChartSpec struct {
	ChartName string   `json:"chartName"`
	RepoName  string   `json:"repoName"`
	RepoUrl   string   `json:"repoUrl"`
	Version   string   `json:"version"`
	Values    []string `json:"values,omitempty"`
}

type ApplicationStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
