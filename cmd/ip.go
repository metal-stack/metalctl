package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/metal-stack/metal-go/api/client/ip"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/tag"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: just having a single ip create endpoint in the metal-api would simplify things a lot... maybe just deprecate the specific one and add ip address field to regular allocate request
type ipAllocateRequest struct {
	SpecificIP                  string `json:"ipaddress" yaml:"ipaddress"`
	*models.V1IPAllocateRequest `yaml:",inline"`
}

type ipCmd struct {
	*config
}

func newIPCmd(c *config) *cobra.Command {
	w := ipCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*ipAllocateRequest, *models.V1IPUpdateRequest, *models.V1IPResponse]{
		BinaryName:           binaryName,
		GenericCLI:           genericcli.NewGenericCLI[*ipAllocateRequest, *models.V1IPUpdateRequest, *models.V1IPResponse](w).WithFS(c.fs),
		Singular:             "ip",
		Plural:               "ips",
		Description:          "an ip address can be attached to a machine or firewall such that network traffic can be routed to these servers.",
		DescribePrinter:      func() printers.Printer { return c.describePrinter },
		ListPrinter:          func() printers.Printer { return c.listPrinter },
		CreateRequestFromCLI: w.createRequestFromCLI,
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("ipaddress", "", "", "a specific ip address to allocate. [optional]")
			cmd.Flags().StringP("description", "d", "", "description of the IP to allocate. [optional]")
			cmd.Flags().StringP("name", "n", "", "name of the IP to allocate. [optional]")
			cmd.Flags().StringP("type", "", models.V1IPAllocateRequestTypeEphemeral, "type of the IP to allocate: "+models.V1IPAllocateRequestTypeEphemeral+"|"+models.V1IPAllocateRequestTypeStatic+" [optional]")
			cmd.Flags().StringP("network", "", "", "network from where the IP should be allocated.")
			cmd.Flags().StringP("project", "", "", "project for which the IP should be allocated.")
			cmd.Flags().StringSliceP("tags", "", nil, "tags to attach to the IP.")
			must(cmd.RegisterFlagCompletionFunc("network", c.comp.NetworkListCompletion))
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			must(cmd.RegisterFlagCompletionFunc("type", cobra.FixedCompletions([]string{models.V1IPAllocateRequestTypeEphemeral, models.V1IPAllocateRequestTypeStatic}, cobra.ShellCompDirectiveNoFileComp)))
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("ipaddress", "", "", "ipaddress to filter [optional]")
			cmd.Flags().StringP("project", "", "", "project to filter [optional]")
			cmd.Flags().StringP("prefix", "", "", "prefix to filter [optional]")
			cmd.Flags().StringP("machineid", "", "", "machineid to filter [optional]")
			cmd.Flags().StringP("type", "", "", "type to filter [optional]")
			cmd.Flags().StringP("network", "", "", "network to filter [optional]")
			cmd.Flags().StringP("name", "", "", "name to filter [optional]")
			cmd.Flags().StringSliceP("tags", "", nil, "tags to filter [optional]")
			must(cmd.RegisterFlagCompletionFunc("ipaddress", c.comp.IpListCompletion))
			must(cmd.RegisterFlagCompletionFunc("network", c.comp.NetworkListCompletion))
			must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			must(cmd.RegisterFlagCompletionFunc("type", cobra.FixedCompletions([]string{models.V1IPAllocateRequestTypeEphemeral, models.V1IPAllocateRequestTypeStatic}, cobra.ShellCompDirectiveNoFileComp)))
			must(cmd.RegisterFlagCompletionFunc("machineid", c.comp.MachineListCompletion))
		},
		DeleteCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Aliases = append(cmd.Aliases, "free")
		},
		Sorter:      sorters.IPSorter(),
		ValidArgsFn: c.comp.IpListCompletion,
	}

	issuesCmd := &cobra.Command{
		Use:   "issues",
		Short: `display ips which are in a potential bad state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.ipIssues()
		},
	}

	return genericcli.NewCmds(cmdsConfig, issuesCmd)
}

func (c ipCmd) Get(id string) (*models.V1IPResponse, error) {
	resp, err := c.client.IP().FindIP(ip.NewFindIPParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCmd) List() ([]*models.V1IPResponse, error) {
	resp, err := c.client.IP().FindIPs(ip.NewFindIPsParams().WithBody(&models.V1IPFindRequest{
		Ipaddress:     viper.GetString("ipaddress"),
		Name:          viper.GetString("name"),
		Type:          viper.GetString("type"),
		Projectid:     viper.GetString("project"),
		Networkid:     viper.GetString("network"),
		Machineid:     viper.GetString("machineid"),
		Networkprefix: viper.GetString("prefix"),
		Tags:          viper.GetStringSlice("tags"),
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCmd) Delete(id string) (*models.V1IPResponse, error) {
	resp, err := c.client.IP().FreeIP(ip.NewFreeIPParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCmd) Create(rq *ipAllocateRequest) (*models.V1IPResponse, error) {
	if rq.SpecificIP == "" {
		resp, err := c.client.IP().AllocateIP(ip.NewAllocateIPParams().WithBody(rq.V1IPAllocateRequest), nil)
		if err != nil {
			var r *ip.AllocateIPConflict
			if errors.As(err, &r) {
				return nil, genericcli.AlreadyExistsError()
			}
			return nil, err
		}

		return resp.Payload, nil
	}

	resp, err := c.client.IP().AllocateSpecificIP(ip.NewAllocateSpecificIPParams().WithIP(rq.SpecificIP).WithBody(rq.V1IPAllocateRequest), nil)
	if err != nil {
		var r *ip.AllocateSpecificIPConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCmd) Update(rq *models.V1IPUpdateRequest) (*models.V1IPResponse, error) {
	resp, err := c.client.IP().UpdateIP(ip.NewUpdateIPParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCmd) Convert(r *models.V1IPResponse) (string, *ipAllocateRequest, *models.V1IPUpdateRequest, error) {
	if r.Ipaddress == nil {
		return "", nil, nil, fmt.Errorf("ipaddress is nil")
	}
	return *r.Ipaddress, ipResponseToCreate(r), ipResponseToUpdate(r), nil
}

func ipResponseToCreate(r *models.V1IPResponse) *ipAllocateRequest {
	var ip string
	if r.Ipaddress != nil {
		ip = *r.Ipaddress
	}
	return &ipAllocateRequest{
		SpecificIP: ip,
		V1IPAllocateRequest: &models.V1IPAllocateRequest{
			Description: r.Description,
			Name:        r.Name,
			Networkid:   r.Networkid,
			Projectid:   r.Projectid,
			Tags:        r.Tags,
			Type:        r.Type,
		},
	}
}

func ipResponseToUpdate(r *models.V1IPResponse) *models.V1IPUpdateRequest {
	return &models.V1IPUpdateRequest{
		Description: r.Description,
		Ipaddress:   r.Ipaddress,
		Name:        r.Name,
		Tags:        r.Tags,
		Type:        r.Type,
	}
}

func (c *ipCmd) createRequestFromCLI() (*ipAllocateRequest, error) {
	return &ipAllocateRequest{
		SpecificIP: viper.GetString("ipaddress"),
		V1IPAllocateRequest: &models.V1IPAllocateRequest{
			Description: viper.GetString("description"),
			Name:        viper.GetString("name"),
			Networkid:   pointer.Pointer(viper.GetString("network")),
			Projectid:   pointer.Pointer(viper.GetString("project")),
			Type:        pointer.Pointer(viper.GetString("type")),
			Tags:        viper.GetStringSlice("tags"),
		},
	}, nil
}

// non-generic command handling

func (c *ipCmd) ipIssues() error {
	ml, err := c.client.Machine().ListMachines(machine.NewListMachinesParams(), nil)
	if err != nil {
		return fmt.Errorf("machine list error:%w", err)
	}

	machines := make(map[string]*models.V1MachineResponse)
	for _, m := range ml.Payload {
		machines[*m.ID] = m
	}

	var ips []*models.V1IPResponse

	resp, err := c.List()
	if err != nil {
		return err
	}

	for _, ip := range resp {
		if *ip.Type == models.V1IPAllocateRequestTypeStatic {
			continue
		}
		if ip.Description == "autoassigned" && len(ip.Tags) == 0 {
			ip.Description = fmt.Sprintf("%s, but no tags", ip.Description)
			ips = append(ips, ip)
		}
		if strings.HasPrefix(ip.Name, "metallb-") && len(ip.Tags) == 0 {
			ip.Description = fmt.Sprintf("metallb ip without tags %s", ip.Description)
			ips = append(ips, ip)
		}

		for _, t := range ip.Tags {
			if strings.HasPrefix(t, tag.MachineID+"=") {
				parts := strings.Split(t, "=")
				m := machines[parts[1]]
				if m == nil || *m.Liveliness != "Alive" || m.Allocation == nil || *m.Events.Log[0].Event != "Phoned Home" {
					ip.Description = "bound to unallocated machine"
					ips = append(ips, ip)
				} else if m.Allocation != nil && *m.Allocation.Name != ip.Name {
					ip.Description = "hostname mismatch"
					ips = append(ips, ip)
				}
			}
		}
	}

	return c.listPrinter.Print(ips)
}
