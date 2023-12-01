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

type dpList struct {
	Name          string   `json:"name"`
	CloudRegion   string   `json:"cloud_region"`
	CloudType     string   `json:"cloud_type"`
	Customers     []string `json:"customers"`
	DataplaneType string   `json:"dataplane_type"`
	Version       string   `json:"version"`
	Status        string   `json:"status"`
}

type action struct {
	Action string `json:"action"`
}

func makeDataplaneUrl() string {
	return common.GetBzUrl() + common.BaazPath + common.DataplanePath
}

func makeListDataplaneUrl() string {
	return common.GetBzUrl() + common.BaazPath + common.DataplanePath
}

func makeAddRemoveDataplaneUrl(customerName, dataplaneName string) string {
	return common.GetBzUrl() + common.BaazPath + common.DataplanePath + "/" + dataplaneName + common.CustomerPath + "/" + customerName
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
		"Dataplane_Type",
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
			dp.DataplaneType,
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
		makeDataplaneUrl(),
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

	action := action{
		Action: "add",
	}

	actionByte, err := json.Marshal(action)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(
		http.MethodPut,
		makeAddRemoveDataplaneUrl(customerName, dataplaneName),
		bytes.NewReader(actionByte),
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

func RemoveDataplane(dataplaneName, customerName string) (string, error) {
	client := &http.Client{}

	action := action{
		Action: "remove",
	}

	actionByte, err := json.Marshal(action)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(
		http.MethodPut,
		makeAddRemoveDataplaneUrl(customerName, dataplaneName),
		bytes.NewReader(actionByte),
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
		return "Dataplane Removed from Customer", nil
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
		CloudType        string                 `json:"cloud_type"`
		CloudRegion      string                 `json:"cloud_region"`
		CloudAuth        map[string]interface{} `json:"cloud_auth"`
		KubernetesConfig map[string]interface{} `json:"kubernetes_config"`
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

	resp, err := http.Post(
		makeDataplaneUrl(),
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
