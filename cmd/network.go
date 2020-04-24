package cmd

import (
	"fmt"
	"net/http"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	networkmodel "github.com/metal-stack/metal-go/api/client/network"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/tag"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	networkCmd = &cobra.Command{
		Use:   "network",
		Short: "manage networks",
		Long:  "networks for metal.",
	}

	networkListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return networkList(driver)
		},
		PreRun: bindPFlags,
	}
	networkCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return networkCreate(driver)
		},
		PreRun: bindPFlags,
	}
	networkAllocateCmd = &cobra.Command{
		Use:   "allocate",
		Short: "allocate a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return networkAllocate(driver)
		},
		PreRun: bindPFlags,
	}
	networkFreeCmd = &cobra.Command{
		Use:   "free <networkid>",
		Short: "free a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return networkFree(driver, args)
		},
		PreRun: bindPFlags,
	}
	networkDeleteCmd = &cobra.Command{
		Use:     "delete <networkID>",
		Short:   "delete a network",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return networkDelete(driver, args)
		},
		PreRun: bindPFlags,
	}
	networkPrefixCmd = &cobra.Command{
		Use:   "prefix",
		Short: "prefix management of a network",
	}

	networkPrefixAddCmd = &cobra.Command{
		Use:   "add <networkid>",
		Short: "add a prefix to a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return networkPrefixAdd(driver, args)
		},
		PreRun: bindPFlags,
	}
	networkPrefixRemoveCmd = &cobra.Command{
		Use:   "remove <networkid>",
		Short: "remove a prefix from a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return networkPrefixRemove(driver, args)
		},
		PreRun: bindPFlags,
	}

	networkIPCmd = &cobra.Command{
		Use:   "ip",
		Short: "manage IPs",
	}

	networkIPListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "manage IPs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipList(driver)
		},
		PreRun: bindPFlags,
	}

	networkIPAllocateCmd = &cobra.Command{
		Use:   "allocate",
		Short: "allocate an IP, if non given the next free is allocated, otherwise the given IP is checked for availability.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipAllocate(driver, args)
		},
		PreRun: bindPFlags,
	}
	networkIPFreeCmd = &cobra.Command{
		Use:   "free <IP>",
		Short: "free an IP",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipFree(driver, args)
		},
		PreRun: bindPFlags,
	}
	networkIPApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update an IP",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipApply(driver)
		},
		PreRun: bindPFlags,
	}
	networkIPEditCmd = &cobra.Command{
		Use:   "edit <IP>",
		Short: "edit a ip",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipEdit(driver, args)
		},
		PreRun: bindPFlags,
	}

	networkApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a network",
		RunE: func(cmd *cobra.Command, args []string) error {
			return networkApply(driver)
		},
		PreRun: bindPFlags,
	}

	networkIPIssuesCmd = &cobra.Command{
		Use:   "issues",
		Short: `display ips which are in a potential bad state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ipIssues(driver)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	networkCreateCmd.Flags().StringP("id", "", "", "id of the network to create. [optional]")
	networkCreateCmd.Flags().StringP("description", "d", "", "description of the network to create. [optional]")
	networkCreateCmd.Flags().StringP("name", "n", "", "name of the network to create. [optional]")
	networkCreateCmd.Flags().StringP("partition", "p", "", "partition where this network should exist.")
	networkCreateCmd.Flags().StringSlice("prefixes", []string{}, "prefixes in this network.")
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
	networkAllocateCmd.MarkFlagRequired("name")
	networkAllocateCmd.MarkFlagRequired("project")
	networkAllocateCmd.MarkFlagRequired("partition")

	networkIPAllocateCmd.Flags().StringP("description", "d", "", "description of the IP to allocate. [optional]")
	networkIPAllocateCmd.Flags().StringP("name", "n", "", "name of the IP to allocate. [optional]")
	networkIPAllocateCmd.Flags().StringP("type", "", metalgo.IPTypeEphemeral, "type of the IP to allocate: "+metalgo.IPTypeEphemeral+"|"+metalgo.IPTypeStatic+" [optional]")
	networkIPAllocateCmd.Flags().StringP("network", "", "", "network from where the IP should be allocated.")
	networkIPAllocateCmd.Flags().StringP("project", "", "", "project for which the IP should be allocated.")
	networkIPAllocateCmd.Flags().StringSliceP("tags", "", nil, "tags to attach to the IP.")

	networkIPApplyCmd.MarkFlagRequired("file")

	networkIPListCmd.Flags().StringP("ipaddress", "", "", "ipaddress to filter [optional]")
	networkIPListCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	networkIPListCmd.Flags().StringP("prefix", "", "", "prefx to filter [optional]")
	networkIPListCmd.Flags().StringP("machineid", "", "", "machineid to filter [optional]")
	networkIPListCmd.Flags().StringP("type", "", "", "type to filter [optional]")
	networkIPListCmd.Flags().StringP("network", "", "", "network to filter [optional]")
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

	networkApplyCmd.MarkFlagRequired("file")

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
	networkCmd.AddCommand(networkAllocateCmd)
	networkCmd.AddCommand(networkFreeCmd)
	networkCmd.AddCommand(networkPrefixCmd)
	networkCmd.AddCommand(networkApplyCmd)
	networkCmd.AddCommand(networkDeleteCmd)
}

func networkList(driver *metalgo.Driver) error {
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
		resp, err = driver.NetworkFind(nfr)
	} else {
		resp, err = driver.NetworkList()
	}
	if err != nil {
		return fmt.Errorf("network list error:%v", err)
	}
	return printer.Print(resp.Networks)
}

func networkAllocate(driver *metalgo.Driver) error {
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
		ncr = metalgo.NetworkAllocateRequest{
			Description: viper.GetString("description"),
			Name:        viper.GetString("name"),
			PartitionID: viper.GetString("partition"),
			ProjectID:   viper.GetString("project"),
			Labels:      labelsFromTags(viper.GetStringSlice("labels")),
		}
	}
	resp, err := driver.NetworkAllocate(&ncr)
	if err != nil {
		return fmt.Errorf("network allocate error:%v", err)
	}
	return detailer.Detail(resp.Network)
}

func networkFree(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no network given")
	}
	nw := args[0]
	resp, err := driver.NetworkFree(nw)
	if err != nil {
		return fmt.Errorf("network allocate error:%v", err)
	}
	return detailer.Detail(resp.Network)
}

func networkDelete(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no network given")
	}
	nw := args[0]
	resp, err := driver.NetworkDelete(nw)
	if err != nil {
		return fmt.Errorf("network delete error:%v", err)
	}
	return detailer.Detail(resp.Network)
}

func networkCreate(driver *metalgo.Driver) error {
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
		}
		id := viper.GetString("id")
		if len(id) > 0 {
			ncr.ID = &id
		}
	}
	resp, err := driver.NetworkCreate(&ncr)
	if err != nil {
		return fmt.Errorf("network create error:%v", err)
	}
	return detailer.Detail(resp.Network)
}

// TODO: General apply method would be useful as these are quite a lot of lines and it's getting erroneous
func networkApply(driver *metalgo.Driver) error {
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
		if nar.ID == nil {
			resp, err := driver.NetworkCreate(&nar)
			if err != nil {
				return fmt.Errorf("network update error:%v", err)
			}
			response = append(response, resp.Network)
			continue
		}

		networkID := *nar.ID
		p, err := driver.NetworkGet(networkID)
		if err != nil {
			if e, ok := err.(*networkmodel.FindNetworkDefault); ok {
				if e.Code() != http.StatusNotFound {
					return fmt.Errorf("network get error:%v", err)
				}
			}
		}
		if p.Network == nil {
			resp, err := driver.NetworkCreate(&nar)
			if err != nil {
				return fmt.Errorf("network update error:%v", err)
			}
			response = append(response, resp.Network)
			continue
		}
		if p.Network.ID != nil {
			resp, err := driver.NetworkUpdate(&nar)
			if err != nil {
				return fmt.Errorf("network create error:%v", err)
			}
			response = append(response, resp.Network)
			continue
		}
	}
	return detailer.Detail(response)
}

func networkPrefixAdd(driver *metalgo.Driver, args []string) error {
	networkID, err := getNetworkID(args)
	if err != nil {
		return err
	}

	nur := &metalgo.NetworkUpdateRequest{
		Networkid: networkID,
		Prefix:    viper.GetString("prefix"),
	}
	resp, err := driver.NetworkAddPrefix(nur)
	if err != nil {
		return fmt.Errorf("prefix add error:%v", err)
	}
	return detailer.Detail(resp.Network)
}

func networkPrefixRemove(driver *metalgo.Driver, args []string) error {
	networkID, err := getNetworkID(args)
	if err != nil {
		return err
	}

	nur := &metalgo.NetworkUpdateRequest{
		Networkid: networkID,
		Prefix:    viper.GetString("prefix"),
	}
	resp, err := driver.NetworkRemovePrefix(nur)
	if err != nil {
		return fmt.Errorf("prefix remove error:%v", err)
	}
	return detailer.Detail(resp.Network)
}

func ipList(driver *metalgo.Driver) error {
	var resp *metalgo.IPListResponse
	var err error
	if atLeastOneViperStringFlagGiven("ipaddress", "project", "prefix", "machineid", "network", "type", "tags") {
		ifr := &metalgo.IPFindRequest{
			IPAddress:        viperString("ipaddress"),
			ProjectID:        viperString("project"),
			ParentPrefixCidr: viperString("prefix"),
			NetworkID:        viperString("network"),
			MachineID:        viperString("machineid"),
			Type:             viperString("type"),
			Tags:             viperStringSlice("tags"),
		}
		resp, err = driver.IPFind(ifr)
	} else {
		resp, err = driver.IPList()
	}
	if err != nil {
		return fmt.Errorf("IP list error:%v", err)
	}
	return printer.Print(resp.IPs)
}

func ipApply(driver *metalgo.Driver) error {
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
		if iar.IPAddress == "" {
			// acquire
			resp, err := driver.IPAllocate(&iar)
			if err != nil {
				return fmt.Errorf("IP allocate error:%v", err)
			}
			response = append(response, resp.IP)
			continue
		}
		i, err := driver.IPGet(iar.IPAddress)
		if err != nil {
			// specific acquire
			resp, err := driver.IPAllocate(&iar)
			if err != nil {
				return fmt.Errorf("IP allocate specific ip error:%v", err)
			}
			response = append(response, resp.IP)
			continue
		}
		// update
		iur := metalgo.IPUpdateRequest{
			IPAddress:   *i.IP.Ipaddress,
			Name:        iar.Name,
			Description: iar.Description,
			Type:        iar.Type,
			Tags:        iar.Tags,
		}
		resp, err := driver.IPUpdate(&iur)
		if err != nil {
			return fmt.Errorf("IP update error:%v", err)
		}
		response = append(response, resp.IP)
	}

	return detailer.Detail(response)
}

func ipEdit(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no IP given")
	}
	ip := args[0]

	getFunc := func(ip string) ([]byte, error) {
		resp, err := driver.IPGet(ip)
		if err != nil {
			return nil, fmt.Errorf("ip get error:%v", err)
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
		uresp, err := driver.IPUpdate(&iurs[0])
		if err != nil {
			return fmt.Errorf("size update error:%v", err)
		}
		return detailer.Detail(uresp.IP)
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

func ipAllocate(driver *metalgo.Driver, args []string) error {
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
	resp, err := driver.IPAllocate(iar)
	if err != nil {
		return fmt.Errorf("IP allocate error:%v", err)
	}
	return detailer.Detail(resp.IP)
}

func ipFree(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no IP given")
	}
	ip := args[0]
	resp, err := driver.IPFree(ip)
	if err != nil {
		return fmt.Errorf("IP free error:%v", err)
	}
	return detailer.Detail(resp.IP)
}

func getNetworkID(args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("no network ID given")
	}

	networkID := args[0]
	_, err := driver.NetworkGet(networkID)
	if err != nil {
		return "", fmt.Errorf("network with ID:%s does not exist", networkID)
	}
	return networkID, nil
}

func ipIssues(driver *metalgo.Driver) error {
	ml, err := driver.MachineList()
	if err != nil {
		return fmt.Errorf("machine list error:%v", err)
	}
	machines := make(map[string]*models.V1MachineResponse)
	for _, m := range ml.Machines {
		machines[*m.ID] = m
	}

	var resp []*models.V1IPResponse

	iplist, err := driver.IPList()
	if err != nil {
		return fmt.Errorf("IP list error:%v", err)
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
					ip.Description = fmt.Sprint("bound to unallocated machine")
					resp = append(resp, ip)
				} else if m != nil && m.Allocation != nil && *m.Allocation.Name != ip.Name {
					ip.Description = fmt.Sprint("hostname mismatch")
					resp = append(resp, ip)
				}
			}
		}
	}
	return printer.Print(resp)
}
