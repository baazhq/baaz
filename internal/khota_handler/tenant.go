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
	"k8s.io/apimachinery/pkg/types"

	v1 "github.com/baazhq/baaz/api/v1/types"
)

var tenantGVK = schema.GroupVersionResource{
	Group:    "baaz.dev",
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
	if patchErr != nil {
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
		sendEventParseable(tenantsEventStream, tenantsCreationFailEvent, tenantLabels, map[string]string{"tenant_name": tenantName})
		return
	}

	res := NewResponse(TenantCreateIntiated, success, nil, http.StatusOK)
	res.SetResponse(&w)
	sendEventParseable(tenantsEventStream, tenantsCreationSuccessEvent, tenantLabels, map[string]string{"tenant_name": tenantName})

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

func DeleteTenant(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	tenantName := vars["tenant_name"]
	customerName := vars["customer_name"]

	kc, dc := getKubeClientset()

	customer, err := kc.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(CustomerNamespaceGetFail, req_error, err, http.StatusNotFound)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	err = dc.Resource(tenantGVK).Namespace(customer.Name).Delete(req.Context(), tenantName, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			res := NewResponse(TenantDeleteFail, req_error, err, http.StatusNotFound)
			res.SetResponse(&w)
			res.LogResponse()
		} else {
			res := NewResponse(TenantDeleteFail, req_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
		}
		return
	}
	res := NewResponse(TenantDeleteIntiated, success, nil, http.StatusOK)
	res.SetResponse(&w)
}

func UpdateTenant(w http.ResponseWriter, req *http.Request) {
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

	_, dc := getKubeClientset()

	ob, err := dc.Resource(tenantGVK).Namespace(customerName).Get(req.Context(), tenantName, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(TenantUpdateFail, resource_not_found, err, http.StatusNotFound)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	curTenant := &v1.Tenants{}
	if errCon := runtime.DefaultUnstructuredConverter.FromUnstructured(ob.Object, curTenant); errCon != nil {
		res := NewResponse(TenantUpdateFail, req_error, errCon, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return

	}

	updatedTenant := curTenant.DeepCopy()
	for i, tn := range updatedTenant.Spec.TenantConfig {
		if string(tn.AppType) == tenant.Application.Name {
			updatedTenant.Spec.TenantConfig[i].Size = tenant.Application.Size

		}
	}

	tenantUns, err := runtime.DefaultUnstructuredConverter.ToUnstructured(updatedTenant)
	if err != nil {
		res := NewResponse(TenantUpdateFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	_, err = dc.Resource(tenantGVK).Namespace(customerName).Update(context.TODO(), &unstructured.Unstructured{Object: tenantUns}, metav1.UpdateOptions{})
	if err != nil {
		res := NewResponse(TenantCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(TenantCreateIntiated, success, nil, http.StatusOK)
	res.SetResponse(&w)
}
