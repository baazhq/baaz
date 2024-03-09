package khota_handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/baazhq/baaz/api/v1/types"
	helm "github.com/baazhq/baaz/pkg/helmchartpath"
	"github.com/gorilla/mux"
)

const (
	labelPrefix          = "baaz_"       // Prefix for labels
	dataplaneUnavailable = "unavailable" // Unavailable dataplane status
)

type CustomerListResponse struct {
	Name      string            `json:"name"`
	SaaSType  string            `json:"saas_type"`
	CloudType string            `json:"cloud_type"`
	Status    string            `json:"status"`
	Dataplane string            `json:"dataplane"`
	Labels    map[string]string `json:"labels"`
}

// ListCustomer handles listing customers
func ListCustomer(w http.ResponseWriter, req *http.Request) {
	client, _ := getKubeClientset()
	nsList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "controlplane=baaz",
	})
	if err != nil {
		handleError(w, err, CustomerNamespaceListEmpty, http.StatusInternalServerError)
		return
	}

	var customerListResponse []CustomerListResponse

	for _, ns := range nsList.Items {
		ns, err := client.CoreV1().Namespaces().Get(context.TODO(), ns.Name, metav1.GetOptions{})
		if err != nil {
			handleError(w, err, CustomerNamespaceGetFail, http.StatusInternalServerError)
			return
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

	resp, err := json.Marshal(customerListResponse)
	if err != nil {
		handleError(w, err, JsonMarshallError, http.StatusInternalServerError)
		return
	}

	sendJsonResponse(resp, http.StatusOK, &w)
}

// CreateCustomer handles creating a customer
func CreateCustomer(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	customerName := vars["customer_name"]

	body, err := io.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		handleError(w, err, ServerReqSizeExceed, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var customer v1.Customer
	if err := json.Unmarshal(body, &customer); err != nil {
		handleError(w, err, ServerUnmarshallError, http.StatusInternalServerError)
		return
	}

	client, _ := getKubeClientset()
	_, err = client.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		labels := map[string]string{
			"saas_type":     string(customer.SaaSType),
			"cloud_type":    string(customer.CloudType),
			"customer_name": string(customerName),
			"dataplane":     dataplaneUnavailable,
			"controlplane":  "baaz",
		}

		allLabels := mergeMaps(labels, setLabelPrefix(customer.Labels))
		_, err := client.CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   customerName,
				Labels: allLabels,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			handleError(w, err, CustomerNamespaceCreateFail, http.StatusInternalServerError)
			return
		}

		helmBuild := helm.NewHelm(
			customerName+"-customer",
			customerName,
			customer_chartpath,
			nil,
			[]string{
				"customer.name=" + customerName,
				"customer.labels.saas_type=" + string(customer.SaaSType),
				"customer.labels.cloud_type=" + string(customer.CloudType),
				"customer.labels.customer_name=" + string(customerName),
				"customer.labels.dataplane=" + dataplaneUnavailable,
				"customer.labels.controlplane=" + "baaz",
			},
		)
		if err := helmBuild.Apply(); err != nil {
			handleError(w, err, CustomerNamespaceCreateFail, http.StatusInternalServerError)
			return
		}

		handleSuccess(w, CustomerNamespaceSuccess, http.StatusOK)
		sendEventParseable(customerEventStream, customerCreateSuccess, allLabels, map[string]string{"customer_name": customerName})
		return
	}

	handleError(w, err, CustomerNamespaceExists, http.StatusConflict)
}

// UpdateCustomer handles updating a customer
func UpdateCustomer(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	customerName := vars["customer_name"]

	body, err := io.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		handleError(w, err, ServerReqSizeExceed, http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var customer v1.Customer
	if err := json.Unmarshal(body, &customer); err != nil {
		handleError(w, err, ServerUnmarshallError, http.StatusInternalServerError)
		return
	}

	client, _ := getKubeClientset()
	ns, err := client.CoreV1().Namespaces().Get(context.TODO(), customerName, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		handleError(w, err, CustomerNamespaceDoesNotExists, http.StatusNotFound)
		return
	}

	for key, val := range setLabelPrefix(customer.Labels) {
		ns.Labels[key] = val
	}
	if _, err := client.CoreV1().Namespaces().Update(context.TODO(), ns, metav1.UpdateOptions{}); err != nil {
		handleError(w, err, CustomerNamespaceUpdateFail, http.StatusInternalServerError)
		return
	}

	handleSuccess(w, CustomerNamespaceUpdateSuccess, http.StatusOK)
}

// handleError logs and handles errors
func handleError(w http.ResponseWriter, err error, msg CustomMsg, code int) {
	res := NewResponse(msg, internal_error, err, code)
	res.SetResponse(&w)
	res.LogResponse()
}

// handleSuccess logs and handles success responses
func handleSuccess(w http.ResponseWriter, msg CustomMsg, code int) {
	res := NewResponse(msg, success, nil, code)
	res.SetResponse(&w)
	res.LogResponse()
}

// setLabelPrefix adds a prefix to all labels
func setLabelPrefix(data map[string]string) map[string]string {
	modifiedLabels := make(map[string]string)
	for key, val := range data {
		modifiedLabels[labelPrefix+key] = val
	}
	return modifiedLabels
}

// getCustomLabel extracts custom labels
func getCustomLabel(data map[string]string) map[string]string {
	filteredMap := make(map[string]string)
	for key, val := range data {
		if strings.HasPrefix(key, labelPrefix) {
			filteredMap[strings.TrimPrefix(key, labelPrefix)] = val
		}
	}
	return filteredMap
}
