package tenantsinfra

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
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

func makeTenantInfraPath(dataplaneName string) string {
	return common.GetBzUrl() + common.BaazPath + common.DataplanePath + "/" + dataplaneName + common.TenantInfraPath
}

type MachinePoolStatus struct {
	Status string `json:"status"`
	Subnet string `json:"subnet"`
}

type MachinePool struct {
	Labels map[string]string `json:"labels"`
	Max    int               `json:"max"`
	Min    int               `json:"min"`
	Name   string            `json:"name"`
	Size   string            `json:"size"`
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

func GetTenantsInfra(dataplane string) error {
	ti, err := getTenantsInfra(dataplane)
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
			fmt.Println(status)
			fmt.Println(tenantSize + "-" + mp.Name + "-" + mp.Size)
			if !uniqueNamePrinted {
				table.Append([]string{
					ti.Name,
					mp.Name,
					mp.Size,
					fmt.Sprintf("%d", mp.Min),
					fmt.Sprintf("%d", mp.Max),
					strings.Join(labels, ", "),
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
					status,
				})
			}
		}
	}

	table.Render()
	return nil
}

func getTenantsInfra(dataplane string) (TenantInfra, error) {
	response, err := http.Get(makeTenantInfraPath(dataplane))
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

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var ti TenantInfra

	err = yaml.Unmarshal(yamlFile, &ti)
	if err != nil {
		return "", err
	}

	tiByte, err := json.Marshal(ti)
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
