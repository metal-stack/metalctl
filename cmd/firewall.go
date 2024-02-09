package cmd

import (
	"encoding/base64"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"

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
			cmd.Flags().StringSlice("egress", nil, "egress firewall rule to deploy on creation format: tcp|udp@cidr#cidr@port#port@comment")
			cmd.Flags().StringSlice("ingress", nil, "ingress firewall rule to deploy on creation format: tcp|udp@cidr#cidr@port#port@comment")
			must(cmd.RegisterFlagCompletionFunc("egress", c.comp.FirewallEgressCompletion))
			must(cmd.RegisterFlagCompletionFunc("ingress", c.comp.FirewallIngressCompletion))
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "ID to filter [optional]")
			cmd.Flags().String("partition", "", "partition to filter [optional]")
			cmd.Flags().String("size", "", "size to filter [optional]")
			cmd.Flags().String("name", "", "allocation name to filter [optional]")
			cmd.Flags().String("project", "", "allocation project to filter [optional]")
			cmd.Flags().String("image", "", "allocation image to filter [optional]")
			cmd.Flags().String("hostname", "", "allocation hostname to filter [optional]")
			cmd.Flags().String("mac", "", "mac to filter [optional]")
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
			return w.firewallSSH(args)
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
	var macs []string
	if viper.IsSet("mac") {
		macs = pointer.WrapInSlice(viper.GetString("mac"))
	}

	resp, err := c.client.Firewall().FindFirewalls(firewall.NewFindFirewallsParams().WithBody(&models.V1FirewallFindRequest{
		ID:                 viper.GetString("id"),
		PartitionID:        viper.GetString("partition"),
		Sizeid:             viper.GetString("size"),
		Name:               viper.GetString("name"),
		AllocationProject:  viper.GetString("project"),
		AllocationImageID:  viper.GetString("image"),
		AllocationHostname: viper.GetString("hostname"),
		NicsMacAddresses:   macs,
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

func (c firewallCmd) Convert(r *models.V1FirewallResponse) (string, *models.V1FirewallCreateRequest, any, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.ID, firewallResponseToCreate(r), nil, nil
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

	egress, err := parseEgressFlags(viper.GetStringSlice("egress"))
	if err != nil {
		return nil, fmt.Errorf("firewall create error:%w", err)
	}
	ingress, err := parseIngressFlags(viper.GetStringSlice("ingress"))
	if err != nil {
		return nil, fmt.Errorf("firewall create error:%w", err)
	}

	var firewallRules *models.V1FirewallRules
	if len(egress) > 0 || len(ingress) > 0 {
		firewallRules = &models.V1FirewallRules{
			Egress:  egress,
			Ingress: ingress,
		}
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
		FirewallRules:      firewallRules,
	}, nil
}

// parseEgressFlags input must be in the form of
// proto@cidr#cidr@port#port#port@comment
// tcp@1.2.3.4/24#2.3.4.1/32@80#443#8080#8443@"Allow apt update"
func parseEgressFlags(inputs []string) ([]*models.V1FirewallEgressRule, error) {
	var rules []*models.V1FirewallEgressRule

	for _, input := range inputs {
		r, err := parseRuleSpec(input)
		if err != nil {
			return nil, err
		}

		rule := &models.V1FirewallEgressRule{
			Protocol: r.protocol,
			ToCidrs:  r.cidrs,
			Ports:    r.ports,
			Comment:  r.comment,
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func parseIngressFlags(inputs []string) ([]*models.V1FirewallIngressRule, error) {
	var rules []*models.V1FirewallIngressRule

	for _, input := range inputs {
		r, err := parseRuleSpec(input)
		if err != nil {
			return nil, err
		}

		rule := &models.V1FirewallIngressRule{
			Protocol:  r.protocol,
			FromCidrs: r.cidrs,
			Ports:     r.ports,
			Comment:   r.comment,
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

type rule struct {
	comment  string
	protocol string
	cidrs    []string
	ports    []int32
}

func parseRuleSpec(spec string) (*rule, error) {
	parts := strings.Split(spec, "@")
	if len(parts) < 3 {
		return nil, fmt.Errorf("at least proto, cidrs and ports must be given, spec:%q parts:%q", spec, parts)
	}
	if len(parts) > 4 {
		return nil, fmt.Errorf("malformed rule spec:%q", spec)
	}

	r := &rule{}
	comment := ""
	if len(parts) == 4 {
		comment = parts[3]
	}
	r.comment = comment
	r.protocol = parts[0]

	cidrs := strings.Split(parts[1], "#")
	ports := strings.Split(parts[2], "#")

	for _, cidr := range cidrs {
		p, err := netip.ParsePrefix(cidr)
		if err != nil {
			return nil, err
		}
		r.cidrs = append(r.cidrs, p.String())
	}

	for _, port := range ports {
		p, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			return nil, err
		}
		r.ports = append(r.ports, int32(p))
	}
	return r, nil
}

func (c *firewallCmd) firewallSSH(args []string) (err error) {
	firewallID, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	firewall, err := c.Get(firewallID)
	if err != nil {
		return fmt.Errorf("failed to find firewall: %w", err)
	}

	if firewall.Allocation != nil && firewall.Allocation.Vpn != nil && firewall.Allocation.Vpn.Connected != nil && *firewall.Allocation.Vpn.Connected {
		return c.firewallSSHViaVPN(firewall)
	}

	// Try to connect to firewall via SSH
	if err := c.firewallPureSSH(firewall.Allocation); err != nil {
		return fmt.Errorf("failed to connect to firewall via SSH: %w", err)
	}

	return nil
}

func (c *firewallCmd) firewallPureSSH(fwAllocation *models.V1MachineAllocation) (err error) {
	networks := fwAllocation.Networks
	for _, nw := range networks {
		if *nw.Underlay || *nw.Private {
			continue
		}
		for _, ip := range nw.Ips {
			if portOpen(ip, "22", time.Second) {
				err = sshClient("metal", viper.GetString("identity"), ip, 22, nil)
				if err != nil {
					return err
				}

				return nil
			}
		}
	}

	return fmt.Errorf("no ip with a open ssh port found")
}
