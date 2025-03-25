package v2

import (
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
)

func AddCmds(cmd *cobra.Command, c *api.Config) {
	v2Cmd := &cobra.Command{
		Use:          "v2",
		Short:        "v2 commands",
		Long:         "",
		SilenceUsage: true,
		Hidden:       true,
	}

	v2Cmd.AddCommand(newLoginCmd(c))
	v2Cmd.AddCommand(newLogoutCmd(c))
	v2Cmd.AddCommand(newImageCmd(c))

	cmd.AddCommand(v2Cmd)
}
