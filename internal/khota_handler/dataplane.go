package khota_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	v1 "datainfra.io/baaz/api/v1/types"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
)

var dpGVK = schema.GroupVersionResource{
	Group:    "datainfra.io",
	Version:  "v1",
	Resource: "dataplanes",
}

var secretGVK = schema.GroupVersionResource{
	Version:  "v1",
	Resource: "secrets",
}

func AddDataPlane(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]
	dataplaneName := vars["dataplane_name"]

	kc, dc := getKubeClientset()

	customer, getErr := kc.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if getErr != nil {
		res := NewResponse(CustomerNamespaceGetFail, internal_error, getErr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	customer.ObjectMeta.Labels = mergeMaps(customer.Labels, map[string]string{
		"dataplane": dataplaneName,
	})

	var dataplaneNs string
	if customer.Labels["saas_type"] == string(v1.SharedSaaS) {
		dataplaneNs = shared_namespace
	} else {
		dataplaneNs = dedicated_namespace
	}

	fmt.Println(dataplaneNs, dataplaneName)
	dataplane, err := dc.Resource(dpGVK).Namespace("shared").Get(context.TODO(), dataplaneName, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	dataplane.SetLabels(mergeMaps(dataplane.GetLabels(), map[string]string{
		"customer_" + customerName: customerName,
	}))

	_, updateErr := kc.CoreV1().Namespaces().Update(context.TODO(), customer, metav1.UpdateOptions{})
	if updateErr != nil {
		res := NewResponse(CustomerNamespaceUpdateFail, internal_error, getErr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(DataplaneAddedSuccess, success, nil, http.StatusOK)
	res.SetResponse(&w)
}

func CreateDataPlane(w http.ResponseWriter, req *http.Request) {
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

	kc, dc := getKubeClientset()

	namespace := getNamespace(customerName, dataplane.SaaSType)
	dpSecret := getAwsEksSecret(dataplaneName, dataplane)

	_, err = dc.Resource(secretGVK).Namespace(namespace).Create(context.TODO(), dpSecret, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(DataPlaneCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, getErr := kc.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}

		result.ObjectMeta.Labels = mergeMaps(result.Labels, map[string]string{
			"dataplane": dataplaneName,
		})
		_, updateErr := kc.CoreV1().Namespaces().Update(context.TODO(), result, metav1.UpdateOptions{})
		return updateErr
	})
	if retryErr != nil {
		res := NewResponse(DataPlaneCreateFail, internal_error, retryErr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	labels := map[string]string{
		"version":                  dataplane.KubeConfig.EKS.Version,
		"cloud_type":               string(dataplane.CloudType),
		"cloud_region":             dataplane.CloudRegion,
		"saas_type":                string(dataplane.SaaSType),
		"customer_" + customerName: customerName,
	}

	dpDeploy := makeAwsEksConfig(dataplaneName, dataplane, labels)

	_, err = dc.Resource(dpGVK).Namespace(namespace).Create(context.TODO(), dpDeploy, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(DataPlaneCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(DataPlaneCreateIntiated, success, nil, http.StatusOK)
	res.LogResponse()
	res.SetResponse(&w)

}

func GetDataPlaneStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]

	kc, dc := getKubeClientset()

	namespace, getErr := kc.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if getErr != nil {
		res := NewResponse(DataPlaneGetFail, internal_error, getErr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	type dataplaneResp struct {
		Name        string `json:"name"`
		CloudRegion string `json:"cloud_region"`
		CloudType   string `json:"cloud_type"`
		SaaSType    string `json:"saas_type"`
		Version     string `json:"version"`
		Status      string `json:"status"`
	}

	if namespace.Labels["saas_type"] == string(v1.SharedSaaS) {
		dpName := namespace.Labels["dataplane"]
		dpObj, err := dc.Resource(dpGVK).Namespace("shared").Get(context.TODO(), dpName, metav1.GetOptions{})
		if err != nil {
			res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}

		status, _, _ := unstructured.NestedString(dpObj.Object, "status", "phase")

		newdataplaneResp := dataplaneResp{
			Name:        dpName,
			CloudRegion: dpObj.GetLabels()["cloud_region"],
			SaaSType:    dpObj.GetLabels()["saas_type"],
			CloudType:   dpObj.GetLabels()["cloud_type"],
			Version:     dpObj.GetLabels()["version"],
			Status:      status,
		}

		dpResp, err := json.Marshal(newdataplaneResp)
		if err != nil {
			res := NewResponse(DataPlaneGetFail, string(JsonMarshallError), err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}
		res := NewResponse(CustomMsg(dpResp), status, nil, http.StatusOK)
		res.SetMsgResponse(&w)
		res.LogResponse()
	}

}

func DeleteDataPlane(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]

	kc, dc := getKubeClientset()

	namespace, getErr := kc.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if getErr != nil {
		res := NewResponse(DataPlaneGetFail, internal_error, getErr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if namespace.Labels["saas_type"] == string(v1.SharedSaaS) {
		dp := namespace.Labels["dataplane"]
		err := dc.Resource(dpGVK).Namespace("shared").Delete(context.TODO(), dp, metav1.DeleteOptions{})
		if err != nil {
			res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}

		res := NewResponse("", string(DataplaneDeletionInitiated), nil, http.StatusOK)
		res.SetResponse(&w)
	} else if namespace.Labels["saas_type"] == string(v1.DedicatedSaaS) {
		dp := namespace.Labels["dataplane"]
		err := dc.Resource(dpGVK).Namespace(customerName).Delete(context.TODO(), dp, metav1.DeleteOptions{})
		if err != nil {
			res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}

		res := NewResponse("", string(DataplaneDeletionInitiated), nil, http.StatusOK)
		res.SetResponse(&w)
	}

}

func ListDataplane(w http.ResponseWriter, req *http.Request) {

}
