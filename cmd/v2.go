package cmd

import (
	"github.com/metal-stack/api/go/client"
	v2 "github.com/metal-stack/metalctl/cmd/v2"

	"github.com/spf13/cobra"
)

func newV2Cmd(c client.Client) *cobra.Command {
	v2Cmd := &cobra.Command{
		Use:   "v2",
		Short: "v2 commands to talk to the grpc apiv2 of metal-stack.io",
	}

	v2Cmd.AddCommand(v2.NewVersionCmd(c))
	v2Cmd.AddCommand(v2.NewHealthCmd(c))
	v2Cmd.AddCommand(v2.NewLoginCmd(c))

	return v2Cmd
}
