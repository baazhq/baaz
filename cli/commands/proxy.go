package commands

import (
	"bz/pkg/proxy"

	"github.com/spf13/cobra"
)

var (
	proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "bz proxy - port-forward baaz control plane locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			proxy.Run()
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(proxyCmd)
}
