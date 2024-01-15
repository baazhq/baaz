package kubeconfig

import (
	"bz/pkg/common"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeConfig struct {
	APIVersion  string   `json:"apiVersion" yaml:"apiVersion"`
	Kind        string   `json:"kind" yaml:"kind"`
	Preferences struct{} `json:"preferences" yaml:"preferences"`
	Clusters    []struct {
		Cluster struct {
			CertificateAuthorityData string `json:"certificate-authority-data" yaml:"certificate-authority-data"`
			Server                   string `json:"server" yaml:"server"`
		} `json:"cluster" yaml:"cluster"`
		Name string `json:"name" yaml:"name"`
	} `json:"clusters" yaml:"clusters"`
	Users []struct {
		Name string `json:"name" yaml:"name"`
		User struct {
			AsUserExtra   struct{}    `json:"as-user-extra" yaml:"as-user-extra"`
			ClientKeyData interface{} `json:"client-key-data" yaml:"client-key-data"`
			Token         string      `json:"token" yaml:"token"`
		} `json:"user" yaml:"user"`
	} `json:"users" yaml:"users"`
	Contexts []struct {
		Context struct {
			Cluster   string `json:"cluster" yaml:"cluster"`
			Namespace string `json:"namespace" yaml:"namespace"`
			User      string `json:"user" yaml:"user"`
		} `json:"context" yaml:"context"`
		Name string `json:"name" yaml:"name"`
	} `json:"contexts" yaml:"contexts"`
	CurrentContext string `json:"current-context" yaml:"current-context"`
}

func makeGetCustomerKubeConfigPath(customerName string) string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + common.KubeConfigPath
}

func GetCustomerKubeConfig(customerName string) (*KubeConfig, error) {

	response, err := http.Get(makeGetCustomerKubeConfigPath(customerName))
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if response.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", response.StatusCode, body)
	}
	if err != nil {
		return nil, err
	}

	var resp map[string]string

	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return nil, err
	}

	newKubeConfig := KubeConfig{
		APIVersion:  "v1",
		Kind:        "Config",
		Preferences: struct{}{},
		Clusters: []struct {
			Cluster struct {
				CertificateAuthorityData string `json:"certificate-authority-data" yaml:"certificate-authority-data"`
				Server                   string `json:"server" yaml:"server"`
			} `json:"cluster" yaml:"cluster"`
			Name string `json:"name" yaml:"name"`
		}{
			{
				Cluster: struct {
					CertificateAuthorityData string `json:"certificate-authority-data" yaml:"certificate-authority-data"`
					Server                   string `json:"server" yaml:"server"`
				}{
					CertificateAuthorityData: resp["cluster_ca"],
					Server:                   "https://127.0.0.1:60646",
				},
				Name: resp["customer"] + "-cluster",
			},
		},
		Users: []struct {
			Name string `json:"name" yaml:"name"`
			User struct {
				AsUserExtra   struct{}    `json:"as-user-extra" yaml:"as-user-extra"`
				ClientKeyData interface{} `json:"client-key-data" yaml:"client-key-data"`
				Token         string      `json:"token" yaml:"token"`
			} `json:"user" yaml:"user"`
		}{
			{
				Name: resp["customer"],
				User: struct {
					AsUserExtra   struct{}    `json:"as-user-extra" yaml:"as-user-extra"`
					ClientKeyData interface{} `json:"client-key-data" yaml:"client-key-data"`
					Token         string      `json:"token" yaml:"token"`
				}{
					Token: resp["user_token_value"],
				},
			},
		},
		Contexts: []struct {
			Context struct {
				Cluster   string `json:"cluster" yaml:"cluster"`
				Namespace string `json:"namespace" yaml:"namespace"`
				User      string `json:"user" yaml:"user"`
			} `json:"context" yaml:"context"`
			Name string `json:"name" yaml:"name"`
		}{
			{
				Context: struct {
					Cluster   string `json:"cluster" yaml:"cluster"`
					Namespace string `json:"namespace" yaml:"namespace"`
					User      string `json:"user" yaml:"user"`
				}{
					Cluster:   resp["customer"] + "-cluster",
					Namespace: resp["namespace"],
					User:      resp["customer"],
				},
				Name: resp["customer"],
			},
		},
		CurrentContext: resp["customer"],
	}

	return &newKubeConfig, nil

}

func WriteKubeConfig2Cm(customerName string, config *KubeConfig, cs *kubernetes.Clientset) error {

	bytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	cmName := customerName + "-kubeconfig"
	cm := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      cmName,
			Namespace: customerName,
		},
		Data: map[string]string{
			cmName: string(bytes),
		},
	}

	_, err = cs.CoreV1().ConfigMaps(customerName).Create(context.TODO(), cm, v1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil

}
