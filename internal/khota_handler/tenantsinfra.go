package khota_handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/baazhq/baaz/api/v1/types"
)

var tenantInfraGVK = schema.GroupVersionResource{
	Group:    "baaz.dev",
	Version:  "v1",
	Resource: "tenantsinfras",
}

func CreateTenantInfra(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	dataplaneName := vars["dataplane_name"]

	body, err := io.ReadAll(io.LimitReader(req.Body, 1048576))
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

	var tenantsInfra map[string]v1.HTTPTenantSizes

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
	var labels map[string]string
	labels = map[string]string{
		"dataplane_name": dataplaneName,
	}

	for _, dp := range dpList.Items {
		if dp.GetName() == dataplaneName {
			dpType := dp.GetLabels()["dataplane_type"]
			if dpType == string(v1.SharedSaaS) {
				namespace = string(v1.SharedSaaS)
			} else if dpType == string(v1.DedicatedSaaS) {
				namespace = matchStringInMap("customer_", dp.GetLabels())
			} else if dpType == string(v1.PrivateSaaS) {
				namespace = matchStringInMap("customer_", dp.GetLabels())
				labels = mergeMaps(labels, map[string]string{
					v1.PrivateObjectLabelKey: "true",
				})
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

	infra := makeTenantsInfra(dataplaneName, tenantsInfra, labels)

	_, err = dc.Resource(tenantInfraGVK).Namespace(namespace).Create(context.TODO(), infra, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(TenantsInfraCreateFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		sendEventParseable(tenantsInfraEventStream, tenantsInfraInitiationFailEvent, labels, map[string]string{"tenant_name": infra.GetName()})
		return
	}

	res := NewResponse(TenantsInfraCreateInitiated, success, nil, http.StatusOK)
	res.SetResponse(&w)
	res.LogResponse()
	sendEventParseable(tenantsInfraEventStream, tenantsInfraInitiationSuccessEvent, labels, map[string]string{"tenant_name": infra.GetName()})

}

func ListTenantInfra(w http.ResponseWriter, req *http.Request) {
	_, dc := getKubeClientset()

	dataplane := mux.Vars(req)["dataplane_name"]

	tenantsInfras, err := dc.Resource(tenantInfraGVK).Namespace("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "dataplane_name=" + dataplane,
	})
	if err != nil {
		res := NewResponse(TenantsInfraGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	type tenantInfraResp struct {
		Name              string                 `json:"name"`
		DataplaneName     string                 `json:"dataplane"`
		MachinePoolStatus map[string]interface{} `json:"machine_pool_status"`
		TenantSizes       map[string]interface{} `json:"tenant_sizes"`
		Status            string                 `json:"status"`
	}

	var tenantsInfrasResp []tenantInfraResp

	for _, ti := range tenantsInfras.Items {
		status, _, _ := unstructured.NestedString(ti.Object, "status", "phase")

		machinePoolStatus, _, _ := unstructured.NestedMap(ti.Object, "status", "machinePoolStatus")

		tenantSizes, _, _ := unstructured.NestedMap(ti.Object, "spec", "tenantSizes")

		resp := tenantInfraResp{
			Name:              ti.GetName(),
			DataplaneName:     ti.GetLabels()["dataplane_name"],
			MachinePoolStatus: machinePoolStatus,
			TenantSizes:       tenantSizes,
			Status:            status,
		}

		tenantsInfrasResp = append(tenantsInfrasResp, resp)
	}

	bytes, _ := json.Marshal(tenantsInfrasResp)
	sendJsonResponse(bytes, http.StatusOK, &w)

}

func DeleteTenantInfra(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	dataplaneName := vars["dataplane_name"]
	tenantsInfraName := vars["tenantsinfra_name"]

	_, dc := getKubeClientset()

	tenantsInfraObj, err := dc.Resource(tenantInfraGVK).Namespace("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "dataplane_name=" + dataplaneName,
	})
	if err != nil {
		res := NewResponse(TenantsInfraListFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	for _, tf := range tenantsInfraObj.Items {
		if tf.GetName() == tenantsInfraName {
			err := dc.Resource(tenantInfraGVK).Namespace(tf.GetNamespace()).Delete(context.TODO(), tf.GetName(), metav1.DeleteOptions{})
			if err != nil {
				res := NewResponse(TenantsInfraDeleteFail, internal_error, err, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
				return
			}
			res := NewResponse(TenantsInfraDeleteInitiated, success, nil, http.StatusOK)
			res.SetResponse(&w)
			res.LogResponse()
			sendEventParseable(tenantsInfraEventStream, tenantsInfraInitiationSuccessEvent, tf.GetLabels(), map[string]string{"tenant_name": tf.GetName()})
		}
	}

}

func UpdateTenantInfra(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	dataplaneName := vars["dataplane_name"]
	tenantsInfraName := vars["tenantsinfra_name"]

	body, err := io.ReadAll(io.LimitReader(req.Body, 1048576))
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

	var tenantsInfra map[string]v1.HTTPTenantSizes

	if err := json.Unmarshal(body, &tenantsInfra); err != nil {
		res := NewResponse(JsonMarshallError, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	_, dc := getKubeClientset()

	dpList, err := dc.Resource(dpGVK).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		res := NewResponse(TenantInfraUpdateFail, req_error, err, http.StatusInternalServerError)
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
			} else if dpType == string(v1.PrivateSaaS) {
				namespace = matchStringInMap("customer_", dp.GetLabels())
			}

			phase, _, _ := unstructured.NestedString(dp.Object, "status", "phase")
			if phase != string(v1.ActiveD) {
				res := NewResponse(TenantInfraUpdateFailDataplaneNotActive, req_error, nil, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
				return
			}
		}
	}

	existingObj, err := dc.Resource(tenantInfraGVK).Namespace(namespace).Get(context.TODO(), tenantsInfraName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			res := NewResponse(TenantInfraUpdateFail, internal_error, err, http.StatusNotFound)
			res.SetResponse(&w)
			res.LogResponse()
		} else {
			res := NewResponse(TenantInfraUpdateFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
		}
		return
	}
	ob := &v1.TenantsInfra{}
	if errCon := runtime.DefaultUnstructuredConverter.FromUnstructured(existingObj.Object, ob); errCon != nil {
		res := NewResponse(TenantInfraUpdateFail, req_error, errCon, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	allTenantSizes := make(map[string]v1.TenantSizes)
	for tName, tenantSize := range tenantsInfra {
		allTenantSizes[tName] = v1.TenantSizes(tenantSize)
	}
	ob.Spec.TenantSizes = allTenantSizes

	upObj, errCon := runtime.DefaultUnstructuredConverter.ToUnstructured(ob)
	if errCon != nil {
		res := NewResponse(TenantInfraUpdateFail, req_error, errCon, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	_, uperr := dc.Resource(tenantInfraGVK).Namespace(namespace).Update(context.TODO(), &unstructured.Unstructured{Object: upObj}, metav1.UpdateOptions{})
	if uperr != nil {
		res := NewResponse(TenantInfraUpdateFail, req_error, uperr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}
	res := NewResponse(TenantInfraUpdateSuccess, success, nil, http.StatusOK)
	res.SetResponse(&w)
	res.LogResponse()
}

func GetTenantInfra(w http.ResponseWriter, req *http.Request) {
	_, dc := getKubeClientset()

	dataplane := mux.Vars(req)["dataplane_name"]
	tenantInfra := mux.Vars(req)["tenantinfra_name"]

	tenantsInfras, err := dc.Resource(tenantInfraGVK).Namespace("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "dataplane_name=" + dataplane,
	})
	if err != nil {
		res := NewResponse(TenantsInfraGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	type tenantInfraResp struct {
		Name              string                 `json:"name"`
		DataplaneName     string                 `json:"dataplane"`
		MachinePoolStatus map[string]interface{} `json:"machine_pool_status"`
		TenantSizes       map[string]interface{} `json:"tenant_sizes"`
		Status            string                 `json:"status"`
	}

	var tenantsInfrasResp []tenantInfraResp

	for _, ti := range tenantsInfras.Items {
		if ti.GetName() == tenantInfra {
			status, _, _ := unstructured.NestedString(ti.Object, "status", "phase")

			machinePoolStatus, _, _ := unstructured.NestedMap(ti.Object, "status", "machinePoolStatus")

			tenantSizes, _, _ := unstructured.NestedMap(ti.Object, "spec", "tenantSizes")

			resp := tenantInfraResp{
				Name:              ti.GetName(),
				DataplaneName:     ti.GetLabels()["dataplane_name"],
				MachinePoolStatus: machinePoolStatus,
				TenantSizes:       tenantSizes,
				Status:            status,
			}

			tenantsInfrasResp = append(tenantsInfrasResp, resp)
			break
		}
	}

	bytes, _ := json.Marshal(tenantsInfrasResp)
	sendJsonResponse(bytes, http.StatusOK, &w)
}
