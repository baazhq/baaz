package commands

import (
	"bz/pkg/helm"
	"bz/pkg/kubeconfig"
	"bz/pkg/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "bz init - initalise baaz control plane, this command deploys the control plane",
		RunE: func(cmd *cobra.Command, args []string) error {

			if private_mode {

				if customer_name == "" {
					return fmt.Errorf("Customer Name cannot be nil")
				}

				config, err := kubeconfig.GetCustomerKubeConfig(customer_name)
				if err != nil {
					return err
				}

				err = kubeconfig.WriteKubeConfig2Cm(customer_name, config, utils.GetLocalKubeClientset())
				if err != nil {
					return err
				}

				helmBuild := helm.NewHelm(
					"baaz",
					customer_name,
					"../chart/baaz/",
					nil,
					[]string{
						"private_mode.enabled=true",
						"private_mode.customer_name=" + customer_name,
						"private_mode.args.kubeconfig=/kubeconfig/" + customer_name + "-kubeconfig",
						"private_mode.args.private_mode=true",
						"private_mode.customer_name=" + customer_name,
					},
				)

				err = helmBuild.Apply()
				if err != nil {
					return err
				}

			} else {

				if kubernetes_config_server_url == "" {
					return fmt.Errorf("kubernetes_config_server_url flag cannot be nil")
				}

				if namespace == "" {
					namespace = "baaz"
				}

				helmBuild := helm.NewHelm(
					"baaz",
					namespace,
					"../chart/baaz/",
					[]string{"env.KUBERNETES_CONFIG_SERVER_URL=" + kubernetes_config_server_url},
					nil,
				)

				err := helmBuild.Apply()
				if err != nil {
					return err
				}

			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&customer_name, "customer", "", "", "Customer Name")
	initCmd.Flags().BoolVarP(&private_mode, "private_mode", "", false, "Run BaaZ control plane in private mode")
	initCmd.Flags().StringVarP(&kubernetes_config_server_url, "kubernetes_config_server_url", "", "", "Kubernetes config server url, make sure it is public accessible")
	initCmd.Flags().StringVarP(&namespace, "namespace", "", "", "Namespace to deploy BaaZ control plane")

}
