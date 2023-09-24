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
	CustomerNamespaceExists  CustomMsg = "Customer Namespace Exists"
	CustomerNamespaceSuccess CustomMsg = "Customer Namespace Created"
	CustomerNamespaceFail    CustomMsg = "Customer Namespace Failed"
)

// DataPlane
const (
	DataPlaneCreateFail     CustomMsg = "DataPlane Creation Fail"
	DataPlaneCreateIntiated CustomMsg = "DataPlane Creation Initated"
	DataPlaneGetFail        CustomMsg = "DataPlane Get Fail"
)

// Tenant
const (
	TenantCreateFail     CustomMsg = "Tenant Creation Fail"
	TenantCreateIntiated CustomMsg = "Tenant Creation Initated"
	TenantGetFail        CustomMsg = "Tenant Get Fail"
)

// Application
const (
	ApplicationCreateFail     CustomMsg = "Application Creation Fail"
	ApplicationCreateIntiated CustomMsg = "Application Creation Initated"
	ApplicationGetFail        CustomMsg = "Application Get Fail"
	ApplicationDeleteFail     CustomMsg = "Application Delete Fail"
	ApplicationDeleteIntiated CustomMsg = "Application Delete Initated"
)
