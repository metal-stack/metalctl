package cmd

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/client/firewall"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/printers"
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
		GenericCLI: genericcli.NewGenericCLI[*models.V1FirewallCreateRequest, any, *models.V1FirewallResponse](w),
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
		AvailableSortKeys:    sorters.FirewallSortKeys(),
		DescribePrinter:      printers.DefaultToYAMLPrinter(),
		ListPrinter:          printers.NewPrinterFromCLI(),
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

	return genericcli.NewCmds(cmdsConfig)
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

	err = sorters.FirewallSort(resp.Payload)
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

func (c *firewallCmd) createRequestFromCLI() (*models.V1FirewallCreateRequest, error) {
	mcr, err := machineCreateRequest()
	if err != nil {
		return nil, fmt.Errorf("firewall create error:%w", err)
	}

	return &models.V1FirewallCreateRequest{
		Description: mcr.Description,
		Partitionid: mcr.Partitionid,
		Hostname:    mcr.Hostname,
		Imageid:     mcr.Imageid,
		Name:        mcr.Name,
		UUID:        mcr.UUID,
		Projectid:   mcr.Projectid,
		Sizeid:      mcr.Sizeid,
		SSHPubKeys:  mcr.SSHPubKeys,
		UserData:    mcr.UserData,
		Tags:        mcr.Tags,
		Networks:    mcr.Networks,
		Ips:         mcr.Ips,
	}, nil
}
