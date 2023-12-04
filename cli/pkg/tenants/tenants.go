package tenants

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

func makeTenantPath(customerName, dataplaneName, tenantName string) string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + customerName + common.DataplanePath + "/" + dataplaneName + common.TenantPath + "/" + tenantName
}

func CreateTenant(filePath, customerName, dataplaneName, tenantName string) (string, error) {
	yamlByte, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var body interface{}

	err = yaml.Unmarshal(yamlByte, &body)
	if err != nil {
		return "", err
	}

	tenantByte, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		makeTenantPath(customerName, dataplaneName, tenantName),
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
