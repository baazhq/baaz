package khota_handler

type CustomMsg string

// Server
const (
	ServerUnmarshallError CustomMsg = "Server Json Unmarshal Error"
	ServerBodyCloseError  CustomMsg = "Server Body Close Error"
	ServerReqSizeExceed   CustomMsg = "Server Req Size Exceed Error"
)

// Customer
const (
	CustomerNamespaceExists     CustomMsg = "Customer Namespace EXISTS"
	CustomerNamespaceSuccess    CustomMsg = "Customer Namespace CREATE success"
	CustomerNamespaceGetFail    CustomMsg = "Customer Namespace GET Failed"
	CustomerNamespaceUpdateFail CustomMsg = "Customer Namespace UPDATE Failed"
	CustomerNamespaceCreateFail CustomMsg = "Customer Namespace CREATE Failed"
	CustomerNamespaceListEmpty  CustomMsg = "Customer Namespace LIST Empty"
	CustomerNamespaceList       CustomMsg = "Customer Namespace LIST"
)

// DataPlane
const (
	DataPlaneCreateFail        CustomMsg = "DataPlane CREATE Fail"
	DataPlaneCreateIntiated    CustomMsg = "DataPlane CREATION Initiated"
	DataPlaneGetFail           CustomMsg = "DataPlane GET Fail"
	DataplaneDeletionInitiated CustomMsg = "Dataplane DELETE Initiated"
	DataplaneAddedSuccess      CustomMsg = "Dataplane ADDED success"
)

// Tenant
const (
	TenantCreateFail     CustomMsg = "Tenant Creation Fail"
	TenantCreateIntiated CustomMsg = "Tenant Creation Initiated"
	TenantGetFail        CustomMsg = "Tenant GET Fail"
)

// Application
const (
	ApplicationCreateFail     CustomMsg = "Application Creation Fail"
	ApplicationCreateIntiated CustomMsg = "Application Creation Initiated"
	ApplicationGetFail        CustomMsg = "Application GET Fail"
	ApplicationDeleteFail     CustomMsg = "Application Delete Fail"
	ApplicationDeleteIntiated CustomMsg = "Application Delete Initiated"
)

// Json
const (
	JsonMarshallError CustomMsg = "Json Marshall Error"
)
