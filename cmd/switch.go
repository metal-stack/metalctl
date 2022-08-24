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
		Short: "puts a switch in replace mode in preparation for physical replacement",
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
