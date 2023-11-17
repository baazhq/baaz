package tenantsize

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

func makeTenantSizePath() string {
	return common.GetBzUrl() + common.BaazPath + common.TenantPath + common.TenantSizesPath
}

type tenantSizes struct {
	TenantSizes []struct {
		Name        string `json:"name"`
		MachinePool []struct {
			Name   string            `json:"name"`
			Size   string            `json:"size"`
			Min    int               `json:"min"`
			Max    int               `json:"max"`
			Labels map[string]string `json:"labels"`
		} `json:"machine_pool"`
	} `json:"tenant_sizes"`
}

func GetTenantSizes() error {
	ts, err := listTenantSize()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name",
		"Machine_Pool",
	},
	)

	table.SetRowLine(true)
	table.Append([]string{"", "Name", "Size", "Min", "Max", "Labels"})

	for _, ts := range ts.TenantSizes {
		var row []string
		for _, mp := range ts.MachinePool {
			row = []string{
				ts.Name,
				mp.Name,
				mp.Size,
				strconv.Itoa(mp.Min),
				strconv.Itoa(mp.Max),
				common.CreateKeyValuePairs(mp.Labels),
			}
			table.SetRowLine(true)
			table.Append(row)
			table.SetAutoMergeCellsByColumnIndex([]int{0})

			table.SetAlignment(1)
		}

	}
	table.SetAlignment(1)

	table.Render()
	return nil
}

func listTenantSize() (tenantSizes, error) {
	response, err := http.Get(makeTenantSizePath())
	if err != nil {
		return tenantSizes{}, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if response.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", response.StatusCode, body)
	}
	if err != nil {
		return tenantSizes{}, err
	}

	var ts tenantSizes

	if string(body) == "[]" {
		return tenantSizes{}, nil
	}

	err = json.Unmarshal(body, &ts)
	if err != nil {
		return tenantSizes{}, err
	}

	return ts, nil
}

func CreateTenantSize(filePath string) (string, error) {
	yamlByte, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var body interface{}

	err = yaml.Unmarshal(yamlByte, &body)
	if err != nil {
		return "", err
	}
	body = convert(body)

	tsByte, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(
		makeTenantSizePath(),
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

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
