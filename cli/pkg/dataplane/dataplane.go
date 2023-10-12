package dataplane

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/viper"
)

var customerUrlPath = "/api/v1/customer"
var dataplanePath = "/dataplane"

func makePostDeleteDataplaneUrl(customerName string) string {
	return common.GetBzUrl() + customerUrlPath + "/" + customerName + dataplanePath
}

func DeleteDataplane(customerName string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodDelete,
		makePostDeleteDataplaneUrl(customerName),
		nil,
	)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		return "", fmt.Errorf("%s", string(body))
	}

	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		return "Dataplane Deletion Initiated Successfully", nil
	}

	return "", nil

}

func CreateDataplane(filePath string) (string, error) {

	viper.SetConfigFile(filePath)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return "", err
	}

	type createDataPlane struct {
		SaaSType         string                 `json:"saas_type"`
		CloudType        string                 `json:"cloud_type"`
		CloudRegion      string                 `json:"cloud_region"`
		CloudAuth        map[string]interface{} `json:"cloud_auth"`
		KubernetesConfig map[string]interface{} `json:"kubernetes_config"`
	}

	newCreateDataplane := createDataPlane{
		SaaSType:         viper.GetString("dataplane.saas_type"),
		CloudType:        viper.GetString("dataplane.cloud_type"),
		CloudRegion:      viper.GetString("dataplane.cloud_region"),
		CloudAuth:        viper.GetStringMap("dataplane.cloud_auth"),
		KubernetesConfig: viper.GetStringMap("dataplane.kubernetes_config"),
	}

	ccByte, err := json.Marshal(newCreateDataplane)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		makePostDeleteDataplaneUrl(viper.GetString("dataplane.customer_name")),
		"application/json",
		bytes.NewBuffer(ccByte),
	)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		return "", fmt.Errorf("%s", string(body))
	}

	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		return "Dataplane Creation Initiated Successfully", nil
	}

	return "", nil
}
