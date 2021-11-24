package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	ipmodel "github.com/metal-stack/metal-go/api/client/ip"
	networkmodel "github.com/metal-stack/metal-go/api/client/network"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/tag"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func newNetworkCmd(c *config) *cobra.Command {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "manage networks",
		Long:  "networks for metal.",
	}

	networkListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkList()
		},
		PreRun: bindPFlags,
	}
	networkCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkCreate()
		},
		PreRun: bindPFlags,
	}
	networkDescribeCmd := &cobra.Command{
		Use:   "describe <networkid>",
		Short: "describe a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkDescribe(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}
	networkAllocateCmd := &cobra.Command{
		Use:   "allocate",
		Short: "allocate a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkAllocate()
		},
		PreRun: bindPFlags,
	}
	networkFreeCmd := &cobra.Command{
		Use:   "free <networkid>",
		Short: "free a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkFree(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}
	networkDeleteCmd := &cobra.Command{
		Use:     "delete <networkID>",
		Short:   "delete a network",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkDelete(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}
	networkPrefixCmd := &cobra.Command{
		Use:   "prefix",
		Short: "prefix management of a network",
	}

	networkPrefixAddCmd := &cobra.Command{
		Use:   "add <networkid>",
		Short: "add a prefix to a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkPrefixAdd(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}
	networkPrefixRemoveCmd := &cobra.Command{
		Use:   "remove <networkid>",
		Short: "remove a prefix from a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkPrefixRemove(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}

	networkDestinationprefixCmd := &cobra.Command{
		Use:   "destinationprefix",
		Short: "destination prefix management of a network",
	}

	networkDestinationprefixAddCmd := &cobra.Command{
		Use:   "add <networkid>",
		Short: "add a destination prefix to a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkDestinationprefixAdd(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}
	networkDestinationprefixRemoveCmd := &cobra.Command{
		Use:   "remove <networkid>",
		Short: "remove a destination prefix from a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkDestinationprefixRemove(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.NetworkListCompletion,
	}

	networkIPCmd := &cobra.Command{
		Use:   "ip",
		Short: "manage IPs",
	}

	networkIPListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "manage IPs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipList()
		},
		PreRun: bindPFlags,
	}

	networkIPAllocateCmd := &cobra.Command{
		Use:   "allocate",
		Short: "allocate an IP, if non given the next free is allocated, otherwise the given IP is checked for availability.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipAllocate(args)
		},
		PreRun: bindPFlags,
	}
	networkIPFreeCmd := &cobra.Command{
		Use:   "free <IP>",
		Short: "free an IP",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipFree(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.IpListCompletion,
	}
	networkIPApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update an IP",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipApply()
		},
		PreRun: bindPFlags,
	}
	networkIPEditCmd := &cobra.Command{
		Use:   "edit <IP>",
		Short: "edit a ip",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipEdit(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.IpListCompletion,
	}

	networkApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.networkApply()
		},
		PreRun: bindPFlags,
	}

	networkIPIssuesCmd := &cobra.Command{
		Use:   "issues",
		Short: `display ips which are in a potential bad state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.ipIssues()
		},
		PreRun: bindPFlags,
	}

	// TODO add completions for project, partition,

	networkCreateCmd.Flags().StringP("id", "", "", "id of the network to create. [optional]")
	networkCreateCmd.Flags().StringP("description", "d", "", "description of the network to create. [optional]")
	networkCreateCmd.Flags().StringP("name", "n", "", "name of the network to create. [optional]")
	networkCreateCmd.Flags().StringP("partition", "p", "", "partition where this network should exist.")
	networkCreateCmd.Flags().StringSlice("prefixes", []string{}, "prefixes in this network.")
	networkCreateCmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
	networkCreateCmd.Flags().StringSlice("destinationprefixes", []string{}, "destination prefixes in this network.")
	networkCreateCmd.Flags().BoolP("primary", "", false, "set primary flag of network, if set to true, this network is used to start machines there.")
	networkCreateCmd.Flags().BoolP("nat", "", false, "set nat flag of network, if set to true, traffic from this network will be natted.")
	networkCreateCmd.Flags().BoolP("underlay", "", false, "set underlay flag of network, if set to true, this is used to transport underlay network traffic")
	networkCreateCmd.Flags().Int64P("vrf", "", 0, "vrf of this network")
	networkCreateCmd.Flags().BoolP("vrfshared", "", false, "vrf shared allows multiple networks to share a vrf")

	networkAllocateCmd.Flags().StringP("name", "n", "", "name of the network to create. [required]")
	networkAllocateCmd.Flags().StringP("partition", "", "", "partition where this network should exist. [required]")
	networkAllocateCmd.Flags().StringP("project", "", "", "partition where this network should exist. [required]")
	networkAllocateCmd.Flags().StringP("description", "d", "", "description of the network to create. [optional]")
	networkAllocateCmd.Flags().StringSlice("labels", []string{}, "labels for this network. [optional]")
	networkAllocateCmd.Flags().BoolP("dmz", "", false, "use this private network as dmz. [optional]")
	networkAllocateCmd.Flags().BoolP("shared", "", false, "shared allows usage of this private network from other networks")

	must(networkAllocateCmd.MarkFlagRequired("name"))
	must(networkAllocateCmd.MarkFlagRequired("project"))
	must(networkAllocateCmd.MarkFlagRequired("partition"))

	networkIPAllocateCmd.Flags().StringP("description", "d", "", "description of the IP to allocate. [optional]")
	networkIPAllocateCmd.Flags().StringP("name", "n", "", "name of the IP to allocate. [optional]")
	networkIPAllocateCmd.Flags().StringP("type", "", metalgo.IPTypeEphemeral, "type of the IP to allocate: "+metalgo.IPTypeEphemeral+"|"+metalgo.IPTypeStatic+" [optional]")
	networkIPAllocateCmd.Flags().StringP("network", "", "", "network from where the IP should be allocated.")
	networkIPAllocateCmd.Flags().StringP("project", "", "", "project for which the IP should be allocated.")
	networkIPAllocateCmd.Flags().StringSliceP("tags", "", nil, "tags to attach to the IP.")

	networkIPApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.`)
	must(networkIPApplyCmd.MarkFlagRequired("file"))

	networkIPListCmd.Flags().StringP("ipaddress", "", "", "ipaddress to filter [optional]")
	networkIPListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	networkIPListCmd.Flags().StringP("prefix", "", "", "prefx to filter [optional]")
	networkIPListCmd.Flags().StringP("machineid", "", "", "machineid to filter [optional]")
	networkIPListCmd.Flags().StringP("type", "", "", "type to filter [optional]")
	networkIPListCmd.Flags().StringP("network", "", "", "network to filter [optional]")
	networkIPListCmd.Flags().StringP("name", "", "", "name to filter [optional]")
	networkIPListCmd.Flags().StringSliceP("tags", "", nil, "tags to filter [optional]")

	networkIPCmd.AddCommand(networkIPListCmd)
	networkIPCmd.AddCommand(networkIPAllocateCmd)
	networkIPCmd.AddCommand(networkIPFreeCmd)
	networkIPCmd.AddCommand(networkIPApplyCmd)
	networkIPCmd.AddCommand(networkIPEditCmd)
	networkIPCmd.AddCommand(networkIPIssuesCmd)

	networkPrefixAddCmd.Flags().StringP("prefix", "", "", "prefix to add.")
	networkPrefixRemoveCmd.Flags().StringP("prefix", "", "", "prefix to remove.")
	networkPrefixCmd.AddCommand(networkPrefixAddCmd)
	networkPrefixCmd.AddCommand(networkPrefixRemoveCmd)

	networkDestinationprefixAddCmd.Flags().StringP("destinationprefix", "", "", "destination prefix to add.")
	networkDestinationprefixRemoveCmd.Flags().StringP("destinationprefix", "", "", "destination prefix to remove.")
	networkPrefixCmd.AddCommand(networkDestinationprefixAddCmd)
	networkPrefixCmd.AddCommand(networkDestinationprefixRemoveCmd)

	networkApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl network describe internet > internet.yaml
