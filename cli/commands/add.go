package commands

import (
	"bz/pkg/dataplanes"
	"fmt"

	"github.com/spf13/cobra"
)

var name string

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "bz add - add a customer to existing dataplane in baaz control plane",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			switch args[0] {
			case "dataplane", "dataplanes":
				resp, err := dataplanes.AddDataplane(args[1], args[2])
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
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&name, "name", "n", "", "name of entity")
}
