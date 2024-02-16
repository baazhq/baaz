package dataplanes

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type Dataplane struct {
	Dataplane struct {
		CloudType    string `yaml:"cloudType" json:"cloud_type"`
		CloudRegion  string `yaml:"cloudRegion" json:"cloud_region"`
		CustomerName string `yaml:"customerName" json:"customer_name"`
		SaasType     string `yaml:"saasType" json:"saas_type"`
		CloudAuth    struct {
			AwsAuth struct {
				AwsAccessKey string `yaml:"awsAccessKey" json:"aws_access_key"`
				AwsSecretKey string `yaml:"awsSecretKey" json:"aws_secret_key"`
			} `yaml:"awsAuth" json:"aws_auth"`
		} `yaml:"cloudAuth" json:"cloud_auth"`
		ProvisionNetwork bool `yaml:"provisionNetwork" json:"provision_network"`
		KubernetesConfig struct {
			Eks struct {
				SubnetIds        []string `yaml:"subnetIds" json:"subnet_ids"`
				SecurityGroupIds []string `yaml:"securityGroupIds" json:"security_group_ids"`
				Version          string   `yaml:"version" json:"version"`
			} `yaml:"eks"`
		} `yaml:"kubernetesConfig" json:"kubernetes_config"`
		ApplicationConfig []struct {
			Name      string   `yaml:"name" json:"name"`
			Namespace string   `yaml:"namespace" json:"namespace"`
			ChartName string   `yaml:"chartName" json:"chart_name"`
			RepoName  string   `yaml:"repoName" json:"repo_name"`
			RepoURL   string   `yaml:"repoUrl" json:"repo_url"`
			Version   string   `yaml:"version" json:"version"`
			Values    []string `yaml:"values,omitempty" json:"values,omitempty"`
		} `yaml:"applicationConfig" json:"application_config"`
	} `yaml:"dataplane" json:"dataplane"`
}

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

	yamlByte, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var dataplane Dataplane

	err = yaml.Unmarshal(yamlByte, &dataplane)
	if err != nil {
		return "", err
	}

	fmt.Println(dataplane.Dataplane.ProvisionNetwork)

	dataplaneByte, err := json.Marshal(dataplane.Dataplane)
	if err != nil {
		return "", err
	}

	// fmt.Println(string(dataplaneByte))

	resp, err := http.Post(
		makeDataplaneUrl(),
		"application/json",
		bytes.NewBuffer(dataplaneByte),
	)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		return "", fmt.Errorf("%s", string(respBody))
	}

	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		return "Dataplane Created Successfully", nil
	}

	return "", nil
}
