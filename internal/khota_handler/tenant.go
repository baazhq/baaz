package khota_handler

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	v1 "datainfra.io/ballastdata/api/v1/types"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	var tenant v1.Tenant

	if err := json.Unmarshal(body, &tenant); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	tenantNew := v1.Tenant{
		TenantName:    tenant.TenantName,
		Type:          tenant.Type,
		DataplaneName: tenant.DataplaneName,
		Application: v1.HTTPTenantApplication{
			Name: tenant.Application.Name,
			Size: tenant.Application.Size,
		},
		Sizes: v1.HTTPTenantSizes{
			Name:  tenant.Sizes.Name,
			Nodes: tenant.Sizes.Nodes,
		},
	}

	_, dc := getKubeClientset()

	tenantDeploy := makeTenantConfig(tenantNew)

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
