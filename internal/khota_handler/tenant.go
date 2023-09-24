package khota_handler

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	v1 "datainfra.io/baaz/api/v1/types"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var tenantGVK = schema.GroupVersionResource{
	Group:    "datainfra.io",
	Version:  "v1",
	Resource: "tenants",
}

func CreateTenant(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]
	dataplaneName := vars["dataplane_name"]
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		res := NewResponse(ServerReqSizeExceed, req_error, err, http.StatusBadRequest)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if err := req.Body.Close(); err != nil {
		res := NewResponse(ServerBodyCloseError, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	var tenant v1.HTTPTenant

	if err := json.Unmarshal(body, &tenant); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	var sizes []v1.HTTPTenantSizes

	for _, size := range tenant.Sizes {
		sizes = append(sizes, size)
	}

	tenantNew := v1.HTTPTenant{
		TenantName: tenant.TenantName,
		Type:       tenant.Type,
		Application: v1.HTTPTenantApplication{
			Name: tenant.Application.Name,
			Size: tenant.Application.Size,
		},
		NetworkSecurity: v1.NetworkSecurity{
			InterNamespaceTraffic: tenant.NetworkSecurity.InterNamespaceTraffic,
			AllowedNamespaces:     tenant.NetworkSecurity.AllowedNamespaces,
		},
		Sizes: sizes,
	}

	_, dc := getKubeClientset()

	tenantDeploy := makeTenantConfig(tenantNew, dataplaneName)

	_, err = dc.Resource(tenantGVK).Namespace(customerName).Create(context.TODO(), tenantDeploy, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(TenantCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(TenantCreateIntiated, success, nil, 200)
	res.SetResponse(&w)

}

func GetTenantStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]
	_ = vars["dataplane_name"]
	tenantName := vars["tenant_name"]

	_, dc := getKubeClientset()

	tenant, err := dc.Resource(tenantGVK).Namespace(customerName).Get(context.TODO(), tenantName, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(TenantGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	status, _, _ := unstructured.NestedString(tenant.Object, "status", "phase")
	res := NewResponse("", status, nil, 200)
	res.SetResponse(&w)

}
