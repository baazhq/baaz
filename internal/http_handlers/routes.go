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
	// 	"cloud_region":"us-east-1",
	// 	"cloud_auth":{
	// 	   "secret_ref":{
	// 		  "secretName":"aws-secret",
	// 		  "accessKeyName":"accessKey",
	// 		  "secretKeyName":"secretKey"
	// 	   }
	// 	},
	// 	"kubernetes_config":{
	// 	   "eks":{
	// 		  "name":"dataplatform-eks",
	// 		  "subnetIds":[
	// 			 "subnet-01cbca574f0d8b8d8",
	// 			 "subnet-0a4d9c31739a9ac87"
	// 		  ],
	// 		  "securityGroupIds":[
	// 			 "sg-0da08285aacbdea70"
	// 		  ],
	// 		  "version":"1.25"
	// 	   }
	// 	}
	//  }
	Route{
		"CREATE DATA PLANE",
		"POST",
		"/api/v1/dataplane/{name}",
		CreateDataPlane,
	},
}
