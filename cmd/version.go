package cmd

import (
	"fmt"

	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
)

func newVersionCmd(c *config) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "print the client and server version information",
		Long:  "print the client and server version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := api.Version{
				Client: v.V.String(),
			}

			resp, err := c.driver.VersionGet()
			if err == nil {
				v.Server = resp.Version
			}
			if err2 := output.NewDetailer().Detail(v); err2 != nil {
				return err2
			}
			if err != nil {
				return fmt.Errorf("failed to get server info: %w", err)
			}
			return nil
		},
		PreRun: bindPFlags,
	}
	return versionCmd
}
