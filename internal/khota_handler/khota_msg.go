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
	DataPlaneCreateFail                   CustomMsg = "DataPlane create fail"
	DataPlaneCreateIntiated               CustomMsg = "DataPlane creation initiated"
	DataPlaneGetFail                      CustomMsg = "DataPlane get fail"
	DataPlaneListFail                     CustomMsg = "DataPlane list fail"
	DataplaneDeletionInitiated            CustomMsg = "Dataplane delete intiated"
	DataplaneDeletionFailed               CustomMsg = "Dataplane delete failed"
	DataplaneDeletionFailedCustomerExists CustomMsg = "Dataplane delete failed, customer exists on dataplane"
	DataplaneAddedSuccess                 CustomMsg = "Dataplane added success"
	DataplaneRemoveSuccess                CustomMsg = "Dataplane remove success"
	DataplanePatchFail                    CustomMsg = "Dataplane patch fail"
)

// Tenant
const (
	TenantCreateFail     CustomMsg = "Tenant creation  Fail"
	TenantCreateIntiated CustomMsg = "Tenant creation Initiated"
	TenantGetFail        CustomMsg = "Tenant get Fail"
)

// TenantsInfra
const (
	TenantsInfraCreateFail                  CustomMsg = "TenantsInfra Creation Failed"
	TenantsInfraGetFail                     CustomMsg = "TenantsInfra Get Failed"
	TenantsInfraCreateSuccess               CustomMsg = "TenantsInfra Creation Success"
	TenantInfraCreateFailDataplaneNotActive CustomMsg = "TenantsInfra Creation Failed, Dataplane is not Active"
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

// constants
const (
	req_error             string = "request_error"
	internal_error        string = "internal_error"
	dataplane_not_active  string = "dataplane not active"
	duplicate_entry       string = "entry already exists"
	entry_not_exists      string = "entry doesn't exist"
	success               string = "success"
	shared_namespace      string = "shared"
	dedicated_namespace   string = "dedicated"
	active                string = "active"
	dataplane_unavailable string = "unavailable"
	label_prefix          string = "b_"
)
