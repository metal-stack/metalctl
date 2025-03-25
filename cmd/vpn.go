package cmd

import (
	"github.com/metal-stack/metal-go/api/client/vpn"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type vpnCmd struct {
	c *api.Config
}

func newVPNCmd(c *api.Config) *cobra.Command {
	w := &vpnCmd{
		c: c,
	}

	vpnCmd := &cobra.Command{
		Use:   "vpn",
		Short: "access VPN",
		Long:  "access VPN",
	}
	vpnKeyCmd := &cobra.Command{
		Use:   "key",
		Short: "create an auth key",
		Long:  "create an auth key to connect to VPN",
		Example: `auth key for tailscale can be created by this command:
metalctl vpn key \
	-- project cluster01
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.vpnAuthKeyCreate()
		},
	}

	vpnKeyCmd.Flags().String("project", "", "project ID for which auth key should be created")
	vpnKeyCmd.Flags().Bool("ephemeral", true, "create an ephemeral key")
	genericcli.Must(vpnKeyCmd.MarkFlagRequired("project"))
	genericcli.Must(vpnKeyCmd.RegisterFlagCompletionFunc("project", c.Comp.ProjectListCompletion))
	vpnCmd.AddCommand(vpnKeyCmd)

	return vpnCmd
}

func (c *vpnCmd) vpnAuthKeyCreate() error {

	resp, err := c.c.Client.VPN().GetVPNAuthKey(
		vpn.NewGetVPNAuthKeyParams().WithBody(
			&models.V1VPNRequest{
				Pid:       pointer.Pointer(viper.GetString("project")),
				Ephemeral: pointer.PointerOrNil(viper.GetBool("ephemeral")),
			}), nil,
	)
	if err != nil {
		return err
	}

	return c.c.DescribePrinter.Print(resp.Payload.AuthKey)
}
