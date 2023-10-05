package khota_handler

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "datainfra.io/baaz/api/v1/types"
	"github.com/gorilla/mux"
)

const (
	req_error                   string = "REQUEST_ERROR"
	internal_error              string = "INTERNAL_ERROR"
	success                     string = "SUCCESS"
	shared_namespace            string = "shared"
	dataplane_creation_initated string = "Dataplane Creation Initiated"
	active                      string = "ACTIVE"
	dataplane_unavailable       string = "UNAVAILABLE"
)

type CustomerListResponse struct {
	Name      string
	SaaSType  string
	CloudType string
	Status    string
	Dataplane string
}

func ListCustomer(w http.ResponseWriter, req *http.Request) {

	client, _ := getKubeClientset()
	nsList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "controlplane=baaz",
	})
	if apierrors.IsNotFound(err) {
		res := NewResponse(CustomerNamespaceListEmpty, internal_error, nil, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if nsList == nil {
		res := NewResponse(CustomerNamespaceListEmpty, success, nil, 204)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}
	var customerListResponse []CustomerListResponse

	if nsList != nil {
		for _, ns := range nsList.Items {
			ns, err := client.CoreV1().Namespaces().Get(context.TODO(), ns.Name, metav1.GetOptions{})
			if err != nil {
				res := NewResponse(CustomerNamespaceGetFail, internal_error, err, http.StatusInternalServerError)
				res.SetResponse(&w)
				res.LogResponse()
			}

			newCrListResp := CustomerListResponse{
				Name:      ns.Name,
				CloudType: ns.Labels["cloud_type"],
				SaaSType:  ns.Labels["saas_type"],
				Dataplane: ns.Labels["dataplane"],
				Status:    active,
			}

			customerListResponse = append(customerListResponse, newCrListResp)

		}

	}
	resp, err := json.Marshal(customerListResponse)
	if err != nil {
		res := NewResponse(JsonMarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
	}
	res := NewResponse(CustomMsg(resp), success, nil, http.StatusOK)
	res.SetResponse(&w)
	res.LogResponse()
}

func CreateCustomer(w http.ResponseWriter, req *http.Request) {

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

	var customer v1.Customer

	if err := json.Unmarshal(body, &customer); err != nil {
		res := NewResponse(ServerUnmarshallError, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	client, _ := getKubeClientset()

	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {

		labels := map[string]string{
			"saas_type":     string(customer.SaaSType),
			"cloud_type":    string(customer.CloudType),
			"customer_name": string(customerName),
			"dataplane":     dataplane_unavailable,
			"controlplane":  "baaz",
		}

		_, err := client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   customerName,
				Labels: mergeMaps(labels, customer.Labels),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			res := NewResponse(CustomerNamespaceGetFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}
		res := NewResponse(CustomerNamespaceSuccess, success, nil, 200)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if ns != nil {
		res := NewResponse(CustomerNamespaceExists, internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
	}
}