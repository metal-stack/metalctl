package cmd

import (
	"errors"
	"fmt"

	"github.com/metal-stack/metal-go/api/client/network"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/util/sets"
)

type networkCmd struct {
	*config
	childCLI *genericcli.GenericCLI[*models.V1NetworkAllocateRequest, any, *models.V1NetworkResponse]
}

func newNetworkCmd(c *config) *cobra.Command {
	w := networkCmd{
		config:   c,
		childCLI: genericcli.NewGenericCLI[*models.V1NetworkAllocateRequest, any, *models.V1NetworkResponse](networkChildCRUD{config: c}).WithFS(c.fs),
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1NetworkCreateRequest, *models.V1NetworkUpdateRequest, *models.V1NetworkResponse]{
		BinaryName:           binaryName,
		GenericCLI:           genericcli.NewGenericCLI[*models.V1NetworkCreateRequest, *models.V1NetworkUpdateRequest, *models.V1NetworkResponse](w).WithFS(c.fs),
		Singular:             "network",
		Plural:               "networks",
		Description:          "networks can be attached to a machine or firewall such that they can communicate with each other.",
		CreateRequestFromCLI: w.createRequestFromCLI,
		UpdateRequestFromCLI: w.updateRequestFromCLI,
		Sorter:               sorters.NetworkSorter(),
		ValidArgsFn:          c.comp.NetworkListCompletion,
		DescribePrinter:      func() printers.Printer { return c.describePrinter },
		ListPrinter:          func() printers.Printer { return c.listPrinter },
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "", "", "id of the network to create. [optional]")
			cmd.Flags().StringP("description", "d", "", "description of the network to create. [optional]")
			cmd.Flags().StringP("name", "n", "", "name of the network to create. [optional]")
			cmd.Flags().StringP("partition", "p", "", "partition where this network should exist.")
			cmd.Flags().StringP("project", "", "", "project of the network to create. [optional]")
			cmd.Flags().StringSlice("prefixes", []string{}, "prefixes in this network.")
			cmd.Flags().StringSlice("labels", []string{}, "add initial labels, must be in the form of key=value, use it like: --labels \"key1=value1,key2=value2\".")
			cmd.Flags().StringSlice("destination-prefixes", []string{}, "destination prefixes in this network.")
			cmd.Flags().BoolP("privatesuper", "", false, "set private super flag of network, if set to true, this network is used to start machines there.")
			cmd.Flags().BoolP("nat", "", false, "set nat flag of network, if set to true, traffic from this network will be natted.")
			cmd.Flags().BoolP("underlay", "", false, "set underlay flag of network, if set to true, this is used to transport underlay network traffic")
			cmd.Flags().Int64P("vrf", "", 0, "vrf of this network")
			cmd.Flags().BoolP("vrfshared", "", false, "vrf shared allows multiple networks to share a vrf")
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "ID to filter [optional]")
			cmd.Flags().String("name", "", "name to filter [optional]")
			cmd.Flags().String("partition", "", "partition to filter [optional]")
			cmd.Flags().String("project", "", "project to filter [optional]")
			cmd.Flags().String("parent", "", "parent network to filter [optional]")
			cmd.Flags().BoolP("nat", "", false, "nat to filter [optional]")
			cmd.Flags().BoolP("privatesuper", "", false, "privatesuper to filter [optional]")
			cmd.Flags().BoolP("underlay", "", false, "underlay to filter [optional]")
			cmd.Flags().Int64P("vrf", "", 0, "vrf to filter [optional]")
			cmd.Flags().StringSlice("prefixes", []string{}, "prefixes to filter, use it like: --prefixes prefix1,prefix2.")
			cmd.Flags().StringSlice("destination-prefixes", []string{}, "destination prefixes to filter, use it like: --destination-prefixes prefix1,prefix2.")
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
		},
		UpdateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("name", "", "the name of the network [optional]")
			cmd.Flags().String("description", "", "the description of the network [optional]")
			cmd.Flags().StringSlice("add-prefixes", []string{}, "prefixes to be added to the network [optional]")
			cmd.Flags().StringSlice("remove-prefixes", []string{}, "prefixes to be removed from the network [optional]")
			cmd.Flags().StringSlice("add-destinationprefixes", []string{}, "destination prefixes to be added to the network [optional]")
			cmd.Flags().StringSlice("remove-destinationprefixes", []string{}, "destination prefixes to be removed from the network [optional]")
			cmd.Flags().StringSlice("labels", []string{}, "the labels of the network, must be in the form of key=value, use it like: --labels \"key1=value1,key2=value2\". [optional]")
			cmd.Flags().Bool("shared", false, "marks a network as shared or not [optional]")
		},
	}

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
				}, c.describePrinter)
			}

			return w.childCLI.CreateFromFileAndPrint(viper.GetString("file"), c.describePrinter)
		},
	}

	freeCmd := &cobra.Command{
		Use:   "free <networkid>",
		Short: "free a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := genericcli.GetExactlyOneArg(args)
			if err != nil {
				return err
			}

			return w.childCLI.DeleteAndPrint(id, c.describePrinter)
		},
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}

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

	return genericcli.NewCmds(
		cmdsConfig,
		newIPCmd(c),
		allocateCmd,
		freeCmd,
	)
}

