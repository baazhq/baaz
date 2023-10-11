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

func makePostDataPlanePath(dataplaneName string) string {
	return common.GetBzUrl() + customerUrlPath + "/" + dataplaneName + dataplanePath
}

func CreateDataplane(filePath string) (string, error) {

	viper.SetConfigFile(filePath)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return "", err
	}

	type createDataPlane struct {
		CloudType        string                 `json:"saas_type"`
		CloudRegion      string                 `json:"cloud_type"`
		CloudAuth        map[string]interface{} `json:"cloud_auth"`
		KubernetesConfig map[string]interface{} `json:"kuberenetes_config"`
	}

	newCreateDataplane := createDataPlane{
		CloudType:        viper.GetString("dataplane.cloud_type"),
		CloudRegion:      viper.GetString("dataplane.cloud_region"),
		CloudAuth:        viper.GetStringMap("dataplane.cloud_auth"),
		KubernetesConfig: viper.GetStringMap("dataplane.kubernetes_config"),
	}

	ccByte, err := json.Marshal(newCreateDataplane)
	if err != nil {
		return "", err
	}

	fmt.Println(string(ccByte))

	resp, err := http.Post(
		makePostDataPlanePath(viper.GetString("dataplane.customer_name")),
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
