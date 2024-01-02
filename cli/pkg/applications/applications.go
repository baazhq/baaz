package applications

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

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
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + common.TenantPath + "/" + tenantName
}

func CreateApplication(filePath, customerName, tenantName string) (string, error) {

	yamlByte, err := ioutil.ReadFile(filePath)
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
