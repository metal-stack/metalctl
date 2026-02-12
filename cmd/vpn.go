package cmd

import (
	"github.com/metal-stack/metal-go/api/client/vpn"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newVPNCmd(c *config) *cobra.Command {
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
			return c.vpnAuthKeyCreate()
		},
	}

	vpnKeyCmd.Flags().String("project", "", "project ID for which auth key should be created")
	vpnKeyCmd.Flags().Bool("ephemeral", true, "create an ephemeral key")
	vpnKeyCmd.Flags().StringP("reason", "", "", "a short description why access to the vpn is required")
	genericcli.Must(vpnKeyCmd.MarkFlagRequired("project"))
	genericcli.Must(vpnKeyCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	vpnCmd.AddCommand(vpnKeyCmd)

	return vpnCmd
}

func (c *config) vpnAuthKeyCreate() error {

	resp, err := c.client.VPN().GetVPNAuthKey(
		vpn.NewGetVPNAuthKeyParams().WithBody(
			&models.V1VPNRequest{
				Pid:       new(viper.GetString("project")),
				Ephemeral: pointer.PointerOrNil(viper.GetBool("ephemeral")),
				Reason:    new(viper.GetString("reason")),
			}), nil,
	)
	if err != nil {
		return err
	}

	return c.describePrinter.Print(resp.Payload.AuthKey)
}
