package cmd

import (
	"fmt"

	"github.com/metal-stack/metalctl/cmd/output"

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
		PreRun: bindPFlags,
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

	resp, err := c.driver.GetVPNAuthKey(vpnOptsInstance.ProjectID)
	if err != nil {
		return err
	}

	return output.New().Print(resp.VPN.AuthKey)
}
