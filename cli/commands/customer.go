package commands

import (
	"github.com/spf13/cobra"
)

var (
	cmd1 = &cobra.Command{
		Use:       "get",
		Short:     "bz - list all customers in baaz server",
		Long:      ``,
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"customerss"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(cmd1)
}
