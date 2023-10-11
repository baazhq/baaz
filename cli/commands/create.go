package commands

import (
	"bz/pkg/customers"
	"bz/pkg/dataplane"
	"fmt"

	"github.com/spf13/cobra"
)

var createValidArgs = []string{
	"customer",
}

var file string

var (
	createCmd = &cobra.Command{
		Use:   "create",
		Short: "bz create - create entites [customers, dataplane, tenants, applications] in baaz control plane",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "customer":
				resp, err := customers.CreateCustomer(file)
				if err != nil {
					return err
				}
				fmt.Println(resp)
				return nil
			case "dataplane":
				resp, err := dataplane.CreateDataplane(file)
				if err != nil {
					return err
				}
				fmt.Println(resp)
				return nil
			default:
				return NotValidArgs(getValidArgs)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&file, "file", "f", "", ".yaml file speficifying entity to be created")
}
