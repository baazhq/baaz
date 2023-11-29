package tenantsinfra

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

func makeTenantInfraPath(dataplaneName string) string {
	return common.GetBzUrl() + common.BaazPath + common.DataplanePath + "/" + dataplaneName + common.TenantInfraPath
}

type Ti struct {
	TenantsInfra []struct {
		Name        string `yaml:"name" json:"name"`
		MachinePool []struct {
			Name   string `yaml:"name" json:"name"`
			Size   string `yaml:"size" json:"size"`
			Min    int    `yaml:"min" json:"min"`
			Max    int    `yaml:"max" json:"max"`
			Labels struct {
				App  string `yaml:"app" json:"app"`
				Size string `yaml:"size" json:"size"`
			} `yaml:"labels" json:"labels"`
		} `yaml:"machinePool" json:"machine_pool"`
	} `yaml:"tenantsInfra"`
}

// func GetTenantSizes() error {
// 	ts, err := listTenantSize()
// 	if err != nil {
// 		return err
// 	}

// 	table := tablewriter.NewWriter(os.Stdout)
// 	table.SetHeader([]string{
// 		"Name",
// 		"Machine_Pool",
// 	},
// 	)

// 	table.SetRowLine(true)
// 	table.Append([]string{"", "Name", "Size", "Min", "Max", "Labels"})

// 	for _, ts := range ts.TenantSizes {
// 		var row []string
// 		for _, mp := range ts.MachinePool {
// 			row = []string{
// 				ts.Name,
// 				mp.Name,
// 				mp.Size,
// 				strconv.Itoa(mp.Min),
// 				strconv.Itoa(mp.Max),
// 				common.CreateKeyValuePairs(mp.Labels),
// 			}
// 			table.SetRowLine(true)
// 			table.Append(row)
// 			table.SetAutoMergeCellsByColumnIndex([]int{0})

// 			table.SetAlignment(1)
// 		}

// 	}
// 	table.SetAlignment(1)

// 	table.Render()
// 	return nil
// }

// func listTenantSize() (tenantInfra, error) {
// 	response, err := http.Get(makeTenantSizePath())
// 	if err != nil {
// 		return tenantSizes{}, err
// 	}

// 	defer response.Body.Close()

// 	body, err := io.ReadAll(response.Body)
// 	if response.StatusCode > 299 {
// 		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", response.StatusCode, body)
// 	}
// 	if err != nil {
// 		return tenantSizes{}, err
// 	}

// 	var ts tenantSizes

// 	if string(body) == "[]" {
// 		return tenantSizes{}, nil
// 	}

// 	err = json.Unmarshal(body, &ts)
// 	if err != nil {
// 		return tenantSizes{}, err
// 	}

//		return ts, nil
//	}
func CreateTenantsInfra(filePath string, dataplane string) (string, error) {

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var ti Ti

	err = yaml.Unmarshal(yamlFile, &ti)
	if err != nil {
		return "", err
	}

	tiByte, err := json.Marshal(ti.TenantsInfra)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		makeTenantInfraPath(dataplane),
		"application/json",
		bytes.NewBuffer(tiByte),
	)
	if err != nil {
		return "", err
	}

	fmt.Println(string(tiByte))
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	fmt.Print(resp.Status)
	if resp.StatusCode > 299 {
		return "", fmt.Errorf("%s", string(body))
	}

	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		return "Tenant Infra Creation Initiated Successfully", nil
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
