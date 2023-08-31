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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var dpGVK = schema.GroupVersionResource{
	Group:    "datainfra.io",
	Version:  "v1",
	Resource: "dataplanes",
}

func CreateDataPlane(w http.ResponseWriter, req *http.Request) {

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

	var dp v1.DataPlane

	if err := json.Unmarshal(body, &dp); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	dataplaneName := makeDataPlaneName(dp.CloudType, dp.CloudRegion, dp.SaaSType)
	dataplane := v1.DataPlane{
		CloudType:   dp.CloudType,
		SaaSType:    dp.SaaSType,
		CloudRegion: dp.CloudRegion,
		CloudAuth: v1.CloudAuth{
			AwsAuth: v1.AwsAuth{
				AwsAccessKey: dp.CloudAuth.AwsAuth.AwsAccessKey,
				AwsSecretKey: dp.CloudAuth.AwsAuth.AwsSecretKey,
			},
		},
		KubeConfig: v1.KubernetesConfig{
			EKS: v1.EKSConfig{
				Name:             dataplaneName,
				SubnetIds:        dp.KubeConfig.EKS.SubnetIds,
				SecurityGroupIds: dp.KubeConfig.EKS.SecurityGroupIds,
				Version:          dp.KubeConfig.EKS.Version,
			},
		},
	}

	_, dc := getKubeClientset()

	namespace := getNamespace(dataplane.CustomerName, dataplane.SaaSType)
	dpSecret := getAwsEksSecret(dataplaneName, dataplane)

	_, err = dc.Resource(secretGVK).Namespace(namespace).Create(context.TODO(), dpSecret, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(DataPlaneCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	dpDeploy := makeAwsEksConfig(dataplaneName, dataplane)

	_, err = dc.Resource(dpGVK).Namespace(namespace).Create(context.TODO(), dpDeploy, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(DataPlaneCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(DataPlaneCreateIntiated, success, nil, 200)
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
		res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	status, _, _ := unstructured.NestedString(dp.Object, "status", "phase")

	res := NewResponse("", status, nil, 200)
	res.SetResponse(&w)
}
