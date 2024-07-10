package khota_handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Route object
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is a slice of Route
type Routes []Route

// NewRouter returns a new router
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {

		var handler http.Handler
		handler = route.HandlerFunc

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

var routes = Routes{
	// -------------------------------------- CUSTOMER ROUTES ---------------------------------------//
	// Request:
	// {
	// 	"saas_type": "SHARED",
	//  "cloud_type": "AWS",
	// 	"labels": {
	// 		"tier": "free",
	// 		"app": "logging"
	// 	}
	// }
	// Response:
	// {
	// 	"Msg": "Customer Namespace Created",
	// 	"Status": "SUCCESS",
	// 	"StatusCode": 200,
	// 	"Err": null
	// }
	Route{
		"CREATE CUSTOMER",
		"POST",
		"/api/v1/customer/{customer_name}",
		CreateCustomer,
	},
	Route{
		"UPDATE CUSTOMER",
		"PUT",
		"/api/v1/customer/{customer_name}",
		UpdateCustomer,
	},
	Route{
		"LIST CUSTOMERS",
		"GET",
		"/api/v1/customer",
		ListCustomer,
	},
	// -------------------------------------- DATAPLANE ROUTES ---------------------------------------//
	// {
	// 	"cloud_type": "aws",
	// 	"cloud_region": "us-east-1",
	// 	"cloud_auth": {
	// 		"aws_auth": {
	// 			"aws_access_key": "gandasa",
	// 			"aws_secret_key": "tunibolditerayaarbolda"
	// 		}
	// 	},
	// 	"kubernetes_config": {
	// 		"eks": {
	// 			"security_group_ids": [
	// 				"sg-0da08285aacbdea70"
	// 			],
	// 			"subnet_ids": [
	// 				"subnet-01cbca574f0d8b8d8",
	// 				"subnet-0a4d9c31739a9ac87"
	// 			],
	// 			"version": "1.27"
	// 		}
	// 	}
	// }
	Route{
		"CREATE DATA PLANE",
		"POST",
		"/api/v1/dataplane",
		CreateDataPlane,
	},
	Route{
		"UPDATE DATA PLANE",
		"PUT",
		"/api/v1/dataplane/{dataplane_name}",
		UpdateDataPlane,
	},
	Route{
		"ADD DATA PLANE",
		"PUT",
		"/api/v1/dataplane/{dataplane_name}/customer/{customer_name}",
		AddRemoveDataPlane,
	},
	Route{
		"GET DATA PLANE STATUS",
		"GET",
		"/api/v1/dataplane/{dataplane_name}",
		GetDataPlaneStatus,
	},
	Route{
		"DELETE DATA PLANE",
		"DELETE",
		"/api/v1/dataplane/{dataplane_name}",
		DeleteDataPlane,
	},
	Route{
		"LIST ALL DATA PLANE",
		"GET",
		"/api/v1/dataplane",
		ListDataPlane,
	},
	// -------------------------------------- TENANT ROUTES ---------------------------------------//
	Route{
		"CREATE TENANT",
		"POST",
		"/api/v1/customer/{customer_name}/tenant/{tenant_name}",
		CreateTenant,
	},
	Route{
		"GET TENANTS",
		"GET",
		"/api/v1/customer/{customer_name}/tenant",
		GetAllTenantInCustomer,
	},
	// Route{
	// 	"LIST TENANT",
	// 	"GET",
	// 	"/api/v1/tenant",
	// 	GetTenantStatus,
	// },
	// --------------------------------------- TENANT INFRA ---------------------------------------------//
	Route{
		"CREATE TENANT INFRA",
		"POST",
		"/api/v1/dataplane/{dataplane_name}/tenantsinfra",
		CreateTenantInfra,
	},
	Route{
		"DELETE TENANT INFRA",
		"DELETE",
		"/api/v1/dataplane/{dataplane_name}/tenantsinfra/{tenantsinfra_name}",
		DeleteTenantInfra,
	},
	Route{
		"GET TENANT SIZES",
		"GET",
		"/api/v1/dataplane/{dataplane_name}/tenantsinfra/{tenantsinfra_name}",
		GetTenantInfra,
	},
	// -------------------------------------- APPLICATIONS ROUTES ---------------------------------------//
	Route{
		"CREATE APPLICATION",
		"POST",
		"/api/v1/customer/{customer_name}/tenant/{tenant_name}/application",
		CreateApplication,
	},
	Route{
		"GET APPLICATION STATUS",
		"GET",
		"/api/v1/customer/{customer_name}/dataplane/{dataplane_name}/application/{application_name}",
		GetApplicationStatus,
	},
	Route{
		"DELETE APPLICATION",
		"DELETE",
		"/api/v1/customer/{customer_name}/dataplane/{dataplane_name}/application/{application_name}",
		DeleteApplicationStatus,
	},
	// Get Kubeconfig for a Private SaaS customer
	Route{
		"GET KUBECONFIG FOR PRIVATE SAAS CUSTOMER",
		"GET",
		"/api/v1/customer/{customer_name}/config",
		GetKubeConfig,
	},
}
