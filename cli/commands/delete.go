package commands

import (
	"bz/pkg/customers"
	"bz/pkg/dataplanes"
	"bz/pkg/tenants"
	"bz/pkg/tenantsinfra"
	"fmt"

	"github.com/spf13/cobra"
)

var deleteValidArgs = []string{
	"dataplane",
}

var (
	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "bz delete - delete entites [customers, dataplane, tenants, applications] in baaz control plane",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "customer":
				if customer_name == "" {
					return fmt.Errorf("customer name cannot be nil")
				}
				resp, err := customers.DeleteCustomer(customer_name)
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "dataplane":
				if dataplane_name == "" {
					return fmt.Errorf("dataplane name cannot be nil")
				}
				resp, err := dataplanes.DeleteDataplane(dataplane_name)
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "tenant":
				if tenant_name == "" {
					return fmt.Errorf("tenant name cannot be nil")
				}
				if customer_name == "" {
					return fmt.Errorf("customer name cannot be nil")
				}
				resp, err := tenants.DeleteTenant(customer_name, tenant_name)
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "tenantinfra":
				if dataplane_name == "" {
					return fmt.Errorf("dataplane name cannot be nil")
				}
				if tenantsinfra_name == "" {
					return fmt.Errorf("tenantinfra name cannot be nil")
				}
				resp, err := tenantsinfra.DeleteTenantsInfra(dataplane_name, tenantsinfra_name)
				if err != nil {
					return err
				}
				fmt.Println(resp)
			default:
				return NotValidArgs(deleteValidArgs)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringVar(&dataplane_name, "dataplane", "", "name of the dataplane to be deleted")
	deleteCmd.Flags().StringVar(&customer_name, "customer", "", "name of the customer to be deleted")
	deleteCmd.Flags().StringVar(&tenant_name, "tenant", "", "name of the tenant to be deleted")
	deleteCmd.Flags().StringVar(&tenantsinfra_name, "tenantinfra", "", "tenane infra name")
}
