package cmd

import (
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "shows the server health",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := driver.HealthGet()
		if err != nil {
			return err
		}

		return detailer.Detail(resp.Health)
	},
	PreRun: bindPFlags,
}
