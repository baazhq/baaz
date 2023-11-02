package commands

import (
	"bz/pkg/customers"
	"bz/pkg/dataplanes"

	"github.com/spf13/cobra"
)

var (
	getCmd = &cobra.Command{
		Use:       "get",
		Short:     "bz get - list entites [Customers, Dataplane, Tenants, Applications] in baaz control plane",
		Args:      cobra.ExactArgs(1),
		ValidArgs: commonValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "customers", "customer":
				return customers.GetCustomers()
			case "dataplanes":
				return dataplanes.GetDataplanes()
			default:
				return NotValidArgs(commonValidArgs)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(getCmd)
}
