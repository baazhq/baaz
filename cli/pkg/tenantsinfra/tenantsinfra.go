package tenantsinfra

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

func makeTenantInfraPath(dataplaneName string) string {
	return common.GetBzUrl() + common.BaazPath + common.DataplanePath + "/" + dataplaneName + common.TenantInfraPath
}

func makeTenantInfraPathWithName(dataplaneName, tenantInfraName string) string {
	return common.GetBzUrl() + common.BaazPath + common.DataplanePath + "/" + dataplaneName + common.TenantInfraPath + "/" + tenantInfraName
}

// tenantsInfra:
//   foo-small:
//     machinePool:
//     - name: foo-app1
//       size: t2.small
//       min: 1
//       max: 3
//       labels:
//         app: iot
//         size: small
//   boo-medium:
//     machinePool:
//     - name: boo-app1
//       size: t2.medium
//       min: 1
//       max: 3
//       labels:
//         app: iot
//         size: small

type TiMachine struct {
	MachinePool []struct {
		Name             string `yaml:"name" json:"name"`
		Size             string `yaml:"size" json:"size"`
		Min              int    `yaml:"min" json:"min"`
		Max              int    `yaml:"max" json:"max"`
		StrictScheduling string `yaml:"strictScheduling" json:"strictScheduling"`
		Type             string `yaml:"type" json:"type"`
		Labels           struct {
			App  string `yaml:"app" json:"app"`
			Size string `yaml:"size" json:"size"`
		} `yaml:"labels" json:"labels"`
	} `yaml:"machinePool" json:"machine_pool"`
}

type Ti struct {
	TenantsInfra map[string]TiMachine `yaml:"tenantsInfra"`
}

type MachinePoolStatus struct {
	Status string `json:"status"`
	Subnet string `json:"subnet"`
}

type MachinePool struct {
	Labels           map[string]string `json:"labels"`
	Max              int               `json:"max"`
	Min              int               `json:"min"`
	Name             string            `json:"name"`
	Size             string            `json:"size"`
	StrictScheduling string            `json:"strictScheduling"`
	Type             string            `json:"type"`
}

type TenantInfra struct {
	Name              string                       `json:"name"`
	DataPlane         string                       `json:"dataplane"`
	MachinePoolStatus map[string]MachinePoolStatus `json:"machine_pool_status"`
	TenantSizes       map[string]struct {
		MachinePool []MachinePool `json:"machinePool"`
	} `json:"tenant_sizes"`
	Status string `json:"status"`
}

func GetTenantsInfra(dataplane, tenantinfra_name string) error {
	ti, err := getTenantsInfra(dataplane, tenantinfra_name)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name",
		"Machine_Pool",
		"Size",
		"Min",
		"Max",
		"Labels",
		"Strict_Scheduling",
		"Type",
		"Status",
	})

	var uniqueNamePrinted bool

	for tenantSize, details := range ti.TenantSizes {
		for _, mp := range details.MachinePool {
			labels := []string{}
			for k, v := range mp.Labels {
				labels = append(labels, fmt.Sprintf("%s: %s", k, v))
			}
			status := ti.MachinePoolStatus[tenantSize+"-"+mp.Name+"-"+strings.ReplaceAll(mp.Size, ".", "-")].Status

			if !uniqueNamePrinted {
				table.Append([]string{
					ti.Name,
					mp.Name,
					mp.Size,
					fmt.Sprintf("%d", mp.Min),
					fmt.Sprintf("%d", mp.Max),
					strings.Join(labels, ", "),
					mp.StrictScheduling,
					mp.Type,
					status,
				})
				uniqueNamePrinted = true
			} else {
				table.Append([]string{
					"",
					mp.Name,
					mp.Size,
					fmt.Sprintf("%d", mp.Min),
					fmt.Sprintf("%d", mp.Max),
					strings.Join(labels, ", "),
					mp.StrictScheduling,
					mp.Type,
					status,
				})
			}
		}
	}

	table.Render()
	return nil
}

func getTenantsInfra(dataplane, tenantinfra_name string) (TenantInfra, error) {
	path := makeTenantInfraPath(dataplane)
	if tenantinfra_name != "" {
		path = makeTenantInfraPathWithName(dataplane, tenantinfra_name)
	}
	response, err := http.Get(path)
	if err != nil {
		return TenantInfra{}, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if response.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", response.StatusCode, body)
	}
	if err != nil {
		return TenantInfra{}, err
	}

	var ti []TenantInfra

	if string(body) == "[]" {
		return TenantInfra{}, nil
	}

	err = json.Unmarshal(body, &ti)
	if err != nil {
		return TenantInfra{}, err
	}

	if len(ti) == 0 {
		return TenantInfra{}, nil
	}

	return ti[0], nil
}

func CreateTenantsInfra(filePath string, dataplane string) (string, error) {

	yamlFile, err := os.ReadFile(filePath)
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

	fmt.Println(string(tiByte))

	fmt.Println(makeTenantInfraPath(dataplane))
	resp, err := http.Post(
		makeTenantInfraPath(dataplane),
		"application/json",
		bytes.NewBuffer(tiByte),
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
		return "Tenant Infra Creation Initiated Successfully", nil
	}

	return "", nil
}

func UpdateTenantsInfra(filePath string, dataplane, tenantInfra string) (string, error) {

	yamlFile, err := os.ReadFile(filePath)
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

	req, err := http.NewRequest(
		http.MethodPut,
		makeTenantInfraPathWithName(dataplane, tenantInfra),
		bytes.NewBuffer(tiByte),
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

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		return "", fmt.Errorf("%s", string(body))
	}

	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		return "Tenant Infra Update Initiated Successfully", nil
	}

	return "", nil
}

func DeleteTenantsInfra(dataplane, tenantInfra string) (string, error) {
	req, err := http.NewRequest(
		http.MethodDelete,
		makeTenantInfraPathWithName(dataplane, tenantInfra),
		nil,
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

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 299 {
		return "", fmt.Errorf("%s", string(body))
	}

	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusOK {
		return "Tenant Infra Delete Initiated Successfully", nil
	}

	return "", nil
}
