package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"bz/pkg/customers"
	"bz/pkg/dataplanes"
	"bz/pkg/events"
	"bz/pkg/tenants"
	"bz/pkg/tenantsinfra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:       "get",
	Short:     "bz get - list entities [Customers, Dataplane, Tenants, Applications] in baaz control plane",
	ValidArgs: commonValidArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle different entity types
		switch args[0] {
		case "customers", "customer":
			return customers.GetCustomers()
		case "dataplanes", "dataplane":
			return dataplanes.GetDataplanes()
		case "tenantinfra", "tenantsinfra":
			// Ensure dataplane name is provided
			if dataplane_name == "" {
				return fmt.Errorf("dataplane name cannot be nil")
			}
			return tenantsinfra.GetTenantsInfra(dataplane_name)
		case "events", "event":
			// Ensure dataplane name is provided
			if entity_name == "" {
				return fmt.Errorf("entity name cannot be nil")
			}
			if duration == "" {
				return fmt.Errorf("duration name cannot be nil")
			}
			return events.GetEvents(entity_name, duration)
		case "tenants", "tenant":
			// Ensure customer name is provided
			if customer_name == "" {
				return fmt.Errorf("customer cannot be nil")
			}
			return tenants.GetTenants(customer_name)
		default:
			// Handle invalid arguments
			return NotValidArgs(commonValidArgs)
		}
	},
}

func init() {
	// Add get command to root command
	rootCmd.AddCommand(getCmd)

	// Define flags for get command
	getCmd.Flags().StringVarP(&customer_name, "customer", "", "", "customer name")
	getCmd.Flags().StringVarP(&tenant_name, "tenant", "", "", "tenant name")
	getCmd.Flags().StringVarP(&dataplane_name, "dataplane", "", "", "dataplane name")
	getCmd.Flags().StringVarP(&entity_name, "entity", "", "", "entity name")
	getCmd.Flags().StringVarP(&duration, "duration", "", "", "duration to get events")

}
