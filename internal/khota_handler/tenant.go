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
	"k8s.io/apimachinery/pkg/types"
)

var tenantGVK = schema.GroupVersionResource{
	Group:    "datainfra.io",
	Version:  "v1",
	Resource: "tenants",
}

func CreateTenant(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	dataplaneName := vars["dataplane_name"]
	tenantName := vars["tenant_name"]
	customerName := vars["customer_name"]

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

	tenantNew := v1.HTTPTenant{
		TenantName:   tenant.TenantName,
		CustomerName: tenant.CustomerName,
		Application: v1.HTTPTenantApplication{
			Name: tenant.Application.Name,
			Size: tenant.Application.Size,
		},
		NetworkSecurity: v1.NetworkSecurity{
			InterNamespaceTraffic: tenant.NetworkSecurity.InterNamespaceTraffic,
			AllowedNamespaces:     tenant.NetworkSecurity.AllowedNamespaces,
		},
	}

	_, dc := getKubeClientset()

	tenant_labels := map[string]string{
		"dataplane":                dataplaneName,
		"customer_" + customerName: customerName,
		"application":              tenant.Application.Name,
		"size":                     tenant.Application.Size,
	}

	tenantDeploy := makeTenantConfig(tenantName, tenantNew, dataplaneName, tenant_labels)

	dataplane, err := dc.Resource(dpGVK).Namespace("shared").Get(context.TODO(), dataplaneName, metav1.GetOptions{})
	if err != nil {
		return
	}

	dpLabels := mergeMaps(dataplane.GetLabels(), map[string]string{
		"tenant_" + tenantName: tenantName,
	})

	patchBytes := NewPatchValue("replace", "/metadata/labels", dpLabels)

	_, patchErr := dc.Resource(dpGVK).Namespace(dataplane.GetNamespace()).Patch(
		context.TODO(),
		dataplane.GetName(),
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		res := NewResponse(DataplanePatchFail, internal_error, patchErr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	_, err = dc.Resource(tenantGVK).Namespace(dataplane.GetNamespace()).Create(context.TODO(), tenantDeploy, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(TenantCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(TenantCreateIntiated, success, nil, http.StatusOK)
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
	res := NewResponse("", status, nil, http.StatusOK)
	res.SetResponse(&w)

}
