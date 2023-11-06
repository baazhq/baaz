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
	CustomerNamespaceExists        CustomMsg = "Customer Namespace exists"
	CustomerNamespaceDoesNotExists CustomMsg = "Customer Namespace doesn't EXISTS"
	CustomerNamespaceSuccess       CustomMsg = "Customer Namespace create success"
	CustomerNamespaceGetFail       CustomMsg = "Customer Namespace get Failed"
	CustomerNamespaceUpdateSuccess CustomMsg = "Customer Namespace update success"
	CustomerNamespaceUpdateFail    CustomMsg = "Customer Namespace update Failed"
	CustomerNamespaceCreateFail    CustomMsg = "Customer Namespace create Failed"
	CustomerNamespaceListEmpty     CustomMsg = "Customer Namespace list Empty"
	CustomerNamespaceList          CustomMsg = "Customer Namespace list"
)

// DataPlane
const (
	DataPlaneCreateFail        CustomMsg = "DataPlane create Fail"
	DataPlaneCreateIntiated    CustomMsg = "DataPlane creation Initiated"
	DataPlaneGetFail           CustomMsg = "DataPlane get Fail"
	DataPlaneListFail          CustomMsg = "DataPlane list Fail"
	DataplaneDeletionInitiated CustomMsg = "Dataplane delete Initiated"
	DataplaneAddedSuccess      CustomMsg = "Dataplane added success"
	DataplaneRemoveSuccess     CustomMsg = "Dataplane remove success"
	DataplanePatchFail         CustomMsg = "Dataplane patch Fail"
)

// Tenant
const (
	TenantCreateFail     CustomMsg = "Tenant creation  Fail"
	TenantCreateIntiated CustomMsg = "Tenant creation Initiated"
	TenantGetFail        CustomMsg = "Tenant get Fail"
)

// TenantSizes
const (
	TenantSizeCreateFail    CustomMsg = "Tenant Size Creation Failed"
	TenantSizeGetFail       CustomMsg = "Tenant Size Get Failed"
	TenantSizeCreateSuccess CustomMsg = "Tenant Size Creation Success"
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
