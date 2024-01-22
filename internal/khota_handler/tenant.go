package khota_handler

import (
	"context"
	"encoding/json"
	"io"
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

	tenantName := vars["tenant_name"]
	customerName := vars["customer_name"]

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

	var tenant v1.HTTPTenant

	if err := json.Unmarshal(body, &tenant); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	tenantNew := v1.HTTPTenant{
		Application: v1.HTTPTenantApplication{
			Name: tenant.Application.Name,
			Size: tenant.Application.Size,
		},
		NetworkSecurity: v1.NetworkSecurity{
			InterNamespaceTraffic: tenant.NetworkSecurity.InterNamespaceTraffic,
			AllowedNamespaces:     tenant.NetworkSecurity.AllowedNamespaces,
		},
	}

	kc, dc := getKubeClientset()

	customer, err := kc.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(CustomerNamespaceGetFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	tenantLabels := map[string]string{
		"dataplane":                customer.GetLabels()["dataplane"],
		"customer_" + customerName: customerName,
		"application":              tenant.Application.Name,
		"size":                     tenant.Application.Size,
	}

	dpList, err := dc.Resource(dpGVK).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		res := NewResponse(DataPlaneListFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	var dpNamespace string
	for _, dp := range dpList.Items {
		if dp.GetName() == customer.GetLabels()["dataplane"] {
			dpType := dp.GetLabels()["dataplane_type"]
			if dpType == string(v1.SharedSaaS) {
				dpNamespace = string(v1.SharedSaaS)
			} else if dpType == string(v1.DedicatedSaaS) {
				dpNamespace = matchStringInMap("customer_", dp.GetLabels())
			} else if dpType == string(v1.PrivateSaaS) {
				dpNamespace = matchStringInMap("customer_", dp.GetLabels())
				tenantLabels = mergeMaps(tenantLabels, map[string]string{
					"private_object": "true",
				})
			}

			phase, _, _ := unstructured.NestedString(dp.Object, "status", "phase")
			if phase != string(v1.ActiveD) {
				res := NewResponse(TenantInfraCreateFailDataplaneNotActive, req_error, nil, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
				return
			}

			if !checkValueInMap(customerName, dp.GetLabels()) {
				res := NewResponse(CustomerNotExistInDataplane, req_error, nil, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
				return
			}

		}
	}

	dataplane, err := dc.Resource(dpGVK).Namespace(dpNamespace).Get(context.TODO(), customer.GetLabels()["dataplane"], metav1.GetOptions{})
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
	tenantDeploy := makeTenantConfig(tenantName, tenantNew, customer.GetLabels()["dataplane"], tenantLabels)

	_, err = dc.Resource(tenantGVK).Namespace(customerName).Create(context.TODO(), tenantDeploy, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(TenantCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(TenantCreateIntiated, success, nil, http.StatusOK)
	res.SetResponse(&w)

}

type tenantListResp struct {
	TenantName    string `json:"tenant"`
	CustomerName  string `json:"customer"`
	DataplaneName string `json:"dataplane"`
	Application   string `json:"application"`
	Size          string `json:"size"`
}

func GetAllTenantInCustomer(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]

	_, dc := getKubeClientset()

	tenantList, err := dc.Resource(tenantGVK).Namespace(customerName).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		res := NewResponse(TenantListFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	var tenantResp []tenantListResp
	for _, tenant := range tenantList.Items {
		newTenantResp := tenantListResp{
			TenantName:    tenant.GetName(),
			CustomerName:  customerName,
			DataplaneName: tenant.GetLabels()["dataplane"],
			Size:          tenant.GetLabels()["size"],
			Application:   tenant.GetLabels()["application"],
		}
		tenantResp = append(tenantResp, newTenantResp)
	}

	bytes, _ := json.Marshal(tenantResp)
	sendJsonResponse(bytes, http.StatusOK, &w)
}
