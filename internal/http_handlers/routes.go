package http_handlers

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
	// 	"description": "logging platform",
	// 	"labels": {
	// 	  "tier": "free"
	// 	}
	// }
	// Response:
	// {
	// 	"Msg": "Namespace Created for Customer",
	// 	"Status": "SUCCESS",
	// 	"StatusCode": 200,
	// 	"Err": null
	// }
	Route{
		"CREATE CUSTOMER",
		"POST",
		"/api/v1/customer/{name}",
		CreateCustomer,
	},
	// -------------------------------------- CREATE DATAPLANE  ---------------------------------------//
	// {
	// 	"cloud_type":"aws",
	//  "saas_type": "shared",
	//  "customer_name": "parseable",
	// 	"cloud_region":"us-east-1",
	// 	"cloud_auth":{
	// 	   "awsAccessKey": "AKIAWLZK4B6ACNA3H43S",
	//        "awsSecretKey": "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"
	// 	},
	// 	"kubernetes_config":{
	// 	   "eks":{
	// 		  "name":"shared",
	// 		  "subnet_ids":[
	// 			 "subnet-01cbca574f0d8b8d8",
	// 			 "subnet-0a4d9c31739a9ac87"
	// 		  ],
	// 		  "security_group_ids":[
	// 			 "sg-0da08285aacbdea70"
	// 		  ],
	// 		  "version":"1.25"
	// 	   }
	// 	}
	//  }
	Route{
		"CREATE DATA PLANE",
		"POST",
		"/api/v1/dataplane/{dataplane_name}",
		CreateDataPlane,
	},
	Route{
		"GET DATA PLANE STATUS",
		"GET",
		"/api/v1/dataplane/{dataplane_name}/customer/{customer_name}/type/{saas_type}",
		GetDataPlaneStatus,
	},
}
