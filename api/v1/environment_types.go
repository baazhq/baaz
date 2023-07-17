package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	// EnvType can be dev, stage, prod environments
	EnvType string `json:"envType"`
	// Cloud can be any pubic name ie aws, gcp, azure.
	CloudInfra CloudInfraConfig `json:"cloudInfra"`
}

type CloudType string

const (
	AWS   string = "aws"
	GCP   string = "gcp"
	AZURE string = "azure"
)

type CloudInfraConfig struct {
	// CloudType
	Type                CloudType `json:"type"`
	AwsCloudInfraConfig `json:",inline,omitempty"`
}

type EnvironmentPhase string

const (
	Pending     EnvironmentPhase = "Pending"
	Creating    EnvironmentPhase = "Creating"
	Success     EnvironmentPhase = "Success"
	Failed      EnvironmentPhase = "Failed"
	Updating    EnvironmentPhase = "Updating"
	Terminating EnvironmentPhase = "Terminating"
)

// EnvironmentStatus defines the observed state of Environment
type EnvironmentStatus struct {
	Phase              EnvironmentPhase       `json:"phase,omitempty"`
	CloudInfraStatus   CloudInfraStatus       `json:"cloudInfraStatus,omitempty"`
	ObservedGeneration int64                  `json:"observedGeneration,omitempty"`
	Conditions         []EnvironmentCondition `json:"conditions,omitempty"`
	Version            string                 `json:"version,omitempty"`
	// NodegroupStatus will contain a map of node group name & status
	// Example:
	// nodegroupStatus:
	//    druid-ng1: CREATING
	//    druid-ng2: ACTIVE
	//    druid-ng3: DELETING
	NodegroupStatus map[string]string `json:"nodegroupStatus,omitempty"`
	// AddonStatus holds a map of addon name & their current status
	// Example:
	// addonStatus:
	//    aws-ebs-csi-driver: CREATING
	//    coredns:            READY
	AddonStatus map[string]string `json:"addonStatus,omitempty"`
}

type EnvironmentConditionType string

const (
	ControlPlaneCreateInitiated EnvironmentConditionType = "ControlPlaneCreateInitiated"
	ControlPlaneCreated         EnvironmentConditionType = "ControlPlaneCreated"
	NodeGroupCreateInitiated    EnvironmentConditionType = "NodeGroupCreateInitiated"
	NodeGroupCreated            EnvironmentConditionType = "NodeGroupCreated"
	VersionUpgradeInitiated     EnvironmentConditionType = "VersionUpgradeInitiated"
	VersionUpgradeSuccessful    EnvironmentConditionType = "VersionUpgradeSuccessful"
)

// EnvironmentCondition describes the state of a deployment at a certain point.
type EnvironmentCondition struct {
	// Type of deployment condition.
	Type EnvironmentConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=DeploymentConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,6,opt,name=lastUpdateTime"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,7,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

type CloudInfraStatus struct {
	Type                      string `json:"type,omitempty"`
	AwsCloudInfraConfigStatus `json:",inline,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// Environment is the Schema for the environments API
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvironmentSpec   `json:"spec,omitempty"`
	Status EnvironmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// EnvironmentList contains a list of Environment
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
