package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/metal-stack/metal-go/api/client/vpn"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	metalssh "github.com/metal-stack/metal-lib/pkg/ssh"
	metalvpn "github.com/metal-stack/metal-lib/pkg/vpn"
	"github.com/spf13/viper"
)

func (c *firewallCmd) firewallSSHViaVPN(firewall *models.V1FirewallResponse) (err error) {
	if firewall.Allocation == nil || firewall.Allocation.Project == nil {
		return fmt.Errorf("firewall allocation or allocation.project is nil")
	}
	projectID := firewall.Allocation.Project
	fmt.Fprintf(c.out, "accessing firewall through vpn ")
	authKeyResp, err := c.client.VPN().GetVPNAuthKey(vpn.NewGetVPNAuthKeyParams().WithBody(&models.V1VPNRequest{
		Pid:       projectID,
		Ephemeral: pointer.Pointer(true),
	}), nil)
	if err != nil {
		return fmt.Errorf("failed to get VPN auth key: %w", err)
	}
	ctx := context.Background()
	v, err := metalvpn.Connect(ctx, *firewall.ID, *authKeyResp.Payload.Address, *authKeyResp.Payload.AuthKey)
	if err != nil {
		return err
	}
	defer v.Close()

	privateKeyFile := viper.GetString("identity")
	if strings.HasPrefix(privateKeyFile, "~/") {
		home, _ := os.UserHomeDir()
		privateKeyFile = filepath.Join(home, privateKeyFile[2:])
	}

	privateKey, err := os.ReadFile(privateKeyFile)
	if err != nil {
		return err
	}
	s, err := metalssh.NewClientWithConnection("metal", v.TargetIP, privateKey, v.Conn)
	if err != nil {
		return err
	}
	return s.Connect(nil)
}

// sshClient opens an interactive ssh session to the host on port with user, authenticated by the key.
func sshClient(user, keyfile, host string, port int, idToken *string, passwordAuth bool) error {
	privateKey, err := os.ReadFile(keyfile)
	if err != nil {
		return err
	}

	var opts []metalssh.ConnectOpt
	if passwordAuth {
		opts = append(opts, metalssh.ConnectOptOutputPassword(*idToken))
	}

	s, err := metalssh.NewClient(user, host, privateKey, port, opts...)
	if err != nil {
		return err
	}
	var env *metalssh.Env
	if idToken != nil {
		env = &metalssh.Env{"LC_METAL_STACK_OIDC_TOKEN": *idToken}
	}
	return s.Connect(env)
}
