package tenants

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type Tenants struct {
	Tenants struct {
		NetworkSecurity struct {
			InterNamespaceTraffic string   `yaml:"interNamespaceTraffic" json:"inter_namespace_traffic"`
			AllowedNamespaces     []string `yaml:"allowedNamespaces" json:"allowed_namespaces"`
		} `yaml:"networkSecurity" json:"network_security"`
		Application struct {
			Name    string `yaml:"name" json:"name"`
			AppSize string `yaml:"appSize" json:"app_size"`
		} `yaml:"application" json:"application"`
	} `yaml:"tenants" json:"tenants"`
}

func makeCreateTenantPath(customerName, tenantName string) string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + common.TenantPath + "/" + tenantName
}

func makeUpdateTenantPath(customerName, tenantName string) string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + common.TenantPath + "/" + tenantName
}

func makeDeleteTenantPath(customerName, tenantName string) string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + common.TenantPath + "/" + tenantName
}

func makeGetTenantPath(customerName string) string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + "/" + common.TenantPath
}

func GetTenants(customerName string) error {
	tenants, err := getTenants(customerName)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Tenant_Name",
		"Customer_Name",
		"Dataplane_Name",
		"Application_Name",
		"Application_Size",
	},
	)

	for _, tenant := range tenants {
		row := []string{
			tenant["tenant"].(string),
			tenant["customer"].(string),
			tenant["dataplane"].(string),
			tenant["application"].(string),
			tenant["size"].(string),
		}
		table.SetRowLine(true)
		table.Append(row)
		table.SetAlignment(1)

	}

	table.Render()
	return nil

}
func getTenants(customerName string) ([]map[string]interface{}, error) {
	resp, err := http.Get(
		makeGetTenantPath(customerName),
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("%s", string(body))
	}

	var tresp []map[string]interface{}
	err = json.Unmarshal(body, &tresp)
	if err != nil {
		return nil, err
	}

	return tresp, nil
}

func CreateTenant(filePath, customerName, tenantName string) (string, error) {
	yamlByte, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var tenants Tenants

	err = yaml.Unmarshal(yamlByte, &tenants)
	if err != nil {
		return "", err
	}

	tenantByte, err := json.Marshal(tenants.Tenants)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		makeCreateTenantPath(customerName, tenantName),
		"application/json",
		bytes.NewBuffer(tenantByte),
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
		return "Tenant Created Successfully", nil
	}

	return "", nil

}

func DeleteTenant(customerName, tenantName string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodDelete,
		makeDeleteTenantPath(customerName, tenantName),
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
		return "Tenant Deletion Initiated Successfully", nil
	}

	return "", nil
}

func UpdateTenant(filePath, customerName, tenantName string) (string, error) {
	yamlByte, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var tenants Tenants

	err = yaml.Unmarshal(yamlByte, &tenants)
	if err != nil {
		return "", err
	}

	tenantByte, err := json.Marshal(tenants.Tenants)
	if err != nil {
		return "", err
	}

	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodPut,
		makeUpdateTenantPath(customerName, tenantName),
		bytes.NewBuffer(tenantByte),
	)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
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
		return "Tenant Updated Successfully", nil
	}

	return "", nil

}
