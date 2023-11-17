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

func makeTenantPath() string {
	return common.GetBzUrl() + common.BaazPath + common.CustomerPath + "/" + "customerName" + common.DataplanePath
}

func CreateTenant(filePath string) (string, error) {
	yamlByte, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var body interface{}

	err = yaml.Unmarshal(yamlByte, &body)
	if err != nil {
		return "", err
	}

	tsByte, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(
		"makeTenantSizePath()",
		"application/json",
		bytes.NewBuffer(tsByte),
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
		return "Tenant Sizes Created Successfully", nil
	}

	return "", nil

}
