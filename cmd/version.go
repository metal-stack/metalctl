package cmd

import (
	"fmt"

	metalmodels "github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
)

type Version struct {
	Client string                   `yaml:"client"`
	Server *metalmodels.RestVersion `yaml:"server,omitempty"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the client and server version information",
	Long:  "print the client and server version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		v := Version{
			Client: v.V.String(),
		}

		resp, err := driver.VersionGet()
		if err == nil {
			v.Server = resp.Version
		}

		if err2 := detailer.Detail(v); err2 != nil {
			return err2
		}
		if err != nil {
			return fmt.Errorf("failed to get server info: %w", err)
		}
		return nil
	},
	PreRun: bindPFlags,
}