# vi internet.yaml
## either via stdin
# cat internet.yaml | metalctl network apply -f -
## or via file
# metalctl network apply -f internet.yaml`)
	must(networkApplyCmd.MarkFlagRequired("file"))

	networkListCmd.Flags().StringP("id", "", "", "ID to filter [optional]")
	networkListCmd.Flags().StringP("name", "", "", "name to filter [optional]")
	networkListCmd.Flags().StringP("partition", "", "", "partition to filter [optional]")
	networkListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	networkListCmd.Flags().StringP("parent", "", "", "parent network to filter [optional]")
	networkListCmd.Flags().BoolP("nat", "", false, "nat to filter [optional]")
	networkListCmd.Flags().BoolP("privatesuper", "", false, "privatesuper to filter [optional]")
	networkListCmd.Flags().BoolP("underlay", "", false, "underlay to filter [optional]")
	networkListCmd.Flags().Int64P("vrf", "", 0, "vrf to filter [optional]")
	networkListCmd.Flags().StringSlice("prefixes", []string{}, "prefixes to filter, use it like: --prefixes prefix1,prefix2.")
	networkListCmd.Flags().StringSlice("destination-prefixes", []string{}, "destination prefixes to filter, use it like: --destination-prefixes prefix1,prefix2.")

	networkCmd.AddCommand(networkIPCmd)
	networkCmd.AddCommand(networkListCmd)
	networkCmd.AddCommand(networkCreateCmd)
	networkCmd.AddCommand(networkDescribeCmd)
	networkCmd.AddCommand(networkAllocateCmd)
	networkCmd.AddCommand(networkFreeCmd)
	networkCmd.AddCommand(networkPrefixCmd)
	networkCmd.AddCommand(networkDestinationprefixCmd)
	networkCmd.AddCommand(networkApplyCmd)
	networkCmd.AddCommand(networkDeleteCmd)

	return networkCmd
}

