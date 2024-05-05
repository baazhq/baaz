package commands

import (
	"bz/pkg/dataplanes"
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
				resp, err := dataplanes.UpdateDataplane(file, dataplane_name)
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
}
