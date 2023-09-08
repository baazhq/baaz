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
	// -------------------------------------- CREATE DATAPLANE  ---------------------------------------//
	// {
	// 	"cloud_type":"AWS",
	// 	"cloud_region":"us-east-1",
	// 	"saas_type": "SHARED",
	// 	"cloud_auth":{
	// 	  "awsAuth": {
	// 		  "awsAccessKey": "adsfadfa",
	// 		  "awsSecretKey": "asfasfasfadf"
	// 	  }
	// 	},
	// 	"kubernetes_config":{
	// 	"eks":{
	// 		"subnet_ids":[
	// 			"subnet-01cbca574f0d8b8d8",
	// 			"subnet-0a4d9c31739a9ac87"
	// 		],
	// 		"security_group_ids":[
	// 			"sg-0da08285aacbdea70"
	// 		],
	// 		"version":"1.25"
	// 	}
	// 	}
	// }
	Route{
		"CREATE DATA PLANE",
		"POST",
		"/api/v1/dataplane/{customer_name}",
		CreateDataPlane,
	},
	Route{
		"GET DATA PLANE STATUS",
		"GET",
		"/api/v1/dataplane/{customer_name}",
		GetDataPlaneStatus,
	},
}
