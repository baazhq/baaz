//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AWSAuthSecretRef) DeepCopyInto(out *AWSAuthSecretRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AWSAuthSecretRef.
func (in *AWSAuthSecretRef) DeepCopy() *AWSAuthSecretRef {
	if in == nil {
		return nil
	}
	out := new(AWSAuthSecretRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSpec) DeepCopyInto(out *AppSpec) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSpec.
func (in *AppSpec) DeepCopy() *AppSpec {
	if in == nil {
		return nil
	}
	out := new(AppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationSpec) DeepCopyInto(out *ApplicationSpec) {
	*out = *in
	if in.Applications != nil {
		in, out := &in.Applications, &out.Applications
		*out = make([]AppSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationSpec.
func (in *ApplicationSpec) DeepCopy() *ApplicationSpec {
	if in == nil {
		return nil
	}
	out := new(ApplicationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationStatus) DeepCopyInto(out *ApplicationStatus) {
	*out = *in
	in.ApplicationCurrentSpec.DeepCopyInto(&out.ApplicationCurrentSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationStatus.
func (in *ApplicationStatus) DeepCopy() *ApplicationStatus {
	if in == nil {
		return nil
	}
	out := new(ApplicationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Applications) DeepCopyInto(out *Applications) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Applications.
func (in *Applications) DeepCopy() *Applications {
	if in == nil {
		return nil
	}
	out := new(Applications)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Applications) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApplicationsList) DeepCopyInto(out *ApplicationsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Applications, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApplicationsList.
func (in *ApplicationsList) DeepCopy() *ApplicationsList {
	if in == nil {
		return nil
	}
	out := new(ApplicationsList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApplicationsList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsAuth) DeepCopyInto(out *AwsAuth) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsAuth.
func (in *AwsAuth) DeepCopy() *AwsAuth {
	if in == nil {
		return nil
	}
	out := new(AwsAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsCloudInfraConfig) DeepCopyInto(out *AwsCloudInfraConfig) {
	*out = *in
	out.AuthSecretRef = in.AuthSecretRef
	in.Eks.DeepCopyInto(&out.Eks)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsCloudInfraConfig.
func (in *AwsCloudInfraConfig) DeepCopy() *AwsCloudInfraConfig {
	if in == nil {
		return nil
	}
	out := new(AwsCloudInfraConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AwsCloudInfraConfigStatus) DeepCopyInto(out *AwsCloudInfraConfigStatus) {
	*out = *in
	out.EksStatus = in.EksStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AwsCloudInfraConfigStatus.
func (in *AwsCloudInfraConfigStatus) DeepCopy() *AwsCloudInfraConfigStatus {
	if in == nil {
		return nil
	}
	out := new(AwsCloudInfraConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChartSpec) DeepCopyInto(out *ChartSpec) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChartSpec.
func (in *ChartSpec) DeepCopy() *ChartSpec {
	if in == nil {
		return nil
	}
	out := new(ChartSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudAuth) DeepCopyInto(out *CloudAuth) {
	*out = *in
	out.AwsAuth = in.AwsAuth
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudAuth.
func (in *CloudAuth) DeepCopy() *CloudAuth {
	if in == nil {
		return nil
	}
	out := new(CloudAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudInfraConfig) DeepCopyInto(out *CloudInfraConfig) {
	*out = *in
	in.AwsCloudInfraConfig.DeepCopyInto(&out.AwsCloudInfraConfig)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudInfraConfig.
func (in *CloudInfraConfig) DeepCopy() *CloudInfraConfig {
	if in == nil {
		return nil
	}
	out := new(CloudInfraConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudInfraStatus) DeepCopyInto(out *CloudInfraStatus) {
	*out = *in
	out.AwsCloudInfraConfigStatus = in.AwsCloudInfraConfigStatus
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudInfraStatus.
func (in *CloudInfraStatus) DeepCopy() *CloudInfraStatus {
	if in == nil {
		return nil
	}
	out := new(CloudInfraStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Customer) DeepCopyInto(out *Customer) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Customer.
func (in *Customer) DeepCopy() *Customer {
	if in == nil {
		return nil
	}
	out := new(Customer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlane) DeepCopyInto(out *DataPlane) {
	*out = *in
	out.CloudAuth = in.CloudAuth
	in.KubeConfig.DeepCopyInto(&out.KubeConfig)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlane.
func (in *DataPlane) DeepCopy() *DataPlane {
	if in == nil {
		return nil
	}
	out := new(DataPlane)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneCondition) DeepCopyInto(out *DataPlaneCondition) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneCondition.
func (in *DataPlaneCondition) DeepCopy() *DataPlaneCondition {
	if in == nil {
		return nil
	}
	out := new(DataPlaneCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneSpec) DeepCopyInto(out *DataPlaneSpec) {
	*out = *in
	in.CloudInfra.DeepCopyInto(&out.CloudInfra)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneSpec.
func (in *DataPlaneSpec) DeepCopy() *DataPlaneSpec {
	if in == nil {
		return nil
	}
	out := new(DataPlaneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlaneStatus) DeepCopyInto(out *DataPlaneStatus) {
	*out = *in
	out.CloudInfraStatus = in.CloudInfraStatus
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]DataPlaneCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.NodegroupStatus != nil {
		in, out := &in.NodegroupStatus, &out.NodegroupStatus
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AddonStatus != nil {
		in, out := &in.AddonStatus, &out.AddonStatus
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlaneStatus.
func (in *DataPlaneStatus) DeepCopy() *DataPlaneStatus {
	if in == nil {
		return nil
	}
	out := new(DataPlaneStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlanes) DeepCopyInto(out *DataPlanes) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlanes.
func (in *DataPlanes) DeepCopy() *DataPlanes {
	if in == nil {
		return nil
	}
	out := new(DataPlanes)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataPlanes) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DataPlanesList) DeepCopyInto(out *DataPlanesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DataPlanes, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DataPlanesList.
func (in *DataPlanesList) DeepCopy() *DataPlanesList {
	if in == nil {
		return nil
	}
	out := new(DataPlanesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DataPlanesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EKSConfig) DeepCopyInto(out *EKSConfig) {
	*out = *in
	if in.SubnetIds != nil {
		in, out := &in.SubnetIds, &out.SubnetIds
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SecurityGroupIds != nil {
		in, out := &in.SecurityGroupIds, &out.SecurityGroupIds
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EKSConfig.
func (in *EKSConfig) DeepCopy() *EKSConfig {
	if in == nil {
		return nil
	}
	out := new(EKSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EksConfig) DeepCopyInto(out *EksConfig) {
	*out = *in
	if in.SubnetIds != nil {
		in, out := &in.SubnetIds, &out.SubnetIds
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SecurityGroupIds != nil {
		in, out := &in.SecurityGroupIds, &out.SecurityGroupIds
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EksConfig.
func (in *EksConfig) DeepCopy() *EksConfig {
	if in == nil {
		return nil
	}
	out := new(EksConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EksStatus) DeepCopyInto(out *EksStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EksStatus.
func (in *EksStatus) DeepCopy() *EksStatus {
	if in == nil {
		return nil
	}
	out := new(EksStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPApplication) DeepCopyInto(out *HTTPApplication) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPApplication.
func (in *HTTPApplication) DeepCopy() *HTTPApplication {
	if in == nil {
		return nil
	}
	out := new(HTTPApplication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPTenant) DeepCopyInto(out *HTTPTenant) {
	*out = *in
	out.Application = in.Application
	in.Sizes.DeepCopyInto(&out.Sizes)
	in.NetworkSecurity.DeepCopyInto(&out.NetworkSecurity)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPTenant.
func (in *HTTPTenant) DeepCopy() *HTTPTenant {
	if in == nil {
		return nil
	}
	out := new(HTTPTenant)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPTenantApplication) DeepCopyInto(out *HTTPTenantApplication) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPTenantApplication.
func (in *HTTPTenantApplication) DeepCopy() *HTTPTenantApplication {
	if in == nil {
		return nil
	}
	out := new(HTTPTenantApplication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPTenantSizes) DeepCopyInto(out *HTTPTenantSizes) {
	*out = *in
	if in.Nodes != nil {
		in, out := &in.Nodes, &out.Nodes
		*out = make([]NodeSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPTenantSizes.
func (in *HTTPTenantSizes) DeepCopy() *HTTPTenantSizes {
	if in == nil {
		return nil
	}
	out := new(HTTPTenantSizes)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IsolationConfig) DeepCopyInto(out *IsolationConfig) {
	*out = *in
	out.Machine = in.Machine
	in.Network.DeepCopyInto(&out.Network)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IsolationConfig.
func (in *IsolationConfig) DeepCopy() *IsolationConfig {
	if in == nil {
		return nil
	}
	out := new(IsolationConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KubernetesConfig) DeepCopyInto(out *KubernetesConfig) {
	*out = *in
	in.EKS.DeepCopyInto(&out.EKS)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KubernetesConfig.
func (in *KubernetesConfig) DeepCopy() *KubernetesConfig {
	if in == nil {
		return nil
	}
	out := new(KubernetesConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MachineConfig) DeepCopyInto(out *MachineConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MachineConfig.
func (in *MachineConfig) DeepCopy() *MachineConfig {
	if in == nil {
		return nil
	}
	out := new(MachineConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkConfig) DeepCopyInto(out *NetworkConfig) {
	*out = *in
	if in.AllowedNamespaces != nil {
		in, out := &in.AllowedNamespaces, &out.AllowedNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkConfig.
func (in *NetworkConfig) DeepCopy() *NetworkConfig {
	if in == nil {
		return nil
	}
	out := new(NetworkConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NetworkSecurity) DeepCopyInto(out *NetworkSecurity) {
	*out = *in
	if in.AllowedNamespaces != nil {
		in, out := &in.AllowedNamespaces, &out.AllowedNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NetworkSecurity.
func (in *NetworkSecurity) DeepCopy() *NetworkSecurity {
	if in == nil {
		return nil
	}
	out := new(NetworkSecurity)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeSpec) DeepCopyInto(out *NodeSpec) {
	*out = *in
	if in.NodeLabels != nil {
		in, out := &in.NodeLabels, &out.NodeLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeSpec.
func (in *NodeSpec) DeepCopy() *NodeSpec {
	if in == nil {
		return nil
	}
	out := new(NodeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantConfig) DeepCopyInto(out *TenantConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantConfig.
func (in *TenantConfig) DeepCopy() *TenantConfig {
	if in == nil {
		return nil
	}
	out := new(TenantConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantSizes) DeepCopyInto(out *TenantSizes) {
	*out = *in
	if in.Spec != nil {
		in, out := &in.Spec, &out.Spec
		*out = make([]NodeSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantSizes.
func (in *TenantSizes) DeepCopy() *TenantSizes {
	if in == nil {
		return nil
	}
	out := new(TenantSizes)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tenants) DeepCopyInto(out *Tenants) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tenants.
func (in *Tenants) DeepCopy() *Tenants {
	if in == nil {
		return nil
	}
	out := new(Tenants)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Tenants) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantsList) DeepCopyInto(out *TenantsList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Tenants, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantsList.
func (in *TenantsList) DeepCopy() *TenantsList {
	if in == nil {
		return nil
	}
	out := new(TenantsList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TenantsList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantsSpec) DeepCopyInto(out *TenantsSpec) {
	*out = *in
	if in.TenantConfig != nil {
		in, out := &in.TenantConfig, &out.TenantConfig
		*out = make([]TenantConfig, len(*in))
		copy(*out, *in)
	}
	if in.TenantSizes != nil {
		in, out := &in.TenantSizes, &out.TenantSizes
		*out = make([]TenantSizes, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Isolation.DeepCopyInto(&out.Isolation)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantsSpec.
func (in *TenantsSpec) DeepCopy() *TenantsSpec {
	if in == nil {
		return nil
	}
	out := new(TenantsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TenantsStatus) DeepCopyInto(out *TenantsStatus) {
	*out = *in
	if in.NodegroupStatus != nil {
		in, out := &in.NodegroupStatus, &out.NodegroupStatus
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TenantsStatus.
func (in *TenantsStatus) DeepCopy() *TenantsStatus {
	if in == nil {
		return nil
	}
	out := new(TenantsStatus)
	in.DeepCopyInto(out)
	return out
}
