package customers

import (
	"bz/pkg/common"
	"encoding/json"
	"io"
	"net/http"
	"os"

	log "k8s.io/klog/v2"

	"github.com/olekukonko/tablewriter"
)

var customersUrlPath = "/api/v1/customer"

type customerList struct {
	Name      string `json:"Name"`
	SaaSType  string `json:"SaaSType"`
	CloudType string `json:"CloudType"`
	Status    string `json:"Status"`
	Dataplane string `json:"Dataplane"`
}

func GetCustomers() error {
	customerList, err := getCustomerList()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Customer Name",
		"SaaSType",
		"CloudType",
		"Status",
		"Dataplane",
	},
	)

	for _, customer := range customerList {
		row := []string{customer.Name, customer.SaaSType, customer.CloudType, customer.Status, customer.Dataplane}
		table.Append(row)
	}

	table.Render()

	return nil
}

func getCustomerList() ([]customerList, error) {

	response, err := http.Get(common.GetBzUrl() + customersUrlPath)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if response.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", response.StatusCode, body)
	}
	if err != nil {
		return nil, err
	}
	var resp common.ServerResp

	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return nil, err
	}
	var customerList []customerList

	err = json.Unmarshal([]byte(resp.Msg), &customerList)
	if err != nil {
		return nil, err
	}

	return customerList, nil

}
