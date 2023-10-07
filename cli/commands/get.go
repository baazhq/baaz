package commands

import (
	"bz/pkg/customers"

	"github.com/spf13/cobra"
)

var getValidArgs = []string{
	"customers",
}

var (
	getCmd = &cobra.Command{
		Use:       "get",
		Short:     "bz get - list entites [Customers, Dataplane, Tenants, Applications] in baaz control plane",
		Args:      cobra.ExactArgs(1),
		ValidArgs: getValidArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				switch arg {
				case "customers":
					return customers.GetCustomers()
				default:
					return NotValidArgs(getValidArgs)
				}
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(getCmd)
}