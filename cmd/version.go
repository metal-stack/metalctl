package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows the version server",
	Long:  "the --version command returns the client version, but this command reaches out to the server's version endpoint.",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := driver.VersionGet()
		if err != nil {
			return err
		}

		return detailer.Detail(resp.Version)
	},
	PreRun: bindPFlags,
}
