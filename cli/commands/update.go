package commands

import (
	"bz/pkg/dataplanes"
	"bz/pkg/tenants"
	"bz/pkg/tenantsinfra"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "bz update - create entites [dataplane] in baaz control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "dataplane", "dataplanes":
				if dataplane_name == "" {
					return fmt.Errorf("dataplane name cannot be empty")
				}
				if file == "" {
					return fmt.Errorf("file path cannot be empty")
				}
				resp, err := dataplanes.UpdateDataplane(file, dataplane_name)
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "tenant", "tenants":
				if customer_name == "" {
					return fmt.Errorf("customer name can not be empty")
				}
				if tenant_name == "" {
					return fmt.Errorf("tenant name can not be empty")
				}
				if file == "" {
					return fmt.Errorf("file path cannot be empty")
				}
				resp, err := tenants.UpdateTenant(file, customer_name, tenant_name)
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "tenantinfra", "tenantsinfra":
				if dataplane_name == "" {
					return fmt.Errorf("dataplane name cannot be nil")
				}
				if tenantsinfra_name == "" {
					return fmt.Errorf("tenantinfra name cannot be nil")
				}
				if file == "" {
					return fmt.Errorf("file path cannot be nil, use --file or -f flag to specify the file path")
				}
				resp, err := tenantsinfra.UpdateTenantsInfra(file, dataplane_name, tenantsinfra_name)
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
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&file, "file", "f", "", ".yaml file speficifying entity to be created")
	updateCmd.Flags().StringVarP(&dataplane_name, "dataplane", "", "", "dataplane name")
	updateCmd.Flags().StringVarP(&customer_name, "customer", "", "", "customer name")
	updateCmd.Flags().StringVarP(&tenant_name, "tenant", "", "", "tenant name")
	updateCmd.Flags().StringVarP(&tenantsinfra_name, "tenantinfra", "", "", "tenane infra name")
}