func (c networkCmd) Get(id string) (*models.V1NetworkResponse, error) {
	resp, err := c.client.Network().FindNetwork(network.NewFindNetworkParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkCmd) List() ([]*models.V1NetworkResponse, error) {
	resp, err := c.client.Network().FindNetworks(network.NewFindNetworksParams().WithBody(&models.V1NetworkFindRequest{
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

	return resp.Payload, nil
}

func (c networkCmd) Delete(id string) (*models.V1NetworkResponse, error) {
	resp, err := c.client.Network().DeleteNetwork(network.NewDeleteNetworkParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkCmd) Create(rq *models.V1NetworkCreateRequest) (*models.V1NetworkResponse, error) {
	resp, err := c.client.Network().CreateNetwork(network.NewCreateNetworkParams().WithBody(rq), nil)
	if err != nil {
		var r *network.CreateNetworkConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkCmd) Update(rq *models.V1NetworkUpdateRequest) (*models.V1NetworkResponse, error) {
	resp, err := c.client.Network().UpdateNetwork(network.NewUpdateNetworkParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkCmd) Convert(r *models.V1NetworkResponse) (string, *models.V1NetworkCreateRequest, *models.V1NetworkUpdateRequest, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.ID, networkResponseToCreate(r), networkResponseToUpdate(r), nil
}

func networkResponseToCreate(r *models.V1NetworkResponse) *models.V1NetworkCreateRequest {
	return &models.V1NetworkCreateRequest{
		Description:         r.Description,
		Destinationprefixes: r.Destinationprefixes,
		ID:                  r.ID,
		Labels:              r.Labels,
		Name:                r.Name,
		Nat:                 r.Nat,
		Parentnetworkid:     r.Parentnetworkid,
		Partitionid:         r.Partitionid,
		Prefixes:            r.Prefixes,
		Privatesuper:        r.Privatesuper,
		Projectid:           r.Projectid,
		Shared:              r.Shared,
		Underlay:            r.Underlay,
		Vrf:                 r.Vrf,
		Vrfshared:           r.Vrfshared,
	}
}

func networkResponseToUpdate(r *models.V1NetworkResponse) *models.V1NetworkUpdateRequest {
	return &models.V1NetworkUpdateRequest{
		Description:         r.Description,
		Destinationprefixes: r.Destinationprefixes,
		ID:                  r.ID,
		Labels:              r.Labels,
		Name:                r.Name,
		Prefixes:            r.Prefixes,
		Shared:              r.Shared,
	}
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
		Destinationprefixes: viper.GetStringSlice("destination-prefixes"),
		Privatesuper:        pointer.Pointer(viper.GetBool("privatesuper")),
		Nat:                 pointer.Pointer(viper.GetBool("nat")),
		Underlay:            pointer.Pointer(viper.GetBool("underlay")),
		Vrf:                 viper.GetInt64("vrf"),
		Vrfshared:           viper.GetBool("vrfshared"),
		Labels:              lbs,
	}, nil
}

type networkChildCRUD struct {
	*config
}

func (c networkChildCRUD) Get(id string) (*models.V1NetworkResponse, error) {
	return nil, fmt.Errorf("not implemented for child netowrks, use network update")
}

func (c networkChildCRUD) List() ([]*models.V1NetworkResponse, error) {
	return nil, fmt.Errorf("not implemented for child netowrks, use network update")
}

func (c networkChildCRUD) Delete(id string) (*models.V1NetworkResponse, error) {
	resp, err := c.client.Network().FreeNetwork(network.NewFreeNetworkParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c networkChildCRUD) Create(rq *models.V1NetworkAllocateRequest) (*models.V1NetworkResponse, error) {
	resp, err := c.client.Network().AllocateNetwork(network.NewAllocateNetworkParams().WithBody(rq), nil)
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

func (c networkChildCRUD) Convert(r *models.V1NetworkResponse) (string, *models.V1NetworkAllocateRequest, any, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.ID, &models.V1NetworkAllocateRequest{
		Description:         r.Description,
		Destinationprefixes: r.Destinationprefixes,
		Labels:              r.Labels,
		Name:                r.Name,
		Nat:                 pointer.SafeDeref(r.Nat),
		Partitionid:         r.Partitionid,
		Projectid:           r.Projectid,
		Shared:              false,
	}, nil, nil
}

func (c *networkCmd) updateRequestFromCLI(args []string) (*models.V1NetworkUpdateRequest, error) {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return nil, err
	}

	resp, err := c.Get(id)
	if err != nil {
		return nil, err
	}

	var labels map[string]string
	if viper.IsSet("labels") {
		labels, err = genericcli.LabelsToMap(viper.GetStringSlice("labels"))
		if err != nil {
			return nil, err
		}
	}

	shared := resp.Shared
	if viper.IsSet("shared") {
		shared = viper.GetBool("shared")
	}

	var (
		ur = &models.V1NetworkUpdateRequest{
			Description:         viper.GetString("description"),
			Destinationprefixes: nil,
			ID:                  pointer.Pointer(id),
			Labels:              labels,
			Name:                viper.GetString("name"),
			Prefixes:            nil,
			Shared:              shared,
		}
		addPrefixes                = sets.New(viper.GetStringSlice("add-prefixes")...)
		removePrefixes             = sets.New(viper.GetStringSlice("remove-prefixes")...)
		addDestinationprefixes     = sets.New(viper.GetStringSlice("add-destinationprefixes")...)
		removeDestinationprefixes  = sets.New(viper.GetStringSlice("remove-destinationprefixes")...)
		currentPrefixes            = sets.New(resp.Prefixes...)
		currentDestinationprefixes = sets.New(resp.Destinationprefixes...)
	)

	newPrefixes := currentPrefixes.Clone()
	if viper.IsSet("remove-prefixes") {
		diff := removePrefixes.Difference(currentPrefixes)
		if diff.Len() > 0 {
			difflist := diff.UnsortedList()
			slices.Sort(difflist)
			return nil, fmt.Errorf("cannot remove prefixes because they are currently not present: %s", difflist)
		}
		newPrefixes = newPrefixes.Difference(removePrefixes)
	}
	if viper.IsSet("add-prefixes") {
		if currentPrefixes.HasAny(addPrefixes.UnsortedList()...) {
			intersection := addPrefixes.Intersection(currentPrefixes).UnsortedList()
			slices.Sort(intersection)
			return nil, fmt.Errorf("cannot add prefixes because they are already present: %s", intersection)
		}
		newPrefixes = newPrefixes.Union(addPrefixes)
	}
	if !newPrefixes.Equal(currentPrefixes) {
		ur.Prefixes = newPrefixes.UnsortedList()
	}

	newDestinationprefixes := currentDestinationprefixes.Clone()
	if viper.IsSet("remove-destinationprefixes") {
		diff := removeDestinationprefixes.Difference(currentDestinationprefixes)
		if diff.Len() > 0 {
			difflist := diff.UnsortedList()
			slices.Sort(difflist)
			return nil, fmt.Errorf("cannot remove destination prefixes because they are currently not present: %s", difflist)
		}
		newDestinationprefixes = newDestinationprefixes.Difference(removeDestinationprefixes)
	}
	if viper.IsSet("add-destinationprefixes") {
		if currentDestinationprefixes.HasAny(addDestinationprefixes.UnsortedList()...) {
			interSection := addDestinationprefixes.Intersection(currentDestinationprefixes).UnsortedList()
			slices.Sort(interSection)
			return nil, fmt.Errorf("cannot add destination prefixes because they are already present: %s", interSection)
		}
		newDestinationprefixes = newDestinationprefixes.Union(addDestinationprefixes)
	}
	if !newDestinationprefixes.Equal(currentDestinationprefixes) {
		ur.Destinationprefixes = newDestinationprefixes.UnsortedList()
	}

	return ur, nil
}
