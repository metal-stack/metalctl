package cmd

import (
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
)

func newHealthCmd(c *config) *cobra.Command {

	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "shows the server health",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := c.driver.HealthGet()
			if err != nil {
				return err
			}

			return output.NewDetailer().Detail(resp.Health)
		},
		PreRun: bindPFlags,
	}

	return healthCmd
}
