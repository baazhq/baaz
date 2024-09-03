package commands

import (
	"bz/pkg/dataplanes"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "bz add - add a customer to existing dataplane in baaz control plane",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			switch args[0] {
			case "dataplane", "dataplanes":
				if dataplane_name == "" {
					return fmt.Errorf("dataplane name can't be empty, use flag --dataplane={dataplane_name}")
				}
				if customer_name == "" {
					return fmt.Errorf("customer name can't be empty, use flag --customer={customer_name}")
				}
				resp, err := dataplanes.AddDataplane(dataplane_name, customer_name)
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
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&dataplane_name, "dataplane", "", "", "dataplane name")
	addCmd.Flags().StringVarP(&customer_name, "customer", "", "", "customer name")
}
