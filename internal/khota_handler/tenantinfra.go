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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var tenantInfraGVK = schema.GroupVersionResource{
	Group:    "datainfra.io",
	Version:  "v1",
	Resource: "tenantsinfras",
}

func CreateTenantInfra(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

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

	var tenantsInfra []v1.HTTPTenantSizes

	if err := json.Unmarshal(body, &tenantsInfra); err != nil {
		res := NewResponse(JsonMarshallError, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	_, dc := getKubeClientset()

	_, err = dc.Resource(tenantInfraGVK).Namespace("shared").Create(context.TODO(), makeTenantsInfra(dataplaneName, &tenantsInfra), metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(TenantSizeCreateFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(TenantSizeCreateSuccess, success, nil, http.StatusOK)
	res.SetResponse(&w)
	res.LogResponse()
	return

}
