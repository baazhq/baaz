package http_handlers

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "datainfra.io/ballastdata/api/v1/types"
	"github.com/gorilla/mux"
)

const (
	req_error        string = "REQUEST_ERROR"
	internal_error   string = "INTERNAL_ERROR"
	success          string = "SUCCESS"
	shared_namespace string = "shared"

	dataplane_creation_initated string = "Dataplane Creation Initiated"
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

	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
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
		res.LogResponse()
		return
	}

	if ns != nil {
		msg := "Customer namespace exists"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
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

	dpSecret := getAwsEksSecret(dataPlaneName, dataplane)

	_, err = dc.Resource(secretGVK).Namespace(shared_namespace).Create(context.TODO(), dpSecret, metav1.CreateOptions{})
	if err != nil {
		msg := "create data plane secret failed"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	dpDeploy := getAwsEksConfig(dataPlaneName, dataplane)

	_, err = dc.Resource(dpGVK).Namespace(shared_namespace).Create(context.TODO(), dpDeploy, metav1.CreateOptions{})
	if err != nil {
		msg := "create data plane config failed"
		res := NewResponse(msg, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(dataplane_creation_initated, success, nil, 200)
	res.SetResponse(&w)

}
