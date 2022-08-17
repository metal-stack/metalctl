package cmd

import (
	"errors"

	"github.com/metal-stack/metal-go/api/client/health"
	"github.com/spf13/cobra"
)

func newHealthCmd(c *config) *cobra.Command {
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "shows the server health",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := c.client.Health().Health(health.NewHealthParams(), nil)
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

			return c.describePrinter.Print(resp.Payload)
		},
	}

	return healthCmd
}
