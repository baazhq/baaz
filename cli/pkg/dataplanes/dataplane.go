package dataplanes

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

var (
	baazPath      = "/api/v1"
	dataplanePath = "/dataplane"
	customerPath  = "/customer"
)

type dpList struct {
	Name        string   `json:"name"`
	CloudRegion string   `json:"cloud_region"`
	CloudType   string   `json:"cloud_type"`
	Customers   []string `json:"customers"`
	SaaSType    string   `json:"saas_type"`
	Version     string   `json:"version"`
	Status      string   `json:"status"`
}

func makeCreateDeleteDataplaneUrl(customerName string) string {
	return common.GetBzUrl() + baazPath + customerPath + "/" + customerName + dataplanePath
}

func makeListDataplaneUrl() string {
	return common.GetBzUrl() + baazPath + dataplanePath
}

func makeAddDataplaneUrl(customerName, dataplaneName string) string {
	return common.GetBzUrl() + baazPath + customerPath + "/" + customerName + dataplanePath + "/" + dataplaneName
}

func GetDataplanes() error {
	dpList, err := listDataplanes()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Dataplane_Name",
		"Cloud_Region",
		"Cloud_Type",
		"Customers",
		"SaaS_Type",
		"Version",
		"Status",
	},
	)

	for _, dp := range dpList {
		row := []string{
			dp.Name,
			dp.CloudRegion,
			dp.CloudType,
			strings.Join(dp.Customers, "\n"),
			dp.SaaSType,
			dp.Version,
			dp.Status,
		}
		table.SetRowLine(true)
		table.Append(row)
		table.SetAlignment(1)

	}

	table.Render()

	return nil

}
func listDataplanes() ([]dpList, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodGet,
		makeListDataplaneUrl(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("%s", string(body))
	}

	var dp []dpList
	err = json.Unmarshal(body, &dp)
	if err != nil {
		return nil, err
	}

	return dp, nil
}

func DeleteDataplane(customerName string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodDelete,
		makeCreateDeleteDataplaneUrl(customerName),
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

func AddDataplane(dataplaneName, customerName string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodPut,
		makeAddDataplaneUrl(customerName, dataplaneName),
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
		return "Dataplane Added to Customer", nil
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
		makeCreateDeleteDataplaneUrl(viper.GetString("dataplane.customer_name")),
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
