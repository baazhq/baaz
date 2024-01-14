package commands

import (
	"bz/pkg/kubeconfig"
	"bz/pkg/utils"

	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "bz init - initalise baaz control plane, this command deploys the control plane",
		RunE: func(cmd *cobra.Command, args []string) error {

			config, err := kubeconfig.GetCustomerKubeConfig(customer_name)
			if err != nil {
				return err
			}

			err = kubeconfig.WriteKubeConfig2Cm(customer_name, config, utils.GetKubeClientset())
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&customer_name, "customer", "", "", "customer name")
}
