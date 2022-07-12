package cmd

import (
	"errors"
	"fmt"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/network"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type networkCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	*genericcli.GenericCLI[*models.V1NetworkCreateRequest, *models.V1NetworkUpdateRequest, *models.V1NetworkResponse]
	childCLI *genericcli.GenericCLI[*models.V1NetworkAllocateRequest, any, *models.V1NetworkResponse]
}

func newNetworkCmd(c *config) *cobra.Command {
	w := networkCmd{
		c:          c.client,
		driver:     c.driver,
		GenericCLI: genericcli.NewGenericCLI[*models.V1NetworkCreateRequest, *models.V1NetworkUpdateRequest, *models.V1NetworkResponse](networkCRUD{Client: c.client}),
		childCLI:   genericcli.NewGenericCLI[*models.V1NetworkAllocateRequest, any, *models.V1NetworkResponse](networkChildCRUD{Client: c.client}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1NetworkCreateRequest, *models.V1NetworkUpdateRequest, *models.V1NetworkResponse]{
		gcli:     w.GenericCLI,
		singular: "network",
		plural:   "networks",

		createRequestFromCLI: w.createRequestFromCLI,

		availableSortKeys: sorters.NetworkSortKeys(),
		validArgsFunc:     c.comp.NetworkListCompletion,
	})

	allocateCmd := &cobra.Command{
		Use:   "allocate",
		Short: "allocate a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !viper.IsSet("file") {
				shared := viper.GetBool("shared")
				nat := false
				var destinationPrefixes []string
				if viper.GetBool("dmz") {
					shared = true
					destinationPrefixes = []string{"0.0.0.0/0"}
					nat = true
				}

				labels, err := genericcli.LabelsToMap(viper.GetStringSlice("labels"))
				if err != nil {
					return err
				}

				return w.childCLI.CreateAndPrint(&models.V1NetworkAllocateRequest{
					Description:         viper.GetString("description"),
					Name:                viper.GetString("name"),
					Partitionid:         viper.GetString("partition"),
					Projectid:           viper.GetString("project"),
					Shared:              shared,
					Labels:              labels,
					Destinationprefixes: destinationPrefixes,
					Nat:                 nat,
				}, defaultToYAMLPrinter())
			}

			return w.childCLI.CreateFromFileAndPrint(viper.GetString("file"), defaultToYAMLPrinter())
		},
		PreRun: bindPFlags,
	}
	freeCmd := &cobra.Command{
		Use:   "free <networkid>",
		Short: "free a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.childCLI.DeleteAndPrint(args, defaultToYAMLPrinter())
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}
	prefixCmd := &cobra.Command{
		Use:   "prefix",
		Short: "prefix management of a network",
	}
	prefixAddCmd := &cobra.Command{
		Use:   "add <networkid>",
		Short: "add a prefix to a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkPrefixAdd(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}
	prefixRemoveCmd := &cobra.Command{
		Use:   "remove <networkid>",
		Short: "remove a prefix from a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkPrefixRemove(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}

	cmds.createCmd.Flags().StringP("id", "", "", "id of the network to create. [optional]")
	cmds.createCmd.Flags().StringP("description", "d", "", "description of the network to create. [optional]")
	cmds.createCmd.Flags().StringP("name", "n", "", "name of the network to create. [optional]")
	cmds.createCmd.Flags().StringP("partition", "p", "", "partition where this network should exist.")
	cmds.createCmd.Flags().StringP("project", "", "", "project of the network to create. [optional]")
	cmds.createCmd.Flags().StringSlice("prefixes", []string{}, "prefixes in this network.")
	cmds.createCmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
	cmds.createCmd.Flags().StringSlice("destinationprefixes", []string{}, "destination prefixes in this network.")
	cmds.createCmd.Flags().BoolP("primary", "", false, "set primary flag of network, if set to true, this network is used to start machines there.")
	cmds.createCmd.Flags().BoolP("nat", "", false, "set nat flag of network, if set to true, traffic from this network will be natted.")
	cmds.createCmd.Flags().BoolP("underlay", "", false, "set underlay flag of network, if set to true, this is used to transport underlay network traffic")
	cmds.createCmd.Flags().Int64P("vrf", "", 0, "vrf of this network")
	cmds.createCmd.Flags().BoolP("vrfshared", "", false, "vrf shared allows multiple networks to share a vrf")
	must(cmds.createCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(cmds.createCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	cmds.listCmd.Flags().StringP("id", "", "", "ID to filter [optional]")
	cmds.listCmd.Flags().StringP("name", "", "", "name to filter [optional]")
	cmds.listCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")
	cmds.listCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	cmds.listCmd.Flags().StringP("parent", "", "", "parent network to filter [optional]")
	cmds.listCmd.Flags().BoolP("nat", "", false, "nat to filter [optional]")
	cmds.listCmd.Flags().BoolP("privatesuper", "", false, "privatesuper to filter [optional]")
	cmds.listCmd.Flags().BoolP("underlay", "", false, "underlay to filter [optional]")
	cmds.listCmd.Flags().Int64P("vrf", "", 0, "vrf to filter [optional]")
	cmds.listCmd.Flags().StringSlice("prefixes", []string{}, "prefixes to filter, use it like: --prefixes prefix1,prefix2.")
	cmds.listCmd.Flags().StringSlice("destination-prefixes", []string{}, "destination prefixes to filter, use it like: --destination-prefixes prefix1,prefix2.")
	must(cmds.listCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(cmds.listCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	allocateCmd.Flags().StringP("name", "n", "", "name of the network to create. [required]")
	allocateCmd.Flags().StringP("partition", "", "", "partition where this network should exist. [required]")
	allocateCmd.Flags().StringP("project", "", "", "partition where this network should exist. [required]")
	allocateCmd.Flags().StringP("description", "d", "", "description of the network to create. [optional]")
	allocateCmd.Flags().StringSlice("labels", []string{}, "labels for this network. [optional]")
	allocateCmd.Flags().BoolP("dmz", "", false, "use this private network as dmz. [optional]")
	allocateCmd.Flags().BoolP("shared", "", false, "shared allows usage of this private network from other networks")
	must(allocateCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(allocateCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	must(allocateCmd.MarkFlagRequired("name"))
	must(allocateCmd.MarkFlagRequired("project"))
	must(allocateCmd.MarkFlagRequired("partition"))

	prefixAddCmd.Flags().StringP("prefix", "", "", "prefix to add.")
	prefixRemoveCmd.Flags().StringP("prefix", "", "", "prefix to remove.")
	prefixCmd.AddCommand(prefixAddCmd)
	prefixCmd.AddCommand(prefixRemoveCmd)

	return cmds.buildRootCmd(
		newIPCmd(c),
		allocateCmd,
		freeCmd,
		prefixCmd,
	)
}

type networkCRUD struct {
	metalgo.Client
}

func (c networkCRUD) Get(id string) (*models.V1NetworkResponse, error) {
	resp, err := c.Network().FindNetwork(network.NewFindNetworkParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkCRUD) List() ([]*models.V1NetworkResponse, error) {
	resp, err := c.Network().FindNetworks(network.NewFindNetworksParams().WithBody(&models.V1NetworkFindRequest{
		ID:                  viper.GetString("id"),
		Name:                viper.GetString("name"),
		Partitionid:         viper.GetString("partition"),
		Projectid:           viper.GetString("project"),
		Nat:                 viper.GetBool("nat"),
		Privatesuper:        viper.GetBool("privatesuper"),
		Underlay:            viper.GetBool("underlay"),
		Vrf:                 viper.GetInt64("vrf"),
		Prefixes:            viper.GetStringSlice("prefixes"),
		Destinationprefixes: viper.GetStringSlice("destination-prefixes"),
		Parentnetworkid:     viper.GetString("parent"),
	}), nil)
	if err != nil {
		return nil, err
	}

	err = sorters.NetworkSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkCRUD) Delete(id string) (*models.V1NetworkResponse, error) {
	resp, err := c.Network().DeleteNetwork(network.NewDeleteNetworkParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkCRUD) Create(rq *models.V1NetworkCreateRequest) (*models.V1NetworkResponse, error) {
	resp, err := c.Network().CreateNetwork(network.NewCreateNetworkParams().WithBody(rq), nil)
	if err != nil {
		var r *network.CreateNetworkConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkCRUD) Update(rq *models.V1NetworkUpdateRequest) (*models.V1NetworkResponse, error) {
	resp, err := c.Network().UpdateNetwork(network.NewUpdateNetworkParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *networkCmd) createRequestFromCLI() (*models.V1NetworkCreateRequest, error) {
	lbs, err := genericcli.LabelsToMap(viper.GetStringSlice("labels"))
	if err != nil {
		return nil, err
	}

	return &models.V1NetworkCreateRequest{
		ID:                  pointer.Pointer(viper.GetString("id")),
		Description:         viper.GetString("description"),
		Name:                viper.GetString("name"),
		Partitionid:         viper.GetString("partition"),
		Projectid:           viper.GetString("project"),
		Prefixes:            viper.GetStringSlice("prefixes"),
		Destinationprefixes: viper.GetStringSlice("destinationprefixes"),
		Privatesuper:        pointer.Pointer(viper.GetBool("privatesuper")),
		Nat:                 pointer.Pointer(viper.GetBool("nat")),
		Underlay:            pointer.Pointer(viper.GetBool("underlay")),
		Vrf:                 viper.GetInt64("vrf"),
		Vrfshared:           viper.GetBool("vrfshared"),
		Labels:              lbs,
	}, nil
}

type networkChildCRUD struct {
	metalgo.Client
}

func (c networkChildCRUD) Get(id string) (*models.V1NetworkResponse, error) {
	return nil, fmt.Errorf("not implemented for child netowrks, use network update")
}

func (c networkChildCRUD) List() ([]*models.V1NetworkResponse, error) {
	return nil, fmt.Errorf("not implemented for child netowrks, use network update")
}

func (c networkChildCRUD) Delete(id string) (*models.V1NetworkResponse, error) {
	resp, err := c.Network().FreeNetwork(network.NewFreeNetworkParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkChildCRUD) Create(rq *models.V1NetworkAllocateRequest) (*models.V1NetworkResponse, error) {
	resp, err := c.Network().AllocateNetwork(network.NewAllocateNetworkParams().WithBody(rq), nil)
	if err != nil {
		var r *network.AllocateNetworkConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkChildCRUD) Update(rq any) (*models.V1NetworkResponse, error) {
	return nil, fmt.Errorf("not implemented for child netowrks, use network update")
}

// non-generic command handling

func (c *config) networkPrefixAdd(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	nur := &metalgo.NetworkUpdateRequest{
		Networkid: id,
		Prefix:    viper.GetString("prefix"),
	}
	resp, err := c.driver.NetworkAddPrefix(nur)
	if err != nil {
		return err
	}

	return defaultToYAMLPrinter().Print(resp.Network)
}

func (c *config) networkPrefixRemove(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	nur := &metalgo.NetworkUpdateRequest{
		Networkid: id,
		Prefix:    viper.GetString("prefix"),
	}
	resp, err := c.driver.NetworkRemovePrefix(nur)
	if err != nil {
		return err
	}

	return defaultToYAMLPrinter().Print(resp.Network)
}
