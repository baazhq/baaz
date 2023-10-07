package customers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/olekukonko/tablewriter"
)

type resp struct {
	Msg        string      `json:"Msg"`
	Status     string      `json:"Status"`
	StatusCode int         `json:"StatusCode"`
	Err        interface{} `json:"Err"`
}

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
	url := "http://localhost:8000/api/v1/customer"
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var resp resp

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
