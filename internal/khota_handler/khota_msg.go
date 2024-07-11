package khota_handler

type CustomMsg string

// Server
const (
	ServerUnmarshallError CustomMsg = "Server json unmarshal error"
	ServerBodyCloseError  CustomMsg = "Server body close error"
	ServerReqSizeExceed   CustomMsg = "Server req size exceed error"
)

// Customer
const (
	CustomerNamespaceExists          CustomMsg = "Customer namespace exists"
	CustomerNamespaceDoesNotExists   CustomMsg = "Customer namespace doesn't exist"
	CustomerNamespaceSuccess         CustomMsg = "Customer namespace create success"
	CustomerNamespaceGetFail         CustomMsg = "Customer namespace get fail"
	CustomerNamespaceUpdateSuccess   CustomMsg = "Customer namespace update success"
	CustomerNamespaceUpdateFail      CustomMsg = "Customer namespace update failed"
	CustomerNamespaceCreateFail      CustomMsg = "Customer namespace create failed"
	CustomerNamespaceListEmpty       CustomMsg = "Customer namespace list empty"
	CustomerNamespaceList            CustomMsg = "Customer namespace list"
	CustomerServiceAccountCreateFail CustomMsg = "Customer service account creation failed."
	CustomerNotExistInDataplane      CustomMsg = "Customer not exist in dataplane"
	CustomerNamespaceDeleteSuccess   CustomMsg = "Customer namespace delete success"
	CustomerNamespaceDeleteFail      CustomMsg = "Customer namespace delete failed"
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
	DataplaneUpdateSuccess                CustomMsg = "Dataplane updated successfully"
	DataplaneUpdateFail                   CustomMsg = "Dataplane update fail"
)

// Tenant
const (
	TenantCreateFail                   CustomMsg = "Tenant creation  fail"
	TenantCreateIntiated               CustomMsg = "Tenant creation success"
	TenantCreateFailDataplaneNotActive CustomMsg = "Tenant creation failed, Dataplane is not Active"
	TenantGetFail                      CustomMsg = "Tenant get fail"
	TenantListFail                     CustomMsg = "Tenant list fail"
	TenantDeleteFail                   CustomMsg = "Tenant delete failed"
	TenantDeleteIntiated               CustomMsg = "Tenant deletion successfully initiated"
)

// TenantsInfra
const (
	TenantsInfraCreateFail                  CustomMsg = "tenantsinfra creation failed"
	TenantsInfraGetFail                     CustomMsg = "tenantsinfra get failed"
	TenantsInfraListFail                    CustomMsg = "tenantsinfra list failed"
	TenantsInfraDeleteFail                  CustomMsg = "tenantsinfra delete failed"
	TenantsInfraDeleteInitiated             CustomMsg = "tenantsinfra delete initiated"
	TenantsInfraCreateInitiated             CustomMsg = "tenantsinfra creation initiated"
	TenantInfraUpdateSuccess                CustomMsg = "tenantsinfra update success"
	TenantInfraCreateFailDataplaneNotActive CustomMsg = "tenantsinfra creation failed, dataplane is not active"
)

// Application
const (
	ApplicationCreateFail     CustomMsg = "Application creation fail"
	ApplicationCreateIntiated CustomMsg = "Application creation initiated"
	ApplicationGetFail        CustomMsg = "Application get fail"
	ApplicationDeleteFail     CustomMsg = "Application delete fail"
	ApplicationDeleteIntiated CustomMsg = "Application delete initiated"
	ApplicationUpdateSuccess  CustomMsg = "Application update success"
	ApplicationUpdateFail     CustomMsg = "Application update failed"
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
	customer_chartpath    string = "internal/chart/customer/"
)

// config
const (
	ConfigGetFail string = "Config get failed for customer"
)
