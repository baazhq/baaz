package commands

import (
	"bz/pkg/kubeconfig"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "bz init - initalise baaz control plane, this command deploys the control plane",
		RunE: func(cmd *cobra.Command, args []string) error {

			fmt.Println(args)
			switch args[0] {
			case "init":
				config, err := kubeconfig.GetCustomerKubeConfig(customer_name)
				if err != nil {
					return err
				}
				bytes, err := json.Marshal(config)
				if err != nil {
					return err
				}
				fmt.Println(string(bytes))
			default:
				return NotValidArgs(commonValidArgs)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&customer_name, "customer", "", "", "customer name")
}
