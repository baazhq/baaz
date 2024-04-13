package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplicationPhase string

const (
	PendingA      ApplicationPhase = "Pending"
	UninstallingA ApplicationPhase = "Uninstall"
	DeployedA     ApplicationPhase = "Deployed"
	InstallingA   ApplicationPhase = "Installing"
	FailedA       ApplicationPhase = "Failed"
)

type ApplicationType string

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	Dataplane    string    `json:"dataplane"`
	Tenant       string    `json:"tenant"`
	Applications []AppSpec `json:"applications"`
}

type AppSpec struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace,omitempty"`
	Spec      ChartSpec `json:"spec"`
}

type ChartSpec struct {
	ChartName string   `json:"chartName"`
	RepoName  string   `json:"repoName"`
	RepoUrl   string   `json:"repoUrl"`
	Version   string   `json:"version"`
	Values    []string `json:"values,omitempty"`
}

type ApplicationStatus struct {
	Phase                  ApplicationPhase `json:"phase,omitempty"`
	ApplicationCurrentSpec ApplicationSpec  `json:"applicationCurrentSpec,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Application is the Schema for the applications API
type Applications struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Applications `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Applications{}, &ApplicationsList{})
}
