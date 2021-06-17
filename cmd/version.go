package cmd

import (
	metalmodels "github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
)

type Version struct {
	Client string                   `json:"client"`
	Server *metalmodels.RestVersion `json:"server"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the client and server version information",
	Long:  "print the client and server version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := driver.VersionGet()
		if err != nil {
			return err
		}

		v := Version{
			Client: v.V.String(),
			Server: resp.Version,
		}
		return detailer.Detail(v)
	},
	PreRun: bindPFlags,
}
