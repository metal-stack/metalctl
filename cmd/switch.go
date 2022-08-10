package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/metal-stack/metal-go/api/client/switch_operations"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type switchCmd struct {
	*config
	*genericcli.GenericCLI[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse]
}

func newSwitchCmd(c *config) *cobra.Command {
	w := switchCmd{
		config:     c,
		GenericCLI: genericcli.NewGenericCLI[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse](switchCRUD{config: c}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse]{
		gcli:        w.GenericCLI,
		singular:    "switch",
		plural:      "switches",
		description: "switch are the leaf switches in the data center that are controlled by metal-stack.",

		availableSortKeys: sorters.SwitchSortKeys(),
	})

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
		PreRun: bindPFlags,
	}

	switchDetailCmd.Flags().StringP("filter", "F", "", "filter for site, rack, ID")
	must(viper.BindPFlags(switchDetailCmd.Flags()))

	cmds.rootCmd.AddCommand(
		cmds.listCmd,
		cmds.describeCmd,
		cmds.updateCmd,
		cmds.deleteCmd,
		cmds.editCmd,
	)

	cmds.rootCmd.AddCommand(switchDetailCmd)
	cmds.rootCmd.AddCommand(switchReplaceCmd)

	return cmds.rootCmd
}

type switchCRUD struct {
	*config
}

func (c switchCRUD) Get(id string) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().FindSwitch(switch_operations.NewFindSwitchParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCRUD) List() ([]*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, err
	}

	err = sorters.SwitchSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCRUD) Delete(id string) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().DeleteSwitch(switch_operations.NewDeleteSwitchParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCRUD) Create(rq any) (*models.V1SwitchResponse, error) {
	return nil, fmt.Errorf("switch entity does not support create operation")
}

func (c switchCRUD) Update(rq *models.V1SwitchUpdateRequest) (*models.V1SwitchResponse, error) {
	resp, err := c.client.SwitchOperations().UpdateSwitch(switch_operations.NewUpdateSwitchParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
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

	return newPrinterFromCLI().Print(result)
}

func (c *switchCmd) switchReplace(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.Interface().Get(id)
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

	return defaultToYAMLPrinter().Print(uresp)
}
