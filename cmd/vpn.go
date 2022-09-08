package cmd

import (
	"fmt"
	"github.com/metal-stack/metal-go/api/client/vpn"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
)

type vpnOpts struct {
	ProjectID string
}

var vpnOptsInstance = &vpnOpts{}

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

	vpnKeyCmd.Flags().StringVar(&vpnOptsInstance.ProjectID, "project", "", "project ID for which auth key should be created")
	must(vpnKeyCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	vpnCmd.AddCommand(vpnKeyCmd)

	return vpnCmd
}

func (c *config) vpnAuthKeyCreate() error {
	if vpnOptsInstance.ProjectID == "" {
		return fmt.Errorf("Project ID should be specified")
	}

	resp, err := c.client.VPN().GetVPNAuthKey(
		vpn.NewGetVPNAuthKeyParams().WithBody(
			&models.V1VPNRequest{
				Pid: &vpnOptsInstance.ProjectID,
			}), nil,
	)
	if err != nil {
		return err
	}

	return c.describePrinter.Print(resp.Payload.AuthKey)
}