func (c *config) networkList() error {
	var resp *metalgo.NetworkListResponse
	var err error
	if atLeastOneViperStringFlagGiven("id", "name", "partition", "project", "parent") ||
		atLeastOneViperBoolFlagGiven("nat", "primary", "underlay") ||
		atLeastOneViperInt64FlagGiven("vrf") ||
		atLeastOneViperStringSliceFlagGiven("prefixes", "destination-prefixes") {
		nfr := &metalgo.NetworkFindRequest{
			ID:                  viperString("id"),
			Name:                viperString("name"),
			PartitionID:         viperString("partition"),
			ProjectID:           viperString("project"),
			Nat:                 viperBool("nat"),
			PrivateSuper:        viperBool("privatesuper"),
			Underlay:            viperBool("underlay"),
			Vrf:                 viperInt64("vrf"),
			Prefixes:            viperStringSlice("prefixes"),
			DestinationPrefixes: viperStringSlice("destination-prefixes"),
			ParentNetworkID:     viperString("parent"),
		}
		resp, err = c.driver.NetworkFind(nfr)
	} else {
		resp, err = c.driver.NetworkList()
	}
	if err != nil {
		return fmt.Errorf("network list error:%w", err)
	}
	return output.New().Print(resp.Networks)
}

func (c *config) networkAllocate() error {
	var ncrs []metalgo.NetworkAllocateRequest
	var ncr metalgo.NetworkAllocateRequest
	if viper.GetString("file") != "" {
		err := readFrom(viper.GetString("file"), &ncr, func(data interface{}) {
			doc := data.(*metalgo.NetworkAllocateRequest)
			ncrs = append(ncrs, *doc)
		})
		if err != nil {
			return err
		}
		if len(ncrs) != 1 {
			return fmt.Errorf("network allocate error: more or less than one network given:%d", len(ncrs))
		}
		ncr = ncrs[0]
	} else {
		shared := viper.GetBool("shared")
		nat := false
		var destinationPrefixes []string
		if viper.GetBool("dmz") {
			shared = true
			destinationPrefixes = []string{"0.0.0.0/0"}
			nat = true
		}
		ncr = metalgo.NetworkAllocateRequest{
			Description:         viper.GetString("description"),
			Name:                viper.GetString("name"),
			PartitionID:         viper.GetString("partition"),
			ProjectID:           viper.GetString("project"),
			Shared:              shared,
			Labels:              labelsFromTags(viper.GetStringSlice("labels")),
			Destinationprefixes: destinationPrefixes,
			Nat:                 nat,
		}
	}
	resp, err := c.driver.NetworkAllocate(&ncr)
	if err != nil {
		return fmt.Errorf("network allocate error:%w", err)
	}
	return output.NewDetailer().Detail(resp.Network)
}

