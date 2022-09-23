package cmd

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/metal-stack/metal-go/api/client/firewall"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/client/vpn"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
	"time"
)

const (
	tailscaleImage        = "tailscale/tailscale:v1.31"
	taiscaleStatusRetries = 50
)

type firewallCmd struct {
	*config
}

func newFirewallCmd(c *config) *cobra.Command {
	w := firewallCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1FirewallCreateRequest, any, *models.V1FirewallResponse]{
		BinaryName: binaryName,
		GenericCLI: genericcli.NewGenericCLI[*models.V1FirewallCreateRequest, any, *models.V1FirewallResponse](w).WithFS(c.fs),
		OnlyCmds: genericcli.OnlyCmds(
			genericcli.ListCmd,
			genericcli.DescribeCmd,
			genericcli.CreateCmd,
		),
		Singular:             "firewall",
		Plural:               "firewalls",
		Description:          "firewalls are used to establish network connectivity between metal-stack networks. firewalls are similar to machines but are managed by the provider. almost every command of the machine command subset works on firewalls, too.",
		Aliases:              []string{"fw"},
		CreateRequestFromCLI: w.createRequestFromCLI,
		Sorter:               sorters.FirewallSorter(),
		DescribePrinter:      func() printers.Printer { return c.describePrinter },
		ListPrinter:          func() printers.Printer { return c.listPrinter },
		ValidArgsFn:          c.comp.FirewallListCompletion,
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			c.addMachineCreateFlags(cmd, "firewall")
			cmd.Aliases = []string{"allocate"}
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "ID to filter [optional]")
			cmd.Flags().String("partition", "", "partition to filter [optional]")
			cmd.Flags().String("size", "", "size to filter [optional]")
			cmd.Flags().String("name", "", "allocation name to filter [optional]")
			cmd.Flags().String("project", "", "allocation project to filter [optional]")
			cmd.Flags().String("image", "", "allocation image to filter [optional]")
			cmd.Flags().String("hostname", "", "allocation hostname to filter [optional]")
			cmd.Flags().StringSlice("mac", []string{}, "mac to filter [optional]")
			cmd.Flags().StringSlice("tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")
			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
			must(cmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			must(cmd.RegisterFlagCompletionFunc("id", c.comp.FirewallListCompletion))
			must(cmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))
		},
	}

	firewallSSHCmd := &cobra.Command{
		Use:   "ssh <firewall ID>",
		Short: "SSH to a firewall",
		Long:  `SSH to a firewall via VPN.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.firewallSSH(args)
		},
		ValidArgsFunction: c.comp.FirewallListCompletion,
	}
	firewallSSHCmd.Flags().StringP("identity", "i", "~/.ssh/id_rsa", "specify identity file to SSH to the firewall like: -i path/to/id_rsa")

	return genericcli.NewCmds(cmdsConfig, firewallSSHCmd)
}

func (c firewallCmd) Get(id string) (*models.V1FirewallResponse, error) {
	resp, err := c.client.Firewall().FindFirewall(firewall.NewFindFirewallParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c firewallCmd) List() ([]*models.V1FirewallResponse, error) {
	resp, err := c.client.Firewall().FindFirewalls(firewall.NewFindFirewallsParams().WithBody(&models.V1FirewallFindRequest{
		ID:                 viper.GetString("id"),
		PartitionID:        viper.GetString("partition"),
		Sizeid:             viper.GetString("size"),
		Name:               viper.GetString("name"),
		AllocationProject:  viper.GetString("project"),
		AllocationImageID:  viper.GetString("image"),
		AllocationHostname: viper.GetString("hostname"),
		NicsMacAddresses:   viper.GetStringSlice("mac"),
		Tags:               viper.GetStringSlice("tags"),
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c firewallCmd) Delete(_ string) (*models.V1FirewallResponse, error) {
	return nil, fmt.Errorf("firewall entity does not support delete operation, use machine delete")
}

func (c firewallCmd) Create(rq *models.V1FirewallCreateRequest) (*models.V1FirewallResponse, error) {
	resp, err := c.client.Firewall().AllocateFirewall(firewall.NewAllocateFirewallParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c firewallCmd) Update(rq any) (*models.V1FirewallResponse, error) {
	return nil, fmt.Errorf("firewall entity does not support update operation, use machine update")
}

func (c firewallCmd) ToCreate(r *models.V1FirewallResponse) (*models.V1FirewallCreateRequest, error) {
	return firewallResponseToCreate(r), nil
}

func (c firewallCmd) ToUpdate(r *models.V1FirewallResponse) (any, error) {
	return nil, fmt.Errorf("firewall entity does not support update operation, use machine update")
}

func firewallResponseToCreate(r *models.V1FirewallResponse) *models.V1FirewallCreateRequest {
	var (
		ips        []string
		networks   []*models.V1MachineAllocationNetwork
		allocation = pointer.SafeDeref(r.Allocation)
	)
	for _, s := range allocation.Networks {
		ips = append(ips, s.Ips...)
		networks = append(networks, &models.V1MachineAllocationNetwork{
			Autoacquire: pointer.Pointer(len(s.Ips) == 0),
			Networkid:   s.Networkid,
		})
	}

	return &models.V1FirewallCreateRequest{
		Description:        allocation.Description,
		Filesystemlayoutid: pointer.SafeDeref(pointer.SafeDeref(allocation.Filesystemlayout).ID),
		Hostname:           pointer.SafeDeref(allocation.Hostname),
		Imageid:            pointer.SafeDeref(allocation.Image).ID,
		Ips:                ips,
		Name:               r.Name,
		Networks:           networks,
		Partitionid:        r.Partition.ID,
		Projectid:          allocation.Project,
		Sizeid:             r.Size.ID,
		SSHPubKeys:         allocation.SSHPubKeys,
		Tags:               r.Tags,
		UserData:           base64.StdEncoding.EncodeToString([]byte(allocation.UserData)),
		UUID:               pointer.SafeDeref(r.ID),
	}
}

func (c *firewallCmd) createRequestFromCLI() (*models.V1FirewallCreateRequest, error) {
	mcr, err := machineCreateRequest()
	if err != nil {
		return nil, fmt.Errorf("firewall create error:%w", err)
	}

	return &models.V1FirewallCreateRequest{
		Description:        mcr.Description,
		Filesystemlayoutid: mcr.Filesystemlayoutid,
		Partitionid:        mcr.Partitionid,
		Hostname:           mcr.Hostname,
		Imageid:            mcr.Imageid,
		Name:               mcr.Name,
		UUID:               mcr.UUID,
		Projectid:          mcr.Projectid,
		Sizeid:             mcr.Sizeid,
		SSHPubKeys:         mcr.SSHPubKeys,
		UserData:           mcr.UserData,
		Tags:               mcr.Tags,
		Networks:           mcr.Networks,
		Ips:                mcr.Ips,
	}, nil
}

func (c *config) firewallSSH(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("machine ID is expected as an argument")
	}
	firewallID := args[0]
	resp, err := c.client.Firewall().FindFirewall(firewall.NewFindFirewallParams().WithID(firewallID), nil)
	if err != nil {
		return fmt.Errorf("failed to find firewall: %w", err)
	}

	if resp.Payload.Vpn != nil && resp.Payload.Vpn.Connected != nil && *resp.Payload.Vpn.Connected {
		return c.firewallSSHViaVPN(firewallID)
	}

	// Try to connect to firewall via SSH
	if err := c.firewallPureSSH(resp.Payload.Allocation); err != nil {
		return fmt.Errorf("failed to connect to firewall via SSH: %w", err)
	}

	return nil
}

func (c *config) firewallPureSSH(fwAllocation *models.V1MachineAllocation) (err error) {
	networks := fwAllocation.Networks
	for _, nw := range networks {
		if *nw.Underlay || *nw.Private {
			continue
		}
		for _, ip := range nw.Ips {
			if portOpen(ip, "22", time.Second) {
				err = SSHClient("metal", viper.GetString("identity"), ip, 22)
				if err != nil {
					return fmt.Errorf("machine console error:%w", err)
				}

				return nil
			}
		}
	}

	return fmt.Errorf("no ip with a open ssh port found")
}

func (c *config) firewallSSHViaVPN(firewallID string) (err error) {
	// Get firewall's project
	machineGetResp, err := c.client.Machine().FindMachine(machine.NewFindMachineParams().WithID(firewallID), nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve firewall details: %w", err)
	}
	projectID := machineGetResp.Payload.Allocation.Project

	authKeyResp, err := c.client.VPN().GetVPNAuthKey(vpn.NewGetVPNAuthKeyParams().WithBody(&models.V1VPNRequest{
		Pid:       projectID,
		Ephemeral: pointer.Pointer(true),
	}), nil)
	if err != nil {
		return fmt.Errorf("failed to get VPN auth key: %w", err)
	}

	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to initialize Docker client: %w", err)
	}

	// Deploy tailscaled
	ctx := context.Background()
	if err := pullImageIfNotExists(ctx, cli, tailscaleImage); err != nil {
		return fmt.Errorf("failed to pull tailscale image: %w", err)
	}

	containerConfig := &container.Config{
		Image: tailscaleImage,
		Cmd:   []string{"tailscaled", "--tun=userspace-networking", "--socks5-server=:1055"},
	}
	hostConfig := &container.HostConfig{
		NetworkMode: container.NetworkMode("host"),
	}
	containerName := "tailscaled"
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return err
	}

	tailscaledContainerID := resp.ID
	defer func() {
		if e := cli.ContainerRemove(ctx, tailscaledContainerID, types.ContainerRemoveOptions{}); e != nil {
			if err != nil {
				e = fmt.Errorf("%s: %w", e, err)
			}
			err = e
		}
	}()

	if err = cli.ContainerStart(ctx, tailscaledContainerID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	defer func() {
		if e := cli.ContainerStop(ctx, tailscaledContainerID, nil); e != nil {
			if err != nil {
				e = fmt.Errorf("%s: %w", e, err)
			}
			err = e
		}
	}()

	// Exec tailscale up
	execConfig := types.ExecConfig{
		Cmd: []string{"tailscale", "up", "--auth-key=" + *authKeyResp.Payload.AuthKey, "--login-server=" + *authKeyResp.Payload.Address},
	}
	execResp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create tailscaled exec: %w", err)
	}
	if err := cli.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{}); err != nil {
		return fmt.Errorf("failed to start tailscaled exec: %w", err)
	}

	// Connect to the firewall via SSH
	firewallVPNAddr, err := c.getFirewallVPNAddr(ctx, cli, containerName, *machineGetResp.Payload.ID)
	if err != nil {
		return fmt.Errorf("failed to get Firewall VPN address: %w", err)
	}

	err = SSHClientOverSOCKS5("metal", viper.GetString("identity"), firewallVPNAddr, 22, ":1055")
	if err != nil {
		return fmt.Errorf("machine console error:%w", err)
	}

	return nil
}

// TailscaleStatus and TailscalePeerStatus structs are used to parse VPN IP for the machine
type TailscaleStatus struct {
	Peer map[string]*TailscalePeerStatus
}

type TailscalePeerStatus struct {
	HostName     string
	TailscaleIPs []string
}

func pullImageIfNotExists(ctx context.Context, cli *dockerclient.Client, tag string) error {
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	for _, i := range images {
		for _, t := range i.RepoTags {
			if t == tag {
				return nil
			}
		}
	}

	reader, err := cli.ImagePull(ctx, tag, types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	return nil
}

func (c *config) getFirewallVPNAddr(ctx context.Context, cli *dockerclient.Client, containerName, fwName string) (addr string, err error) {
	// Wait until Peers info is filled
	for i := 0; i < taiscaleStatusRetries; i++ {
		execConfig := types.ExecConfig{
			Cmd:          []string{"tailscale", "status", "--json"},
			AttachStdout: true,
		}
		execResp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
		if err != nil {
			return "", fmt.Errorf("failed to create tailscale status exec: %w", err)
		}
		resp, err := cli.ContainerExecAttach(ctx, execResp.ID, types.ExecStartCheck{})
		if err != nil {
			return "", fmt.Errorf("failed to attach to tailscale status exec: %w", err)
		}

		var data string
		s := bufio.NewScanner(resp.Reader)
		for s.Scan() {
			data += s.Text()
		}

		// Skipping noise at the beginning
		var i int
		for _, c := range data {
			if c == '{' {
				break
			}
			i++
		}
		ts := &TailscaleStatus{}
		if err := json.Unmarshal([]byte(data[i:]), ts); err != nil {
			continue
		}

		if ts.Peer != nil {
			for _, p := range ts.Peer {
				if strings.HasPrefix(p.HostName, fwName) {
					return p.TailscaleIPs[0], nil
				}
			}
		}
	}

	return "", fmt.Errorf("failed to find IP for specified firewall")
}
