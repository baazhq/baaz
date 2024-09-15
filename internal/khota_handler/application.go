package khota_handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/baazhq/baaz/api/v1/types"
)

var applicationGVK = schema.GroupVersionResource{
	Group:    "baaz.dev",
	Version:  "v1",
	Resource: "applications",
}

func CreateApplication(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]
	tenantName := vars["tenant_name"]
	applicationName := customerName + "-" + tenantName + "-" + "apps"

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
	var applications []v1.HTTPApplication

	if err := json.Unmarshal(body, &applications); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	kc, dc := getKubeClientset()

	customer, err := kc.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(CustomerNamespaceGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	dataplaneName := customer.GetLabels()["dataplane"]
	var labels map[string]string
	if customer.GetLabels()["saas_type"] == string(v1.PrivateSaaS) {
		labels = map[string]string{
			"customer_" + customerName: customerName,
			v1.PrivateObjectLabelKey:   "true",
			"tenant":                   tenantName,
		}
	}
	appDeploy := makeApplicationConfig(applications, dataplaneName, tenantName, applicationName, labels)

	_, err = dc.Resource(applicationGVK).Namespace(customerName).Create(context.TODO(), appDeploy, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(ApplicationCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		sendEventParseable(applicationsEventStream, ApplicationCreationFailEvent, appDeploy.GetLabels(), map[string]string{"application_name": appDeploy.GetName()})
		return
	}

	res := NewResponse(ApplicationCreateIntiated, success, nil, http.StatusOK)
	sendEventParseable(applicationsEventStream, ApplicationCreationSuccessEvent, appDeploy.GetLabels(), map[string]string{"application_name": appDeploy.GetName()})
	res.SetResponse(&w)

}

func ListApplicationStatus(w http.ResponseWriter, req *http.Request) {
	//vars := mux.Vars(req)

	// customerName := vars["customer_name"]
	// applicationName := vars["application_name"]

	_, dc := getKubeClientset()

	application, err := dc.Resource(applicationGVK).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		res := NewResponse(ApplicationGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	type listApplicationResp struct {
		Name          string                 `json:"name"`
		DataplaneName string                 `json:"dataplane"`
		TenantName    string                 `json:"tenant"`
		AppStatus     map[string]interface{} `json:"status"`
	}
	var appResp []listApplicationResp

	for _, app := range application.Items {

		dpName, _, _ := unstructured.NestedString(app.Object, "status", "applicationCurrentSpec", "dataplane")
		tenantName, _, _ := unstructured.NestedString(app.Object, "status", "applicationCurrentSpec", "tenant")

		appStatus, _, _ := unstructured.NestedMap(app.Object, "status", "appStatus")

		resp := listApplicationResp{
			Name:          app.GetName(),
			DataplaneName: dpName,
			TenantName:    tenantName,
			AppStatus:     appStatus,
		}

		appResp = append(appResp, resp)
	}
	bytes, _ := json.Marshal(appResp)
	sendJsonResponse(bytes, http.StatusOK, &w)

}

func DeleteApplicationStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]
	applicationName := vars["application_name"]

	_, dc := getKubeClientset()

	err := dc.Resource(applicationGVK).Namespace(customerName).Delete(context.TODO(), applicationName, metav1.DeleteOptions{})
	if err != nil {
		res := NewResponse(ApplicationGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse("", string(ApplicationDeleteIntiated), nil, http.StatusOK)
	res.SetResponse(&w)

}

func UpdateApplication(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]
	applicationName := vars["application_name"]

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
	var applications []v1.HTTPApplication

	if err := json.Unmarshal(body, &applications); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	_, dc := getKubeClientset()

	existingObj, err := dc.Resource(applicationGVK).Namespace(customerName).Get(context.TODO(), applicationName, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(ApplicationUpdateFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}
	ob := &v1.Applications{}
	if errCon := runtime.DefaultUnstructuredConverter.FromUnstructured(existingObj.Object, ob); errCon != nil {
		res := NewResponse(ApplicationUpdateFail, req_error, errCon, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	inputAppMap := make(map[string]v1.HTTPApplication)
	for _, app := range applications {
		inputAppMap[app.ApplicationName] = app
	}

	for idx, app := range ob.Spec.Applications {
		if inputApp, found := inputAppMap[app.Name]; found {
			ob.Spec.Applications[idx].Spec.Version = inputApp.Version
		}
	}

	upObj, errCon := runtime.DefaultUnstructuredConverter.ToUnstructured(ob)
	if errCon != nil {
		res := NewResponse(ApplicationUpdateFail, req_error, errCon, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	_, uperr := dc.Resource(applicationGVK).Namespace(customerName).Update(context.TODO(), &unstructured.Unstructured{Object: upObj}, metav1.UpdateOptions{})
	if uperr != nil {
		res := NewResponse(ApplicationUpdateFail, req_error, uperr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}
	res := NewResponse(ApplicationUpdateSuccess, success, nil, http.StatusOK)
	res.SetResponse(&w)
	res.LogResponse()
}
