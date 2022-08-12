package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/metal-stack/metal-go/api/client/machine"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/metal-stack/metal-go/api/client/firewall"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// for test
// args = [firewall id, identity, control plane address, auth key]
func (c *config) firewallSSH(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("Machine ID is expected as an argument")
	}
	firewallID := args[0]

	// Get firewall's project
	machineGetResp, err := c.client.Machine().FindMachine(machine.NewFindMachineParams().WithID(firewallID), nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve firewall details: %w", err)
	}
	projectId := *machineGetResp.Payload.Allocation.Project
	fmt.Println(projectId)

	// TODO: get VPN info

	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to initialize Docker client: %w", err)
	}
	ctx := context.Background()

	// Deploy tailscaled
	// docker run -d --name=tailscaled -v /var/lib:/var/lib -v /dev/net/tun:/dev/net/tun
	// --network=host --privileged tailscale/tailscale tailscaled
	containerConfig := &container.Config{
		Image: "tailscale/tailscale",
		Cmd:   []string{"tailscaled"},
	}
	hostConfig := &container.HostConfig{
		Binds: []string{
			"/var/lib:/var/lib",
			"/dev/net/tun:/dev/net/tun",
		},
		NetworkMode: container.NetworkMode("host"),
		Privileged:  true,
	}
	containerName := "tailscaled"
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return c.describePrinter.Print(err)
	}

	tailscaledContainerID := resp.ID
	defer func() {
		if e := cli.ContainerRemove(ctx, tailscaledContainerID, types.ContainerRemoveOptions{}); e != nil {
			if err != nil {
				e = fmt.Errorf("%s: %w", e, err)
			}
			err = c.describePrinter.Print(e)
		}
	}()

	if err = cli.ContainerStart(ctx, tailscaledContainerID, types.ContainerStartOptions{}); err != nil {
		return c.describePrinter.Print(err)
	}
	defer func() {
		if e := cli.ContainerStop(ctx, tailscaledContainerID, nil); e != nil {
			if err != nil {
				e = fmt.Errorf("%s: %w", e, err)
			}
			err = c.describePrinter.Print(e)
		}
	}()

	// Exec tailscale up
	execConfig := types.ExecConfig{
		Cmd: []string{"tailscale", "up", "--auth-key=", "--login-server="},
	}
	execResp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return c.describePrinter.Print(
			fmt.Errorf("failed to create tailscaled exec: %w", err),
		)
	}
	if err := cli.ContainerExecStart(ctx, execResp.ID, types.ExecStartCheck{}); err != nil {
		return c.describePrinter.Print(
			fmt.Errorf("failed to start tailscaled exec: %w", err),
		)
	}

	// ssh

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	return nil
}
