package khota_handler

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"

	"net/http"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "datainfra.io/baaz/api/v1/types"
	"github.com/gorilla/mux"
)

type CustomerListResponse struct {
	Name      string            `json:"name"`
	SaaSType  string            `json:"saas_type"`
	CloudType string            `json:"cloud_type"`
	Status    string            `json:"status"`
	Dataplane string            `json:"dataplane"`
	Labels    map[string]string `json:"labels"`
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

			custLabels := getCustomLabel(ns.Labels)
			newCrListResp := CustomerListResponse{
				Name:      ns.Name,
				CloudType: ns.Labels["cloud_type"],
				SaaSType:  ns.Labels["saas_type"],
				Dataplane: ns.Labels["dataplane"],
				Status:    active,
				Labels:    custLabels,
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
	sendJsonResponse(resp, http.StatusOK, &w)
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
				Labels: mergeMaps(labels, setLabelPrefix(customer.Labels)),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			res := NewResponse(CustomerNamespaceCreateFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}
		// _, err = client.CoreV1().ServiceAccounts(customerName).Create(context.TODO(), &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{
		// 	Name:      customerName,
		// 	Namespace: customerName,
		// 	Labels:    mergeMaps(labels, setLabelPrefix(customer.Labels)),
		// }}, metav1.CreateOptions{})
		// if err != nil {
		// 	res := NewResponse(CustomerServiceAccountCreateFail, internal_error, err, http.StatusInternalServerError)
		// 	res.SetResponse(&w)
		// 	res.LogResponse()
		// 	return
		// }

		if err := createSaToken(client, customerName); err != nil {
			res := NewResponse(CustomerServiceAccountCreateFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}
		// customer_secret, err := Base62Random(50)
		// if err != nil {
		// 	klog.Errorf("ErrMsg: Base62 encoder not able to generate UUID ", err)
		// 	return
		// }
		// // create secret for customer
		// _, err = client.CoreV1().Secrets(customerName).Create(context.TODO(), &corev1.Secret{
		// 	ObjectMeta: metav1.ObjectMeta{
		// 		Name:      customerName,
		// 		Namespace: customerName,
		// 	},
		// 	StringData: map[string]string{
		// 		"token": customer_secret,
		// 	},
		// 	Type: corev1.SecretTypeOpaque,
		// }, metav1.CreateOptions{})

		// if err != nil {
		// 	// remove namespace if secret token creation fails
		// 	namespace_deletion_policy := metav1.DeletePropagationBackground
		// 	var grace_period int64 = 0
		// 	ns_err := client.CoreV1().Namespaces().Delete(context.TODO(), customerName, metav1.DeleteOptions{
		// 		PropagationPolicy:  &namespace_deletion_policy,
		// 		GracePeriodSeconds: &grace_period,
		// 	})
		// 	// returns error if namespace cleanup fails
		// 	if ns_err != nil {
		// 		res := NewResponse(CustomerNamespaceCreateFail, internal_error, ns_err, http.StatusInternalServerError)
		// 		res.SetResponse(&w)
		// 		res.LogResponse()
		// 		return
		// 	}
		// 	// returns error if secret token creation fails
		// 	res := NewResponse(CustomerNamespaceCreateFail, internal_error, err, http.StatusInternalServerError)
		// 	res.SetResponse(&w)
		// 	res.LogResponse()
		// 	return
		// }
		res := NewResponse(CustomerNamespaceSuccess, success, nil, 200)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if ns != nil {
		res := NewResponse(CustomerNamespaceExists, duplicate_entry, err, http.StatusConflict)
		res.SetResponse(&w)
		res.LogResponse()
	}

}

func UpdateCustomer(w http.ResponseWriter, req *http.Request) {

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
		res := NewResponse(CustomerNamespaceDoesNotExists, entry_not_exists, err, http.StatusNotFound)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if ns != nil {
		for key, val := range setLabelPrefix(customer.Labels) {
			ns.Labels[key] = val
		}
		_, err := client.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{})

		if err != nil {
			res := NewResponse(CustomerNamespaceUpdateFail, internal_error, err, http.StatusInternalServerError)
			res.SetResponse(&w)
			res.LogResponse()
			return
		}
		res := NewResponse(CustomerNamespaceUpdateSuccess, success, nil, 200)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

}

func setLabelPrefix(data map[string]string) map[string]string {
	modified_labels := make(map[string]string)
	for key, val := range data {
		modified_labels[label_prefix+key] = val
	}
	return modified_labels
}

func getCustomLabel(data map[string]string) map[string]string {
	filtered_map := make(map[string]string)
	for key, val := range data {
		if key[:len(label_prefix)] == label_prefix {
			filtered_map[strings.TrimPrefix(key, label_prefix)] = val
		}
	}
	return filtered_map
}
