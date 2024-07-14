package khota_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	v1 "github.com/baazhq/baaz/api/v1/types"
)

var dpGVK = schema.GroupVersionResource{
	Group:    "baaz.dev",
	Version:  "v1",
	Resource: "dataplanes",
}

var secretGVK = schema.GroupVersionResource{
	Version:  "v1",
	Resource: "secrets",
}

func AddRemoveDataPlane(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	customerName := vars["customer_name"]
	dataplaneName := vars["dataplane_name"]

	kc, dc := getKubeClientset()

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

	type action struct {
		Action string `json:"action"`
	}

	var a action

	err = json.Unmarshal(body, &a)
	if err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	customer, getErr := kc.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if getErr != nil {
		res := NewResponse(CustomerNamespaceGetFail, internal_error, getErr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	var customerLabels map[string]string
	if a.Action == "add" {
		customerLabels = mergeMaps(customer.Labels, map[string]string{
			"dataplane": dataplaneName,
		})
	} else if a.Action == "remove" {
		customerLabels = mergeMaps(customer.Labels, map[string]string{
			"dataplane": "unavailable",
		})
	}
	customer.ObjectMeta.Labels = customerLabels

	var dataplaneNs string
	if customer.Labels["saas_type"] == string(v1.SharedSaaS) {
		dataplaneNs = shared_namespace
	} else {
		dataplaneNs = customerName
	}

	dataplane, err := dc.Resource(dpGVK).Namespace(dataplaneNs).Get(context.TODO(), dataplaneName, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if a.Action == "add" {
		dpLabels := mergeMaps(dataplane.GetLabels(), map[string]string{
			"customer_" + customerName: customerName,
		})
		patchBytes := NewPatchValue("replace", "/metadata/labels", dpLabels)

		_, patchErr := dc.Resource(dpGVK).Namespace(dataplaneNs).Patch(
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

		_, updateErr := kc.CoreV1().Namespaces().Update(context.TODO(), customer, metav1.UpdateOptions{})
		if updateErr != nil {
			res := NewResponse(CustomerNamespaceUpdateFail, internal_error, getErr, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}

		res := NewResponse(DataplaneAddedSuccess, success, nil, http.StatusOK)
		res.SetResponse(&w)
		res.LogResponse()
		return
	} else if a.Action == "remove" {

		labels := dataplane.GetLabels()
		delete(labels, "customer_"+customerName)

		patchBytes := NewPatchValue("replace", "/metadata/labels", labels)

		_, patchErr := dc.Resource(dpGVK).Namespace(dataplaneNs).Patch(
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

		_, updateErr := kc.CoreV1().Namespaces().Update(context.TODO(), customer, metav1.UpdateOptions{})
		if updateErr != nil {
			res := NewResponse(CustomerNamespaceUpdateFail, internal_error, getErr, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}

		res := NewResponse(DataplaneRemoveSuccess, success, nil, http.StatusOK)
		res.SetResponse(&w)
		res.LogResponse()
		return

	}
}

func CreateDataPlane(w http.ResponseWriter, req *http.Request) {

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

	var dp v1.DataPlane

	if err := json.Unmarshal(body, &dp); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	dpName := makeDataPlaneName(dp.CloudType, dp.CustomerName, dp.CloudRegion)
	dpNamespace := getNamespace(dp.CustomerName)

	var appConfig []v1.HTTPApplication

	for _, app := range dp.ApplicationConfig {
		appConfig = append(appConfig, v1.HTTPApplication{
			ApplicationName: app.ApplicationName,
			Namespace:       app.Namespace,
			ChartName:       app.ChartName,
			RepoName:        app.RepoName,
			RepoURL:         app.RepoURL,
			Version:         app.Version,
			Values:          app.Values,
		})
	}

	dataplane := v1.DataPlane{
		CustomerName: dp.CustomerName,
		CloudType:    dp.CloudType,
		CloudRegion:  dp.CloudRegion,
		CloudAuth: v1.CloudAuth{
			AwsAuth: v1.AwsAuth{
				AwsAccessKey: dp.CloudAuth.AwsAuth.AwsAccessKey,
				AwsSecretKey: dp.CloudAuth.AwsAuth.AwsSecretKey,
			},
		},
		ProvisionNetwork: dp.ProvisionNetwork,
		VpcCidr:          dp.VpcCidr,
		KubeConfig: v1.KubernetesConfig{
			EKS: v1.EKSConfig{
				Name:             dpName,
				SubnetIds:        dp.KubeConfig.EKS.SubnetIds,
				SecurityGroupIds: dp.KubeConfig.EKS.SecurityGroupIds,
				Version:          dp.KubeConfig.EKS.Version,
			},
		},
		ApplicationConfig: appConfig,
	}

	kc, dc := getKubeClientset()

	dpNS, err := kc.CoreV1().Namespaces().Get(context.TODO(), dpNamespace, metav1.GetOptions{})
	if err != nil {
		res := NewResponse(DataPlaneCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	// create secret based on saas type
	if dpNS.GetLabels()[v1.PrivateModeNSLabelKey] != "true" {
		dpSecret := getAwsEksSecret(dpName, dataplane)
		_, err = dc.Resource(secretGVK).Namespace(dpNamespace).Create(context.TODO(), dpSecret, metav1.CreateOptions{})
		if err != nil {
			res := NewResponse(DataPlaneCreateFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}
	}

	labels := map[string]string{
		"version":      dataplane.KubeConfig.EKS.Version,
		"cloud_type":   string(dataplane.CloudType),
		"cloud_region": dataplane.CloudRegion,
	}

	if dp.CustomerName != "" {
		var customer *corev1.Namespace
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			customer, err = kc.CoreV1().Namespaces().Get(context.TODO(), dp.CustomerName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if customer.GetLabels()["dataplane"] != "unavailable" {
				return fmt.Errorf("dataplane exists for customer")
			}

			customer.ObjectMeta.Labels = mergeMaps(customer.Labels, map[string]string{
				"dataplane": dpName,
			})
			_, updateErr := kc.CoreV1().Namespaces().Update(context.TODO(), customer, metav1.UpdateOptions{})
			return updateErr
		})

		if retryErr != nil {
			res := NewResponse(DataPlaneCreateFail, internal_error, retryErr, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}
		labels = mergeMaps(labels, map[string]string{
			"customer_" + dataplane.CustomerName: dataplane.CustomerName,
			"dataplane_type":                     customer.GetLabels()["saas_type"],
		})

		if customer.GetLabels()["saas_type"] == "private" {
			labels = mergeMaps(labels, map[string]string{
				v1.PrivateObjectLabelKey: "true",
			})
		}
	} else {
		labels = mergeMaps(labels, map[string]string{
			"dataplane_type": string(v1.SharedSaaS),
		})
	}

	dpDeploy := makeAwsEksConfig(dpName, dataplane, labels)

	_, err = dc.Resource(dpGVK).Namespace(dpNamespace).Create(context.TODO(), dpDeploy, metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(DataPlaneCreateFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		sendEventParseable(dataplanesEventStream, dataplaneInitiationFailEvent, nil, map[string]string{"dataplane_name": dpName})
		return
	}

	sendEventParseable(dataplanesEventStream, dataplaneInitiationSuccessEvent, labels, map[string]string{"dataplane_name": dpName})
	res := NewResponse(DataPlaneCreateIntiated, success, nil, http.StatusOK)
	res.LogResponse()
	res.SetResponse(&w)

}

func UpdateDataPlane(w http.ResponseWriter, req *http.Request) {

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

	var dp v1.DataPlane

	if err := json.Unmarshal(body, &dp); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	vars := mux.Vars(req)
	dpName := vars["dataplane_name"]

	dpNamespace := getNamespace(dp.CustomerName)

	var appConfig []v1.HTTPApplication

	for _, app := range dp.ApplicationConfig {
		appConfig = append(appConfig, v1.HTTPApplication{
			ApplicationName: app.ApplicationName,
			Namespace:       app.Namespace,
			ChartName:       app.ChartName,
			RepoName:        app.RepoName,
			RepoURL:         app.RepoURL,
			Version:         app.Version,
			Values:          app.Values,
		})
	}

	dataplane := v1.DataPlane{
		CustomerName: dp.CustomerName,
		CloudType:    dp.CloudType,
		CloudRegion:  dp.CloudRegion,
		CloudAuth: v1.CloudAuth{
			AwsAuth: v1.AwsAuth{
				AwsAccessKey: dp.CloudAuth.AwsAuth.AwsAccessKey,
				AwsSecretKey: dp.CloudAuth.AwsAuth.AwsSecretKey,
			},
		},
		ProvisionNetwork: dp.ProvisionNetwork,
		KubeConfig: v1.KubernetesConfig{
			EKS: v1.EKSConfig{
				Name:             dpName,
				SubnetIds:        dp.KubeConfig.EKS.SubnetIds,
				SecurityGroupIds: dp.KubeConfig.EKS.SecurityGroupIds,
				Version:          dp.KubeConfig.EKS.Version,
			},
		},
		ApplicationConfig: appConfig,
	}

	_, dc := getKubeClientset()

	labels := map[string]string{
		"version":      dataplane.KubeConfig.EKS.Version,
		"cloud_type":   string(dataplane.CloudType),
		"cloud_region": dataplane.CloudRegion,
	}

	dpDeploy := makeAwsEksConfig(dpName, dataplane, labels)

	existingObj, err := dc.Resource(dpGVK).Namespace(dpNamespace).Get(context.TODO(), dpDeploy.GetName(), metav1.GetOptions{})
	if err != nil {
		res := NewResponse(DataplaneUpdateFail, req_error, err, http.StatusBadRequest)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	ob := &v1.DataPlanes{}
	if errCon := runtime.DefaultUnstructuredConverter.FromUnstructured(existingObj.Object, ob); errCon != nil {
		res := NewResponse(DataplaneUpdateFail, req_error, errCon, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return

	}

	// for now only k8s is updatable by user
	ob.Spec.CloudInfra.Eks.Version = dataplane.KubeConfig.EKS.Version

	upObj, errCon := runtime.DefaultUnstructuredConverter.ToUnstructured(ob)
	if errCon != nil {
		res := NewResponse(DataplaneUpdateFail, req_error, errCon, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	_, uperr := dc.Resource(dpGVK).Namespace(dpNamespace).Update(context.TODO(), &unstructured.Unstructured{Object: upObj}, metav1.UpdateOptions{})
	if uperr != nil {
		res := NewResponse(DataplaneUpdateFail, req_error, uperr, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}
	res := NewResponse(DataplaneUpdateFail, success, nil, http.StatusOK)
	res.SetResponse(&w)
	res.LogResponse()
	sendEventParseable(dataplanesEventStream, dataplaneInitiationSuccessEvent, labels, map[string]string{"dataplane_name": dpName})
}

func GetDataPlaneStatus(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)
	_, dc := getKubeClientset()

	type dataplaneResp struct {
		Name          string `json:"name"`
		CloudRegion   string `json:"cloud_region"`
		CloudType     string `json:"cloud_type"`
		DataplaneType string `json:"dataplane_type"`
		Version       string `json:"version"`
		Status        string `json:"status"`
	}

	dpObjList, err := dc.Resource(dpGVK).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	for _, dpObj := range dpObjList.Items {
		if dpObj.GetName() == vars["dataplane_name"] {
			status, _, _ := unstructured.NestedString(dpObj.Object, "status", "phase")

			newdataplaneResp := dataplaneResp{
				Name:          vars["dataplane_name"],
				CloudRegion:   dpObj.GetLabels()["cloud_region"],
				DataplaneType: dpObj.GetLabels()["dataplane_type"],
				CloudType:     dpObj.GetLabels()["cloud_type"],
				Version:       dpObj.GetLabels()["version"],
				Status:        status,
			}

			dpResp, err := json.Marshal(newdataplaneResp)
			if err != nil {
				res := NewResponse(DataPlaneGetFail, string(JsonMarshallError), err, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
				return
			}
			sendJsonResponse(dpResp, http.StatusOK, &w)
		} else {
			sendJsonResponse([]byte("[]"), http.StatusOK, &w)
		}
	}
}

func ListDataPlane(w http.ResponseWriter, req *http.Request) {
	_, dc := getKubeClientset()

	listDp, err := dc.Resource(dpGVK).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	type dataplaneListResp struct {
		Name          string   `json:"name"`
		CloudRegion   string   `json:"cloud_region"`
		CloudType     string   `json:"cloud_type"`
		Customers     []string `json:"customers"`
		DataplaneType string   `json:"dataplane_type"`
		Version       string   `json:"version"`
		Status        string   `json:"status"`
	}

	var dpListResp []dataplaneListResp

	for _, dp := range listDp.Items {
		phase, _, _ := unstructured.NestedString(dp.Object, "status", "phase")

		newDpList := dataplaneListResp{
			Name:          dp.GetName(),
			CloudRegion:   dp.GetLabels()["cloud_region"],
			CloudType:     dp.GetLabels()["cloud_type"],
			Customers:     labels2Slice(dp.GetLabels()),
			DataplaneType: dp.GetLabels()["dataplane_type"],
			Version:       dp.GetLabels()["version"],
			Status:        phase,
		}
		dpListResp = append(dpListResp, newDpList)
	}

	bytes, _ := json.Marshal(dpListResp)
	sendJsonResponse(bytes, http.StatusOK, &w)

}

func DeleteDataPlane(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	_, dc := getKubeClientset()

	dpObjList, err := dc.Resource(dpGVK).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		res := NewResponse(DataPlaneGetFail, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	for _, dpObj := range dpObjList.Items {
		if dpObj.GetName() == vars["dataplane_name"] {
			exists := checkKeyInMap("customer_", dpObj.GetLabels())
			if exists {
				res := NewResponse(DataplaneDeletionFailedCustomerExists, internal_error, err, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
				return
			}
			err = dc.Resource(dpGVK).Namespace(dpObj.GetNamespace()).Delete(context.TODO(), dpObj.GetName(), metav1.DeleteOptions{})
			if err != nil {
				res := NewResponse(DataplaneDeletionFailed, internal_error, err, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
				return
			}
			res := NewResponse("", string(DataplaneDeletionInitiated), nil, http.StatusOK)
			res.SetResponse(&w)
			sendEventParseable(dataplanesEventStream, dataplaneTerminationEvent, dpObj.GetLabels(), map[string]string{"dataplane_name": dpObj.GetName()})
		}
	}

}
