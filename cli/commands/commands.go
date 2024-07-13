package commands

import (
	"github.com/spf13/cobra"
)

// Valid command arguments
var commonValidArgs = []string{
	"customers",
	"customer",
	"dataplanes",
	"dataplane",
	"tenants",
	"tenant",
	"tenantinfra",
	"tenantsinfra",
	"events",
	"event",
}

var (
	file                         string
	dataplane_name               string
	entity_name                  string
	duration                     string
	customer_name                string
	tenant_name                  string
	tenantsinfra_name            string
	private_mode                 bool
	kubernetes_config_server_url string
	namespace                    string
	aws_access_key               string
	aws_secret_key               string
)

var (
	rootCmd = &cobra.Command{
		Use:           "bz",
		Short:         "baaz cli - cli to interact with baaz server",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	return rootCmd.Execute()
}
