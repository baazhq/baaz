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

var applicationGVK = schema.GroupVersionResource{
	Group:    "datainfra.io",
	Version:  "v1",
	Resource: "applications",
}

func CreateApplications(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]
	dataplaneName := vars["dataplane_name"]
	applicationName := vars["application_name"]

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

	var application v1.HTTPApplication

	if err := json.Unmarshal(body, &application); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	appNew := v1.HTTPApplication{
		Scope:      application.Scope,
		TenantName: application.TenantName,
		ChartName:  application.ChartName,
		RepoName:   application.RepoName,
		RepoURL:    application.RepoURL,
		Version:    application.Version,
	}

	_, dc := getKubeClientset()

	appDeploy := makeApplicationConfig(appNew, dataplaneName, applicationName)

	_, err = dc.Resource(applicationGVK).Namespace(customerName).Create(context.TODO(), appDeploy, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(ApplicationCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(ApplicationCreateIntiated, success, nil, 200)
	res.SetResponse(&w)

}
