package commands

import (
	"bz/pkg/customers"
	"bz/pkg/dataplanes"

	"github.com/spf13/cobra"
)

var getValidArgs = []string{
	"customers",
	"dataplanes",
}

var (
	getCmd = &cobra.Command{
		Use:       "get",
		Short:     "bz get - list entites [Customers, Dataplane, Tenants, Applications] in baaz control plane",
		Args:      cobra.ExactArgs(1),
		ValidArgs: getValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "customers":
				return customers.GetCustomers()
			case "dataplanes":
				return dataplanes.GetDataplanes()
			default:
				return NotValidArgs(getValidArgs)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(getCmd)
}
