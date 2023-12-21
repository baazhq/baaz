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
	CustomerNotExistInDataplane    CustomMsg = "Customer not exist in dataplane"
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
	TenantCreateFail                   CustomMsg = "Tenant creation  fail"
	TenantCreateIntiated               CustomMsg = "Tenant creation success"
	TenantCreateFailDataplaneNotActive CustomMsg = "Tenant creation failed, Dataplane is not Active"
	TenantGetFail                      CustomMsg = "Tenant get fail"
	TenantListFail                     CustomMsg = "Tenant list fail"
)

// TenantsInfra
const (
	TenantsInfraCreateFail                  CustomMsg = "TenantsInfra creation Failed"
	TenantsInfraGetFail                     CustomMsg = "TenantsInfra get Failed"
	TenantsInfraCreateSuccess               CustomMsg = "TenantsInfra creation Success"
	TenantInfraCreateFailDataplaneNotActive CustomMsg = "TenantsInfra creation Failed, Dataplane is not Active"
)

// Application
const (
	ApplicationCreateFail     CustomMsg = "Application creation fail"
	ApplicationCreateIntiated CustomMsg = "Application creation initiated"
	ApplicationGetFail        CustomMsg = "Application get fail"
	ApplicationDeleteFail     CustomMsg = "Application delete fail"
	ApplicationDeleteIntiated CustomMsg = "Application delete initiated"
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
