package commands

import (
	"bz/pkg/customers"
	"bz/pkg/dataplanes"
	"bz/pkg/tenantsinfra"

	"fmt"

	"github.com/spf13/cobra"
)

var (
	file           string
	dataplane_name string
)

var (
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "bz create - create entites [customers, dataplane, tenants, applications] in baaz control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "customer", "customers":
				resp, err := customers.CreateCustomer(file)
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "dataplane", "dataplanes":
				resp, err := dataplanes.CreateDataplane(file)
				if err != nil {
					fmt.Println(err)
					return err
				}
				fmt.Println(resp)
			case "tenantinfra", "tenantsinfra":
				resp, err := tenantsinfra.CreateTenantsInfra(file, dataplane_name)
				if err != nil {
					return err
				}
				fmt.Println(resp)
			default:
				return NotValidArgs(commonValidArgs)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&file, "file", "f", "", ".yaml file speficifying entity to be created")
	createCmd.Flags().StringVarP(&dataplane_name, "dataplane_name", "", "", "dataplane name")

}
