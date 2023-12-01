package commands

import (
	"bz/pkg/customers"
	"bz/pkg/dataplanes"
	"bz/pkg/tenantsinfra"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getCmd = &cobra.Command{
		Use:       "get",
		Short:     "bz get - list entites [Customers, Dataplane, Tenants, Applications] in baaz control plane",
		ValidArgs: commonValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "customers", "customer":
				return customers.GetCustomers()
			case "dataplanes":
				return dataplanes.GetDataplanes()
			case "tenantinfra", "tenantsinfra":
				if dataplane_name == "" {
					return fmt.Errorf("Dataplane named cannot be nil")
				}
				return tenantsinfra.GetTenantsInfra(dataplane_name)
			default:
				return NotValidArgs(commonValidArgs)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&dataplane_name, "dataplane", "", "", "dataplane name")
}