func (c *config) networkFree(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no network given")
	}
	nw := args[0]
	resp, err := c.driver.NetworkFree(nw)
	if err != nil {
		return fmt.Errorf("network allocate error:%w", err)
	}
	return output.NewDetailer().Detail(resp.Network)
}

func (c *config) networkDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no network given")
	}
	nw := args[0]
	resp, err := c.driver.NetworkDelete(nw)
	if err != nil {
		return fmt.Errorf("network delete error:%w", err)
	}
	return output.NewDetailer().Detail(resp.Network)
}

func (c *config) networkDescribe(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no network given")
	}
	nw := args[0]
	resp, err := c.driver.NetworkGet(nw)
	if err != nil {
		return fmt.Errorf("network describe error:%w", err)
	}
	return output.NewDetailer().Detail(resp.Network)
}

func (c *config) networkCreate() error {
	var ncrs []metalgo.NetworkCreateRequest
	var ncr metalgo.NetworkCreateRequest
	if viper.GetString("file") != "" {
		err := readFrom(viper.GetString("file"), &ncr, func(data interface{}) {
			doc := data.(*metalgo.NetworkCreateRequest)
			ncrs = append(ncrs, *doc)
		})
		if err != nil {
			return err
		}
		if len(ncrs) != 1 {
			return fmt.Errorf("network create error more or less than one network given:%d", len(ncrs))
		}
		ncr = ncrs[0]
	} else {
		lbs, err := annotationsAsMap(viper.GetStringSlice("labels"))
		if err != nil {
			return err
		}
		ncr = metalgo.NetworkCreateRequest{
			Description:         viper.GetString("description"),
			Name:                viper.GetString("name"),
			Partitionid:         viper.GetString("partition"),
			Prefixes:            viper.GetStringSlice("prefixes"),
			Destinationprefixes: viper.GetStringSlice("destinationprefixes"),
			PrivateSuper:        viper.GetBool("privatesuper"),
			Nat:                 viper.GetBool("nat"),
			Underlay:            viper.GetBool("underlay"),
			Vrf:                 viper.GetInt64("vrf"),
			VrfShared:           viper.GetBool("vrfshared"),
			Labels:              lbs,
		}
		id := viper.GetString("id")
		if len(id) > 0 {
			ncr.ID = &id
		}
	}
	resp, err := c.driver.NetworkCreate(&ncr)
	if err != nil {
		return fmt.Errorf("network create error:%w", err)
	}
	return output.NewDetailer().Detail(resp.Network)
}

// TODO: General apply method would be useful as these are quite a lot of lines and it's getting erroneous
func (c *config) networkApply() error {
	var iars []metalgo.NetworkCreateRequest
	var iar metalgo.NetworkCreateRequest
	err := readFrom(viper.GetString("file"), &iar, func(data interface{}) {
		doc := data.(*metalgo.NetworkCreateRequest)
		iars = append(iars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		iar = metalgo.NetworkCreateRequest{}
	})
	if err != nil {
		return err
	}

	var response []*models.V1NetworkResponse
	for _, nar := range iars {
		nar := nar
		if nar.ID == nil {
			resp, err := c.driver.NetworkCreate(&nar)
			if err != nil {
				return err
			}
			response = append(response, resp.Network)
			continue
		}

		resp, err := c.driver.NetworkGet(*nar.ID)
		if err != nil {
			var r *networkmodel.FindNetworkDefault
			if !errors.As(err, &r) {
				return err
			}
			if r.Code() != http.StatusNotFound {
				return err
			}
		}
		if resp.Network == nil {
			resp, err := c.driver.NetworkCreate(&nar)
			if err != nil {
				return err
			}
			response = append(response, resp.Network)
			continue
		}

		detailResp, err := c.driver.NetworkUpdate(&nar)
		if err != nil {
			return err
		}
		response = append(response, detailResp.Network)
	}
	return output.NewDetailer().Detail(response)
}

func (c *config) networkPrefixAdd(args []string) error {
	networkID, err := c.getNetworkID(args)
	if err != nil {
		return err
	}

	nur := &metalgo.NetworkUpdateRequest{
		Networkid: networkID,
		Prefix:    viper.GetString("prefix"),
	}
	resp, err := c.driver.NetworkAddPrefix(nur)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Network)
}

