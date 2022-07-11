package cmd

import (
	"errors"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/health"
	"github.com/spf13/cobra"
)

func newHealthCmd(c *config) *cobra.Command {

	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "shows the server health",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := c.driver.HealthGet()
			if err != nil {
				var r *health.HealthInternalServerError
				if errors.As(err, &r) {
					resp = &metalgo.HealthGetResponse{
						Health: r.Payload,
					}
				} else {
					return err
				}
			}

			return defaultToYAMLPrinter().Print(resp.Health)
		},
		PreRun: bindPFlags,
	}

	return healthCmd
}
