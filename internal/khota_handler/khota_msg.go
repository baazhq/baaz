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
	CustomerNamespaceExists     CustomMsg = "Customer Namespace Exists"
	CustomerNamespaceSuccess    CustomMsg = "Customer Namespace Created"
	CustomerNamespaceGetFail    CustomMsg = "Customer Namespace GET Failed"
	CustomerNamespaceCreateFail CustomMsg = "Customer Namespace CREATE Failed"
	CustomerNamespaceListEmpty  CustomMsg = "Customer Namespace LIST Empty"
	CustomerNamespaceList       CustomMsg = "Customer Namespace LIST"
)

// DataPlane
const (
	DataPlaneCreateFail        CustomMsg = "DataPlane Creation Fail"
	DataPlaneCreateIntiated    CustomMsg = "DataPlane Creation Initiated"
	DataPlaneGetFail           CustomMsg = "DataPlane GET Fail"
	DataplaneDeletionInitiated CustomMsg = "Dataplane Deletion Initiated"
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
