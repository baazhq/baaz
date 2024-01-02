package khota_handler

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeConfig struct {
	CurrentContext string `json:"current_context"`
	Customer       string `json:"customer"`
	Namespace      string `json:"namespace"`
	ClusterCA      string `json:"cluster_ca"`
	ClusterServer  string `json:"cluster_server"`
	UserTokenValue string `json:"user_token_value"`
}

func NewKubeConfig(
	customerName string,
	clientset *kubernetes.Clientset,
) (*KubeConfig, error) {
	secret, err := clientset.CoreV1().Secrets(customerName).Get(context.TODO(), customerName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &KubeConfig{
		CurrentContext: customerName + "-context",
		Customer:       customerName,
		Namespace:      string(secret.Data["namespace"]),
		ClusterCA:      b64.StdEncoding.EncodeToString([]byte(secret.Data["ca.crt"])),
		ClusterServer:  "https://127.0.0.1:60646",
		UserTokenValue: string(secret.Data["token"]),
	}, nil

}

func GetKubeConfig(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	customerName := vars["customer_name"]

	kc, _ := getKubeClientset()

	config, err := NewKubeConfig(customerName, kc)
	if err != nil {
		res := NewResponse(CustomMsg(ConfigGetFail), internal_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	bytes, _ := json.Marshal(config)
	sendJsonResponse(bytes, http.StatusOK, &w)

}
