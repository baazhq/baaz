package commands

import (
	"github.com/spf13/cobra"
)

var (
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "bz get - list entites [Customers, Dataplane, Tenants, Applications] in baaz control plane",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// read flag
			// for _, arg := range args {

			// 	switch arg {
			// 	// case "a":
			// 	// 	return customers.GetCustomers()
			// 	// default:
			// 	// 	return NotValidArgs(getValidArgs)
			// 	// }
			// }
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(createCmd)
	var file string
	createCmd.Flags().StringVarP(&file, "file", "f", "", ".yaml file speficifying entity to be created")

}
