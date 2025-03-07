package v2

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/metal-stack/api/go/client"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/spf13/cobra"
)

func NewVersionCmd(c client.Client) *cobra.Command {
	healthCmd := &cobra.Command{
		Use:   "version",
		Short: "shows the server version",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := c.Apiv2().Version().Get(context.Background(), connect.NewRequest(&apiv2.VersionServiceGetRequest{}))
			if err != nil {
				return err
			}

			fmt.Printf("Version:%#v", resp.Msg)
			return nil
		},
	}

	return healthCmd
}
