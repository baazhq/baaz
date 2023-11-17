package commands

import (
	"bz/pkg/customers"
	"bz/pkg/dataplanes"
	"bz/pkg/tenantsize"

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
			case "tenant", "tenants":
				if args[1] == "size" || args[1] == "sizes" {
					return tenantsize.GetTenantSizes()
				}
			default:
				return NotValidArgs(commonValidArgs)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(getCmd)
}
