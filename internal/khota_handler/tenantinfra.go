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

	dpList, err := dc.Resource(dpGVK).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		res := NewResponse(DataPlaneListFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	var namespace string
	for _, dp := range dpList.Items {
		if dp.GetName() == dataplaneName {
			dpType := dp.GetLabels()["dataplane_type"]
			if dpType == string(v1.SharedSaaS) {
				namespace = string(v1.SharedSaaS)
			} else if dpType == string(v1.DedicatedSaaS) {
				namespace = matchStringInMap("customer_", dp.GetLabels())
			}

			phase, _, _ := unstructured.NestedString(dp.Object, "status", "phase")
			if phase != string(v1.ActiveD) {
				res := NewResponse(TenantInfraCreateFailDataplaneNotActive, req_error, nil, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
				return
			}
		 }
	}

	_, err = dc.Resource(tenantInfraGVK).Namespace(namespace).Create(context.TODO(), makeTenantsInfra(dataplaneName, &tenantsInfra), metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(TenantsInfraCreateFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(TenantsInfraCreateSuccess, success, nil, http.StatusOK)
	res.SetResponse(&w)
	res.LogResponse()
	return

}

func GetTenantInfra(w http.ResponseWriter, req *http.Request) {
	_, dc := getKubeClientset()

	dataplane := mux.Vars(req)["dataplane_name"]

	tenantInfra, err := dc.Resource(tenantInfraGVK).Namespace("shared").Get(context.TODO(), dataplane+"-tenantinfra", metav1.GetOptions{})
	if err != nil {
		res := NewResponse(TenantsInfraGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	type tenantInfraResp struct {
		Name              string            `json:"name"`
		DataplaneName     string            `json:"dataplane"`
		MachinePoolStatus map[string]string `json:"machine_pool_status"`
		TenantSizes       interface{}       `json:"tenant_sizes"`
		Status            string            `json:"status"`
	}

	status, _, _ := unstructured.NestedString(tenantInfra.Object, "status", "phase")

	machinePoolStatus, _, _ := unstructured.NestedStringMap(tenantInfra.Object, "status", "machinePoolStatus")

	tenantSizes, _, _ := unstructured.NestedSlice(tenantInfra.Object, "spec", "tenantSizes")

	resp := tenantInfraResp{
		Name:              tenantInfra.GetName(),
		DataplaneName:     tenantInfra.GetLabels()["dataplane_name"],
		MachinePoolStatus: machinePoolStatus,
		TenantSizes:       tenantSizes,
		Status:            status,
	}
	bytes, _ := json.Marshal(resp)
	sendJsonResponse(bytes, http.StatusOK, &w)

}
