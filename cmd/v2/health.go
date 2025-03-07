package v2

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/metal-stack/api/go/client"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/spf13/cobra"
)

func NewHealthCmd(c client.Client) *cobra.Command {
	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "shows the server health",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := c.Apiv2().Health().Get(context.Background(), connect.NewRequest(&apiv2.HealthServiceGetRequest{}))
			if err != nil {
				return err
			}

			fmt.Printf("health:%#v", resp.Msg.Health)
			return nil
		},
	}

	return healthCmd
}
