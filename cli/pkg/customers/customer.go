package customers

import (
	"bytes"
	"bz/pkg/common"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/viper"
	log "k8s.io/klog/v2"

	"github.com/olekukonko/tablewriter"
)

var customersUrlPath = "/api/v1/customer"

func makeGetCustomerPath() string { return common.GetBzUrl() + customersUrlPath }
func makePostCustomerPath(customerName string) string {
	return common.GetBzUrl() + customersUrlPath + "/" + customerName
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
		"Customer_Name",
		"SaaS_Type",
		"Cloud_Type",
		"Status",
		"Dataplane",
	},
	)

	for _, customer := range customerList {
		row := []string{
			customer.Name,
			customer.SaaSType,
			customer.CloudType,
			customer.Status,
			customer.Dataplane,
		}
		table.Append(row)
	}

	table.Render()

	return nil
}

func getCustomerList() ([]customerList, error) {

	response, err := http.Get(makeGetCustomerPath())
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

func CreateCustomer(filePath string) (string, error) {

	viper.SetConfigFile(filePath)
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		return "", err
	}

	type createCustomer struct {
		Name      string            `json:"name"`
		SaaSType  string            `json:"saas_type"`
		CloudType string            `json:"cloud_type"`
		Labels    map[string]string `json:"lables"`
	}

	newCreateCustomer := createCustomer{
		SaaSType:  viper.GetString("customer.saas_type"),
		CloudType: viper.GetString("customer.cloud_type"),
		Labels:    viper.GetStringMapString("customer.labels"),
	}

	ccByte, err := json.Marshal(newCreateCustomer)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		makePostCustomerPath(viper.GetString("customer.name")),
		"application/json",
		bytes.NewBuffer(ccByte),
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
		return "Customer Created Successfully", nil
	}

	return "", nil
}