func (c *config) networkPrefixRemove(args []string) error {
	networkID, err := c.getNetworkID(args)
	if err != nil {
		return err
	}

	nur := &metalgo.NetworkUpdateRequest{
		Networkid: networkID,
		Prefix:    viper.GetString("prefix"),
	}
	resp, err := c.driver.NetworkRemovePrefix(nur)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Network)
}

func (c *config) networkDestinationprefixAdd(args []string) error {
	networkID, err := c.getNetworkID(args)
	if err != nil {
		return err
	}

	nur := &metalgo.NetworkUpdateRequest{
		Networkid: networkID,
		Prefix:    viper.GetString("destinationprefix"),
	}
	resp, err := c.driver.NetworkAddDestinationprefix(nur)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Network)
}

func (c *config) networkDestinationprefixRemove(args []string) error {
	networkID, err := c.getNetworkID(args)
	if err != nil {
		return err
	}

	nur := &metalgo.NetworkUpdateRequest{
		Networkid: networkID,
		Prefix:    viper.GetString("destinationprefix"),
	}
	resp, err := c.driver.NetworkRemoveDestinationprefix(nur)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Network)
}

func (c *config) ipList() error {
	var resp *metalgo.IPListResponse
	var err error
	if atLeastOneViperStringFlagGiven("ipaddress", "project", "prefix", "machineid", "network", "type", "tags", "name") {
		ifr := &metalgo.IPFindRequest{
			IPAddress:        viperString("ipaddress"),
			ProjectID:        viperString("project"),
			ParentPrefixCidr: viperString("prefix"),
			NetworkID:        viperString("network"),
			MachineID:        viperString("machineid"),
			Type:             viperString("type"),
			Tags:             viperStringSlice("tags"),
			Name:             viperString("name"),
		}
		resp, err = c.driver.IPFind(ifr)
	} else {
		resp, err = c.driver.IPList()
	}
	if err != nil {
		return fmt.Errorf("IP list error:%w", err)
	}
	return output.New().Print(resp.IPs)
}

func (c *config) ipApply() error {
	var iars []metalgo.IPAllocateRequest
	var iar metalgo.IPAllocateRequest
	err := readFrom(viper.GetString("file"), &iar, func(data interface{}) {
		doc := data.(*metalgo.IPAllocateRequest)
		iars = append(iars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		iar = metalgo.IPAllocateRequest{}
	})
	if err != nil {
		return err
	}

	var response []*models.V1IPResponse
	for _, iar := range iars {
		iar := iar
		if iar.IPAddress == "" {
			// acquire
			resp, err := c.driver.IPAllocate(&iar)
			if err != nil {
				return err
			}
			response = append(response, resp.IP)
			continue
		}
		i, err := c.driver.IPGet(iar.IPAddress)
		if err != nil {
			var r *ipmodel.FindIPDefault
			if !errors.As(err, &r) {
				return err
			}
			if r.Code() != http.StatusNotFound {
				return err
			}
		}

		if i == nil {
			resp, err := c.driver.IPAllocate(&iar)
			if err != nil {
				return err
			}
			response = append(response, resp.IP)
			continue
		}

		iur := metalgo.IPUpdateRequest{
			IPAddress:   *i.IP.Ipaddress,
			Name:        iar.Name,
			Description: iar.Description,
			Type:        iar.Type,
			Tags:        iar.Tags,
		}
		resp, err := c.driver.IPUpdate(&iur)
		if err != nil {
			return err
		}
		response = append(response, resp.IP)
	}

	return output.NewDetailer().Detail(response)
}

func (c *config) ipEdit(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no IP given")
	}
	ip := args[0]

	getFunc := func(ip string) ([]byte, error) {
		resp, err := c.driver.IPGet(ip)
		if err != nil {
			return nil, err
		}
		content, err := yaml.Marshal(resp.IP)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		iurs, err := readIPUpdateRequests(filename)
		if err != nil {
			return err
		}
		if len(iurs) != 1 {
			return fmt.Errorf("ip update error more or less than one ip given:%d", len(iurs))
		}
		uresp, err := c.driver.IPUpdate(&iurs[0])
		if err != nil {
			return err
		}
		return output.NewDetailer().Detail(uresp.IP)
	}

	return edit(ip, getFunc, updateFunc)
}

