package cmd

import (
	"fmt"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/firewall"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type firewallCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	*genericcli.GenericCLI[*models.V1FirewallCreateRequest, any, *models.V1FirewallResponse]
}

func newFirewallCmd(c *config) *cobra.Command {
	w := firewallCmd{
		c:          c.client,
		driver:     c.driver,
		GenericCLI: genericcli.NewGenericCLI[*models.V1FirewallCreateRequest, any, *models.V1FirewallResponse](firewallCRUD{Client: c.client}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1FirewallCreateRequest, any, *models.V1FirewallResponse]{
		gcli:        w.GenericCLI,
		singular:    "firewall",
		plural:      "firewalls",
		description: "firewalls are used to establish network connectivity between metal-stack networks. firewalls are similar to machines but are managed by the provider. almost every command of the machine command subset works on firewalls, too.",
		aliases:     []string{"fw"},

		createRequestFromCLI: w.createRequestFromCLI,
		validArgsFunc:        c.comp.FirewallListCompletion,
	})

	c.addMachineCreateFlags(cmds.createCmd, "firewall")

	cmds.createCmd.Use = "allocate"

	cmds.listCmd.Flags().String("id", "", "ID to filter [optional]")
	cmds.listCmd.Flags().String("partition", "", "partition to filter [optional]")
	cmds.listCmd.Flags().String("size", "", "size to filter [optional]")
	cmds.listCmd.Flags().String("name", "", "allocation name to filter [optional]")
	cmds.listCmd.Flags().String("project", "", "allocation project to filter [optional]")
	cmds.listCmd.Flags().String("image", "", "allocation image to filter [optional]")
	cmds.listCmd.Flags().String("hostname", "", "allocation hostname to filter [optional]")
	cmds.listCmd.Flags().StringSlice("mac", []string{}, "mac to filter [optional]")
	cmds.listCmd.Flags().StringSlice("tags", []string{}, "tags to filter, use it like: --tags \"tag1,tag2\" or --tags \"tag3\".")
	must(cmds.listCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
	must(cmds.listCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(cmds.listCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(cmds.listCmd.RegisterFlagCompletionFunc("id", c.comp.FirewallListCompletion))
	must(cmds.listCmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))

	return cmds.BuildRootCmd()
}

type firewallCRUD struct {
	metalgo.Client
}

func (c firewallCRUD) Get(id string) (*models.V1FirewallResponse, error) {
	resp, err := c.Firewall().FindFirewall(firewall.NewFindFirewallParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c firewallCRUD) List() ([]*models.V1FirewallResponse, error) {
	resp, err := c.Firewall().FindFirewalls(firewall.NewFindFirewallsParams().WithBody(&models.V1FirewallFindRequest{
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

func (c firewallCRUD) Delete(_ string) (*models.V1FirewallResponse, error) {
	return nil, fmt.Errorf("firewall entity does not support delete operation, use machine delete")
}

func (c firewallCRUD) Create(rq *models.V1FirewallCreateRequest) (*models.V1FirewallResponse, error) {
	resp, err := c.Firewall().AllocateFirewall(firewall.NewAllocateFirewallParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c firewallCRUD) Update(rq any) (*models.V1FirewallResponse, error) {
	return nil, fmt.Errorf("firewall entity does not support update operation, use machine update")
}

func (c *firewallCmd) createRequestFromCLI() (*models.V1FirewallCreateRequest, error) {
	mcr, err := machineCreateRequest()
	if err != nil {
		return nil, fmt.Errorf("firewall create error:%w", err)
	}

	var nets []*models.V1MachineAllocationNetwork
	for i := range mcr.Networks {
		net := models.V1MachineAllocationNetwork{
			Networkid:   &mcr.Networks[i].NetworkID,
			Autoacquire: &mcr.Networks[i].Autoacquire,
		}
		nets = append(nets, &net)
	}

	return &models.V1FirewallCreateRequest{
		Description: mcr.Description,
		Partitionid: &mcr.Partition,
		Hostname:    mcr.Hostname,
		Imageid:     &mcr.Image,
		Name:        mcr.Name,
		UUID:        mcr.UUID,
		Projectid:   &mcr.Project,
		Sizeid:      &mcr.Size,
		SSHPubKeys:  mcr.SSHPublicKeys,
		UserData:    mcr.UserData,
		Tags:        mcr.Tags,
		Networks:    nets,
		Ips:         mcr.IPs,
	}, nil
}
