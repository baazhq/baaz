package http_handlers

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	v1 "datainfra.io/ballastdata/api/v1/types"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var dpGVK = schema.GroupVersionResource{
	Group:    "datainfra.io",
	Version:  "v1",
	Resource: "dataplanes",
}

func CreateDataPlane(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	dataPlaneName := vars["name"]

	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		msg := "Request Size exceeds LimitReader"
		res := NewResponse(msg, req_error, err, http.StatusBadRequest)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if err := req.Body.Close(); err != nil {
		msg := "req body failed to close"
		res := NewResponse(msg, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	var dp v1.DataPlane

	if err := json.Unmarshal(body, &dp); err != nil {
		msg := "Unmarshall error"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	dataplane := v1.DataPlane{
		CloudType:    dp.CloudType,
		SaaSType:     dp.SaaSType,
		CustomerName: dp.CustomerName,
		CloudRegion:  dp.CloudRegion,
		CloudAuth: v1.CloudAuth{
			AwsAccessKey: dp.CloudAuth.AwsAccessKey,
			AwsSecretKey: dp.CloudAuth.AwsSecretKey,
		},
		KubeConfig: v1.KubernetesConfig{
			EKS: v1.EKSConfig{
				Name:             dataPlaneName,
				SubnetIds:        dp.KubeConfig.EKS.SubnetIds,
				SecurityGroupIds: dp.KubeConfig.EKS.SecurityGroupIds,
				Version:          dp.KubeConfig.EKS.Version,
			},
		},
	}

	_, dc := getKubeClientset()

	namespace := getNamespace(dataplane.CustomerName, dataplane.SaaSType)
	dpSecret := getAwsEksSecret(dataPlaneName, dataplane)

	_, err = dc.Resource(secretGVK).Namespace(namespace).Create(context.TODO(), dpSecret, metav1.CreateOptions{})
	if err != nil {
		msg := "POST DataPlane Secret Creation Failed"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	dpDeploy := getAwsEksConfig(dataPlaneName, dataplane)

	_, err = dc.Resource(dpGVK).Namespace(namespace).Create(context.TODO(), dpDeploy, metav1.CreateOptions{})
	if err != nil {
		msg := "POST DataPlane Creation Failed"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(dataplane_creation_initated, success, nil, 200)
	res.SetResponse(&w)

}

func GetDataPlaneStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	dataPlaneName := vars["dataplane_name"]
	customerName := vars["customer_name"]
	saasType := vars["saas_type"]

	namespace := getNamespace(customerName, v1.SaaSTypes(saasType))
	_, dc := getKubeClientset()

	dp, err := dc.Resource(dpGVK).Namespace(namespace).Get(context.TODO(), dataPlaneName, metav1.GetOptions{})
	if err != nil {
		msg := "GET DataPlane Status Fail"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	status, _, _ := unstructured.NestedString(dp.Object, "status", "phase")

	res := NewResponse("", status, nil, 200)
	res.SetResponse(&w)
}
