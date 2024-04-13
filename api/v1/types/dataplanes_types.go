package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DataPlaneSpec defines the desired state of DataPlane
type DataPlaneSpec struct {
	// Cloud can be any pubic name ie aws, gcp, azure.
	CloudInfra   CloudInfraConfig `json:"cloudInfra"`
	Applications []AppSpec        `json:"applications"`
}

type CloudType string

const (
	AWS   CloudType = "aws"
	GCP   CloudType = "gcp"
	AZURE CloudType = "azure"
)

type CloudInfraConfig struct {
	// CloudType
	CloudType           CloudType `json:"cloudType"`
	Region              string    `json:"region"`
	AwsCloudInfraConfig `json:",inline,omitempty"`
}

type DataPlanePhase string

const (
	PendingD     DataPlanePhase = "Pending"
	CreatingD    DataPlanePhase = "Creating"
	ActiveD      DataPlanePhase = "Active"
	FailedD      DataPlanePhase = "Failed"
	UpdatingD    DataPlanePhase = "Updating"
	TerminatingD DataPlanePhase = "Terminating"
)

// DataPlaneStatus defines the observed state of DataPlane
type DataPlaneStatus struct {
	Phase              DataPlanePhase       `json:"phase,omitempty"`
	CloudInfraStatus   CloudInfraStatus     `json:"cloudInfraStatus,omitempty"`
	ObservedGeneration int64                `json:"observedGeneration,omitempty"`
	Conditions         []DataPlaneCondition `json:"conditions,omitempty"`
	Version            string               `json:"version,omitempty"`
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
	// AppStatus holds a map of app helm chart name and thier current status
	// Example:
	// appStatus:
	// 	  nginx: Deployed
	//    druid: Installing
	AppStatus map[string]ApplicationPhase `json:"appStatus,omitempty"`
}

type DataPlaneConditionType string

const (
	ControlPlaneCreateInitiated DataPlaneConditionType = "ControlPlaneCreateInitiated"
	ControlPlaneCreated         DataPlaneConditionType = "ControlPlaneCreated"
	NodeGroupCreateInitiated    DataPlaneConditionType = "NodeGroupCreateInitiated"
	NodeGroupCreated            DataPlaneConditionType = "NodeGroupCreated"
	VersionUpgradeInitiated     DataPlaneConditionType = "VersionUpgradeInitiated"
	VersionUpgradeSuccessful    DataPlaneConditionType = "VersionUpgradeSuccessful"
)

// DataPlaneCondition describes the state of a deployment at a certain point.
type DataPlaneCondition struct {
	// Type of deployment condition.
	Type DataPlaneConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=DeploymentConditionType"`
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
// DataPlane is the Schema for the DataPlanes API
type DataPlanes struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataPlaneSpec   `json:"spec,omitempty"`
	Status DataPlaneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DataPlaneList contains a list of DataPlane
type DataPlanesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataPlanes `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataPlanes{}, &DataPlanesList{})
}
