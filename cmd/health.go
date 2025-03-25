package cmd

import (
	"errors"

	"github.com/metal-stack/metal-go/api/client/health"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
)

func newHealthCmd(c *api.Config) *cobra.Command {
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "shows the server health",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := c.Client.Health().Health(health.NewHealthParams(), nil)
			if err != nil {
				var r *health.HealthInternalServerError
				if errors.As(err, &r) {
					resp = &health.HealthOK{
						Payload: r.Payload,
					}
				} else {
					return err
				}
			}

			return c.DescribePrinter.Print(resp.Payload)
		},
	}

	return healthCmd
}
