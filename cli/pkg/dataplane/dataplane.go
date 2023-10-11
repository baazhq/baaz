package dataplane

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
)

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
		CloudType:        viper.GetString("dataplane.cloudType"),
		CloudRegion:      viper.GetString("dataplane.cloudRegion"),
		CloudAuth:        viper.GetStringMap("dataplane.cloudAuth"),
		KubernetesConfig: viper.GetStringMap("dataplane.kubernetesConfig"),
	}

	ccByte, err := json.Marshal(newCreateDataplane)
	if err != nil {
		return "", err
	}

	fmt.Println(string(ccByte))

	// resp, err := http.Post(
	// 	makePostCustomerPath(newCreateCustomer.Name),
	// 	"application/json",
	// 	bytes.NewBuffer(ccByte),
	// )
	// if err != nil {
	// 	return "", err
	// }

	// defer resp.Body.Close()

	// body, err := io.ReadAll(resp.Body)
	// if resp.StatusCode > 299 {
	// 	log.Fatalf("Response failed with status code: %d and\nbody: %s\n", resp.StatusCode, string(body))
	// }

	// if err != nil {
	// 	return "", err
	// }
	// if resp.StatusCode == http.StatusOK {
	// 	return "Customer Created Successfully", nil
	// }

	return "", nil
}
