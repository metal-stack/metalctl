package cmd

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/client/version"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
)

func newVersionCmd(c *api.Config) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "print the client and server version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := api.Version{
				Client: v.V.String(),
			}

			resp, err := c.Client.Version().Info(version.NewInfoParams(), nil)
			if err == nil {
				v.Server = resp.Payload
			}
			if err2 := c.DescribePrinter.Print(v); err2 != nil {
				return err2
			}
			if err != nil {
				return fmt.Errorf("failed to get server info: %w", err)
			}
			return nil
		},
	}

	return versionCmd
}
