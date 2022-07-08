package cmd

import (
	"fmt"
	"log"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/switch_operations"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type switchCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	*genericcli.GenericCLI[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse]
}

func newSwitchCmd(c *config) *cobra.Command {
	w := switchCmd{
		c:          c.client,
		driver:     c.driver,
		GenericCLI: genericcli.NewGenericCLI[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse](switchCRUD{Client: c.client}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[any, *models.V1SwitchUpdateRequest, *models.V1SwitchResponse]{
		gcli:        w.GenericCLI,
		singular:    "switch",
		plural:      "switches",
		description: "switch are the leaf switches in the data center that are controlled by metal-stack.",
	})

	switchDetailCmd := &cobra.Command{
		Use:   "detail",
		Short: "switch details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.switchDetail()
		},
	}
	switchReplaceCmd := &cobra.Command{
		Use:   "replace <switchID>",
		Short: "puts a switch in replace mode in preparation for physical replacement",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.switchReplace(args)
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
	metalgo.Client
}

func (c switchCRUD) Get(id string) (*models.V1SwitchResponse, error) {
	resp, err := c.SwitchOperations().FindSwitch(switch_operations.NewFindSwitchParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCRUD) List() ([]*models.V1SwitchResponse, error) {
	resp, err := c.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCRUD) Delete(id string) (*models.V1SwitchResponse, error) {
	resp, err := c.SwitchOperations().DeleteSwitch(switch_operations.NewDeleteSwitchParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c switchCRUD) Create(rq any) (*models.V1SwitchResponse, error) {
	return nil, fmt.Errorf("switch entity has no create operation")
}

func (c switchCRUD) Update(rq *models.V1SwitchUpdateRequest) (*models.V1SwitchResponse, error) {
	resp, err := c.SwitchOperations().UpdateSwitch(switch_operations.NewUpdateSwitchParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *config) switchDetail() error {
	resp, err := c.driver.SwitchList()
	if err != nil {
		return err
	}
	result := make([]*models.V1SwitchResponse, 0)
	filter := viper.GetString("filter")
	for _, s := range resp.Switch {
		partitionID := ""
		if s.Partition != nil {
			partitionID = *s.Partition.ID
		}
		if strings.Contains(*s.ID, filter) ||
			strings.Contains(partitionID, filter) ||
			strings.Contains(*s.RackID, filter) {
			result = append(result, s)
		}
	}

	if len(result) < 1 {
		log.Printf("no switch detail for filter: %s", filter)
		return nil
	}

	return output.NewDetailer().Detail(result)
}

func (c *config) switchReplace(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no switch ID given")
	}
	switchID := args[0]

	resp, err := c.driver.SwitchGet(switchID)
	if err != nil {
		return err
	}
	s := resp.Switch
	sur := metalgo.SwitchUpdateRequest{
		ID:          *s.ID,
		Name:        s.Name,
		Description: s.Description,
		RackID:      *s.RackID,
		Mode:        "replace",
	}
	uresp, err := c.driver.SwitchUpdate(sur)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(uresp.Switch)
}
