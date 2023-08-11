package http_handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "datainfra.io/ballastdata/api/v1/types"
	"github.com/gorilla/mux"
)

const (
	req_error      string = "REQUEST_ERROR"
	internal_error string = "INTERNAL_ERROR"
	success        string = "SUCCESS"
)

func CreateCustomer(w http.ResponseWriter, req *http.Request) {

	vars := mux.Vars(req)

	customerName := vars["name"]

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

	var customer v1.Customer

	if err := json.Unmarshal(body, &customer); err != nil {
		msg := "Unmarshall error"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	client, _ := getKubeClientset()

	_, err = client.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		labels := map[string]string{
			"description": strings.ReplaceAll(customer.Description, " ", "_"),
			"saas_type":   string(customer.SaaSType),
		}

		_, err := client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   customerName,
				Labels: mergeMaps(labels, customer.Labels),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			msg := "Create Namespace error"
			res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}
		res := NewResponse("Namespace Created for Customer", success, nil, 200)
		res.SetResponse(&w)
	}
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
		CloudType:   dp.CloudType,
		CloudRegion: dp.CloudRegion,
		CloudAuth: v1.CloudAuth{
			SecretRef: v1.SecretRef{
				SecretName:    dp.CloudAuth.SecretRef.SecretName,
				AccessKeyName: dp.CloudAuth.SecretRef.AccessKeyName,
				SecretKeyName: dp.CloudAuth.SecretRef.SecretKeyName,
			},
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

	var dpRes = schema.GroupVersionResource{
		Group:    "datainfra.io",
		Version:  "v1",
		Resource: "dataplanes",
	}

	fmt.Println(dataplane)

	dpDeploy := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "datainfra.io/v1",
			"kind":       "DataPlanes",
			"metadata": map[string]interface{}{
				"name": dataPlaneName,
			},
			"spec": map[string]interface{}{
				"saasType": dataplane.SaaSType,
				"cloudInfra": map[string]interface{}{
					"cloudType": dataplane.CloudType,
					"region":    dataplane.CloudRegion,
					"authSecretRef": map[string]interface{}{
						"secretName":    dataplane.CloudAuth.SecretRef.SecretName,
						"accessKeyName": dataplane.CloudAuth.SecretRef.AccessKeyName,
						"secretKeyName": dataplane.CloudAuth.SecretRef.SecretKeyName,
					},
					"eks": map[string]interface{}{
						"name":             dataPlaneName,
						"subnetIds":        dataplane.KubeConfig.EKS.SubnetIds,
						"securityGroupIds": dataplane.KubeConfig.EKS.SecurityGroupIds,
						"version":          dataplane.KubeConfig.EKS.Version,
					},
				},
			},
		},
	}

	_, err = dc.Resource(dpRes).Namespace("shared").Create(context.TODO(), dpDeploy, metav1.CreateOptions{})
	if err != nil {
		msg := "create data plane config failed"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse("create dp ", success, nil, 200)
	res.SetResponse(&w)

}