func readIPUpdateRequests(filename string) ([]metalgo.IPUpdateRequest, error) {
	var iurs []metalgo.IPUpdateRequest
	var iur metalgo.IPUpdateRequest
	err := readFrom(filename, &iur, func(data interface{}) {
		doc := data.(*metalgo.IPUpdateRequest)
		iurs = append(iurs, *doc)
	})
	if err != nil {
		return nil, err
	}
	if len(iurs) != 1 {
		return nil, fmt.Errorf("ip update error more or less than one ip given:%d", len(iurs))
	}
	return iurs, nil
}

func (c *config) ipAllocate(args []string) error {
	specificIP := ""
	if len(args) > 0 {
		specificIP = args[0]
	}
	iar := &metalgo.IPAllocateRequest{
		Description: viper.GetString("description"),
		Name:        viper.GetString("name"),
		Networkid:   viper.GetString("network"),
		Projectid:   viper.GetString("project"),
		IPAddress:   specificIP,
		Type:        viper.GetString("type"),
		Tags:        viper.GetStringSlice("tags"),
	}
	resp, err := c.driver.IPAllocate(iar)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.IP)
}

func (c *config) ipFree(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no IP given")
	}
	ip := args[0]
	resp, err := c.driver.IPFree(ip)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.IP)
}

func (c *config) getNetworkID(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("no network ID given")
	}

	networkID := args[0]
	_, err := c.driver.NetworkGet(networkID)
	if err != nil {
		return "", err
	}
	return networkID, nil
}

func (c *config) ipIssues() error {
	ml, err := c.driver.MachineList()
	if err != nil {
		return fmt.Errorf("machine list error:%w", err)
	}
	machines := make(map[string]*models.V1MachineResponse)
	for _, m := range ml.Machines {
		machines[*m.ID] = m
	}

	var resp []*models.V1IPResponse

	iplist, err := c.driver.IPList()
	if err != nil {
		return err
	}
	for _, ip := range iplist.IPs {
		if *ip.Type == metalgo.IPTypeStatic {
			continue
		}
		if ip.Description == "autoassigned" && len(ip.Tags) == 0 {
			ip.Description = fmt.Sprintf("%s, but no tags", ip.Description)
			resp = append(resp, ip)
		}
		if strings.HasPrefix(ip.Name, "metallb-") && len(ip.Tags) == 0 {
			ip.Description = fmt.Sprintf("metallb ip without tags %s", ip.Description)
			resp = append(resp, ip)
		}

		for _, t := range ip.Tags {
			if strings.HasPrefix(t, tag.MachineID+"=") {
				parts := strings.Split(t, "=")
				m := machines[parts[1]]
				if m == nil || *m.Liveliness != "Alive" || m.Allocation == nil || *m.Events.Log[0].Event != "Phoned Home" {
					ip.Description = "bound to unallocated machine"
					resp = append(resp, ip)
				} else if m != nil && m.Allocation != nil && *m.Allocation.Name != ip.Name {
					ip.Description = "hostname mismatch"
					resp = append(resp, ip)
				}
			}
		}
	}
	return output.New().Print(resp)
}
