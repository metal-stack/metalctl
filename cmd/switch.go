package cmd

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/metal-stack/metal-go/api/client/switch_operations"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type switchCmd struct {
	*config
}

func newSwitchCmd(c *config) *cobra.Command {
	w := switchCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse]{
		BinaryName: binaryName,
		GenericCLI: genericcli.NewGenericCLI[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse](w).WithFS(c.fs),
		OnlyCmds: genericcli.OnlyCmds(
			genericcli.ListCmd,
			genericcli.DescribeCmd,
			genericcli.UpdateCmd,
			genericcli.DeleteCmd,
			genericcli.EditCmd,
		),
		Singular:        "switch",
		Plural:          "switches",
		Description:     "switch are the leaf switches in the data center that are controlled by metal-stack.",
		Sorter:          sorters.SwitchSorter(),
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
	}

	switchDetailCmd := &cobra.Command{
		Use:   "detail",
		Short: "switch details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.switchDetail()
		},
	}
	switchReplaceCmd := &cobra.Command{
		Use:   "replace <switchID>",
		Short: "put a leaf switch into replace mode in preparation for physical replacement. For a description of the steps involved see the long help.",
		Long: `Put a leaf switch into replace mode in preparation for physical replacement

Operational steps to replace a switch:

- Put the switch that needs to be replaced in replace mode with this command
- Replace the switch MAC address in the metal-stack deployment configuration
- Make sure that interfaces on the new switch do not get connected to the PXE-bridge immediately by setting the interfaces list of the respective leaf switch to [] in the metal-stack deployment configuration
- Deploy the management servers so that the dhcp servers will serve the right address and DHCP options to the new switch
- Replace the switch physically. Be careful to ensure that the cabling mirrors the remaining leaf exactly because the new switch information will be cloned from the remaining switch! Also make sure to have console access to the switch so you can start and monitor the install process
- If the switch is not in onie install mode but already has an operating system installed, put it into install mode with "sudo onie-select -i -f -v" and reboot it. Now the switch should be provisioned with a management IP from a management server, install itself with the right software image and receive license and ssh keys through ZTP. You can check whether that process has completed successfully with the command "sudo ztp -s". The ZTP state should be disabled and the result should be success.
- Deploy the switch plane and metal-core through metal-stack deployment CI job
- The switch will now register with its metal-api, and the metal-core service will receive the cloned interface and routing information. You can verify successful switch replacement by checking the interface and BGP configuration, and checking the switch status with "metalctl switch ls -o wide"; it should now be operational again`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.switchReplace(args)
		},
	}

	switchDetailCmd.Flags().StringP("filter", "F", "", "filter for site, rack, ID")

	return genericcli.NewCmds(cmdsConfig, switchDetailCmd, switchReplaceCmd)
}

func (c switchCmd) Get(id string) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().FindSwitch(switch_operations.NewFindSwitchParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCmd) List() ([]*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, err
	}

	for _, s := range resp.Payload {
		s := s
		sort.SliceStable(s.Connections, func(i, j int) bool {
			return pointer.SafeDeref(pointer.SafeDeref((pointer.SafeDeref(s.Connections[i])).Nic).Name) < pointer.SafeDeref(pointer.SafeDeref((pointer.SafeDeref(s.Connections[j])).Nic).Name)
		})
	}

	return resp.Payload, nil
}

func (c switchCmd) Delete(id string) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().DeleteSwitch(switch_operations.NewDeleteSwitchParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCmd) Create(rq any) (*models.V1SwitchResponse, error) {
	return nil, fmt.Errorf("switch entity does not support create operation")
}

func (c switchCmd) Update(rq *models.V1SwitchUpdateRequest) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().UpdateSwitch(switch_operations.NewUpdateSwitchParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCmd) ToCreate(r *models.V1SwitchResponse) (any, error) {
	return nil, fmt.Errorf("switch entity does not support create operation")
}

func (c switchCmd) ToUpdate(r *models.V1SwitchResponse) (*models.V1SwitchUpdateRequest, error) {
	return switchResponseToUpdate(r), nil
}

func switchResponseToUpdate(r *models.V1SwitchResponse) *models.V1SwitchUpdateRequest {
	return &models.V1SwitchUpdateRequest{
		Description: r.Description,
		ID:          r.ID,
		Mode:        r.Mode,
		Name:        r.Name,
		RackID:      r.RackID,
	}
}

// non-generic command handling

func (c *switchCmd) switchDetail() error {
	resp, err := c.List()
	if err != nil {
		return err
	}

	var result []*tableprinters.SwitchDetail
	filter := viper.GetString("filter")
	for _, s := range resp {
		partitionID := ""
		if s.Partition != nil {
			partitionID = *s.Partition.ID
		}
		if strings.Contains(*s.ID, filter) ||
			strings.Contains(partitionID, filter) ||
			strings.Contains(*s.RackID, filter) {
			result = append(result, &tableprinters.SwitchDetail{V1SwitchResponse: s})
		}
	}

	if len(result) < 1 {
		log.Printf("no switch detail for filter: %s", filter)
		return nil
	}

	return c.listPrinter.Print(result)
}

func (c *switchCmd) switchReplace(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.Get(id)
	if err != nil {
		return err
	}

	uresp, err := c.Update(&models.V1SwitchUpdateRequest{
		ID:          resp.ID,
		Name:        resp.Name,
		Description: resp.Description,
		RackID:      resp.RackID,
		Mode:        "replace",
	})
	if err != nil {
		return err
	}

	return c.describePrinter.Print(uresp)
}
