package commands

import (
	"github.com/spf13/cobra"
)

var commonValidArgs = []string{
	"customers",
	"customer",
	"dataplanes",
	"dataplane",
	"size",
	"sizes",
}

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
