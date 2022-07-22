package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/metal-stack/metal-go/api/client/ip"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/tag"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: just having a single ip create endpoint in the metal-api would simplify things a lot... maybe just deprecate the specific one and add ip address field to regular allocate request
type ipAllocateRequest struct {
	SpecificIP string `json:"ipaddress"`
	*models.V1IPAllocateRequest
}

type ipCmd struct {
	*config
	*genericcli.GenericCLI[*ipAllocateRequest, *models.V1IPUpdateRequest, *models.V1IPResponse]
}

func newIPCmd(c *config) *cobra.Command {
	w := ipCmd{
		config:     c,
		GenericCLI: genericcli.NewGenericCLI[*ipAllocateRequest, *models.V1IPUpdateRequest, *models.V1IPResponse](ipCRUD{config: c}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*ipAllocateRequest, *models.V1IPUpdateRequest, *models.V1IPResponse]{
		gcli:     w.GenericCLI,
		singular: "ip",
		plural:   "ips",

		createRequestFromCLI: w.createRequestFromCLI,

		availableSortKeys: sorters.IPSortKeys(),
		validArgsFunc:     c.comp.IpListCompletion,
	})

	issuesCmd := &cobra.Command{
		Use:   "issues",
		Short: `display ips which are in a potential bad state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.ipIssues()
		},
		PreRun: bindPFlags,
	}

	cmds.createCmd.Flags().StringP("ipaddress", "", "", "a specific ip address to allocate. [optional]")
	cmds.createCmd.Flags().StringP("description", "d", "", "description of the IP to allocate. [optional]")
	cmds.createCmd.Flags().StringP("name", "n", "", "name of the IP to allocate. [optional]")
	cmds.createCmd.Flags().StringP("type", "", models.V1IPAllocateRequestTypeEphemeral, "type of the IP to allocate: "+models.V1IPAllocateRequestTypeEphemeral+"|"+models.V1IPAllocateRequestTypeStatic+" [optional]")
	cmds.createCmd.Flags().StringP("network", "", "", "network from where the IP should be allocated.")
	cmds.createCmd.Flags().StringP("project", "", "", "project for which the IP should be allocated.")
	cmds.createCmd.Flags().StringSliceP("tags", "", nil, "tags to attach to the IP.")
	must(cmds.createCmd.RegisterFlagCompletionFunc("network", c.comp.NetworkListCompletion))
	must(cmds.createCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(cmds.createCmd.RegisterFlagCompletionFunc("type", cobra.FixedCompletions([]string{models.V1IPAllocateRequestTypeEphemeral, models.V1IPAllocateRequestTypeStatic}, cobra.ShellCompDirectiveNoFileComp)))

	cmds.listCmd.Flags().StringP("ipaddress", "", "", "ipaddress to filter [optional]")
	cmds.listCmd.Flags().StringP("project", "", "", "project to filter [optional]")
	cmds.listCmd.Flags().StringP("prefix", "", "", "prefix to filter [optional]")
	cmds.listCmd.Flags().StringP("machineid", "", "", "machineid to filter [optional]")
	cmds.listCmd.Flags().StringP("type", "", "", "type to filter [optional]")
	cmds.listCmd.Flags().StringP("network", "", "", "network to filter [optional]")
	cmds.listCmd.Flags().StringP("name", "", "", "name to filter [optional]")
	cmds.listCmd.Flags().StringSliceP("tags", "", nil, "tags to filter [optional]")
	must(cmds.listCmd.RegisterFlagCompletionFunc("ipaddress", c.comp.IpListCompletion))
	must(cmds.listCmd.RegisterFlagCompletionFunc("network", c.comp.NetworkListCompletion))
	must(cmds.listCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	must(cmds.listCmd.RegisterFlagCompletionFunc("type", cobra.FixedCompletions([]string{models.V1IPAllocateRequestTypeEphemeral, models.V1IPAllocateRequestTypeStatic}, cobra.ShellCompDirectiveNoFileComp)))
	must(cmds.listCmd.RegisterFlagCompletionFunc("machineid", c.comp.MachineListCompletion))

	return cmds.buildRootCmd(issuesCmd)
}

type ipCRUD struct {
	*config
}

func (c ipCRUD) Get(id string) (*models.V1IPResponse, error) {
	resp, err := c.client.IP().FindIP(ip.NewFindIPParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCRUD) List() ([]*models.V1IPResponse, error) {
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

	err = sorters.IPSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCRUD) Delete(id string) (*models.V1IPResponse, error) {
	resp, err := c.client.IP().FreeIP(ip.NewFreeIPParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCRUD) Create(rq *ipAllocateRequest) (*models.V1IPResponse, error) {
	if rq.SpecificIP == "" {
		resp, err := c.client.IP().AllocateIP(ip.NewAllocateIPParams().WithBody(rq.V1IPAllocateRequest), nil)
		if err != nil {
			var r *ip.AllocateIPDefault // FIXME: API should define to return conflict
			if errors.As(err, &r) && r.Code() == http.StatusConflict {
				return nil, genericcli.AlreadyExistsError()
			}
			return nil, err
		}

		return resp.Payload, nil
	}

	resp, err := c.client.IP().AllocateSpecificIP(ip.NewAllocateSpecificIPParams().WithIP(rq.SpecificIP).WithBody(rq.V1IPAllocateRequest), nil)
	if err != nil {
		var r *ip.AllocateIPDefault // FIXME: API should define to return conflict
		if errors.As(err, &r) && r.Code() == http.StatusConflict {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c ipCRUD) Update(rq *models.V1IPUpdateRequest) (*models.V1IPResponse, error) {
	resp, err := c.client.IP().UpdateIP(ip.NewUpdateIPParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
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
				} else if m != nil && m.Allocation != nil && *m.Allocation.Name != ip.Name {
					ip.Description = "hostname mismatch"
					ips = append(ips, ip)
				}
			}
		}
	}

	return newPrinterFromCLI().Print(ips)
}
