package applications

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Applications struct {
	Application []struct {
		Name      string   `yaml:"name" json:"name"`
		ChartName string   `yaml:"chartName" json:"chart_name"`
		RepoName  string   `yaml:"repoName" json:"repo_name"`
		RepoURL   string   `yaml:"repoUrl" json:"repo_url"`
		Version   string   `yaml:"version" json:"version"`
		Values    []string `yaml:"values" json:"values"`
	} `yaml:"application" json:"application"`
}

func makeApplicationUrl(customerName, tenantName string) string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + common.TenantPath + "/" + tenantName + common.Application
}

func makeUpdateApplicationUrl(customerName, applicationName string) string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + common.Application + "/" + applicationName
}

func CreateApplication(filePath, customerName, tenantName string) (string, error) {

	yamlByte, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var applications Applications

	err = yaml.Unmarshal(yamlByte, &applications)
	if err != nil {
		return "", err
	}

	appByte, err := json.Marshal(applications.Application)
	if err != nil {
		return "", err
	}

	fmt.Println(string(appByte))
	resp, err := http.Post(
		makeApplicationUrl(customerName, tenantName),
		"application/json",
		bytes.NewBuffer(appByte),
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
		return "Application Creation Initated Successfully", nil
	}

	return "", nil
}

func UpdateApplication(filePath, customerName, applicationName string) (string, error) {
	yamlByte, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var applications Applications

	err = yaml.Unmarshal(yamlByte, &applications)
	if err != nil {
		return "", err
	}

	appByte, err := json.Marshal(applications.Application)
	if err != nil {
		return "", err
	}

	fmt.Println(string(appByte))
	req, err := http.NewRequest(
		http.MethodPut,
		makeUpdateApplicationUrl(customerName, applicationName),
		bytes.NewBuffer(appByte),
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
		return "Application Update Initated Successfully", nil
	}

	return "", nil
}
