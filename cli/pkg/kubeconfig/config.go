package kubeconfig

import (
	"bz/pkg/common"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type KubeConfig struct {
	APIVersion string `yaml:"apiVersion" json:"apiVersion"`
	Kind       string `yaml:"kind" json:"kind"`
	Clusters   []struct {
		Name    string `yaml:"name" json:"name"`
		Cluster struct {
			CertificateAuthorityData string `yaml:"certificate-authority-data" json:"certificate-authority-data"`
			Server                   string `yaml:"server" json:"server"`
		} `yaml:"cluster" json:"cluster"`
	} `yaml:"clusters" json:"clusters"`
	Contexts []struct {
		Name    string `yaml:"name" json:"name"`
		Context struct {
			Cluster   string `yaml:"cluster" json:"cluster"`
			Namespace string `yaml:"namespace" json:"namespace"`
			User      string `yaml:"user" json:"user"`
		} `yaml:"context" json:"context"`
	} `yaml:"contexts" json:"contexts"`
	CurrentContext string `yaml:"current-context" json:"current-contexts"`
	Users          []struct {
		Name string `yaml:"name" json:"name"`
		User struct {
			Token string `yaml:"token" json:"token"`
		} `yaml:"user" json:"user"`
	} `yaml:"users" json:"users"`
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
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: resp["current_context"],
		Clusters: []struct {
			Name    string `yaml:"name" json:"name"`
			Cluster struct {
				CertificateAuthorityData string `yaml:"certificate-authority-data" json:"certificate-authority-data"`
				Server                   string `yaml:"server" json:"server"`
			} `yaml:"cluster" json:"cluster"`
		}{
			{
				Name: resp["current_context"],
				Cluster: struct {
					CertificateAuthorityData string `yaml:"certificate-authority-data" json:"certificate-authority-data"`
					Server                   string `yaml:"server" json:"server"`
				}{
					CertificateAuthorityData: base64.StdEncoding.EncodeToString([]byte(resp["cluster_ca"])),
					Server:                   resp["cluster_server"],
				},
			},
		},
		Contexts: []struct {
			Name    string `yaml:"name" json:"name"`
			Context struct {
				Cluster   string `yaml:"cluster" json:"cluster"`
				Namespace string `yaml:"namespace" json:"namespace"`
				User      string `yaml:"user" json:"user"`
			} `yaml:"context" json:"context"`
		}{
			{
				Name: resp["current_context"],
				Context: struct {
					Cluster   string `yaml:"cluster" json:"cluster"`
					Namespace string `yaml:"namespace" json:"namespace"`
					User      string `yaml:"user" json:"user"`
				}{
					Cluster:   resp["current_context"],
					Namespace: resp["namespace"],
					User:      resp["current_context"],
				},
			},
		},
		Users: []struct {
			Name string `yaml:"name" json:"name"`
			User struct {
				Token string `yaml:"token" json:"token"`
			} `yaml:"user" json:"user"`
		}{
			{
				Name: resp["current_context"],
				User: struct {
					Token string `yaml:"token" json:"token"`
				}{
					Token: resp["user_token_value"],
				},
			},
		},
	}

	return &newKubeConfig, nil

}
