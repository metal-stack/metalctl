package cmd

import (
	"fmt"
	"log"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func newSwitchCmd(c *config) *cobra.Command {
	switchCmd := &cobra.Command{
		Use:   "switch",
		Short: "manage switches",
	}

	switchListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all switches",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.switchList()
		},
	}

	switchDetailCmd := &cobra.Command{
		Use:   "detail",
		Short: "switch details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.switchDetail()
		},
	}

	switchUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a switch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.switchUpdate()
		},
		PreRun: bindPFlags,
	}

	switchEditCmd := &cobra.Command{
		Use:   "edit <switchID>",
		Short: "edit a switch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.switchEdit(args)
		},
		PreRun: bindPFlags,
	}

	switchReplaceCmd := &cobra.Command{
		Use:   "replace <switchID>",
		Short: "puts a switch in replace mode in preparation for physical replacement",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.switchReplace(args)
		},
		PreRun: bindPFlags,
	}

	switchCmd.AddCommand(switchListCmd)
	switchCmd.AddCommand(switchUpdateCmd)
	switchCmd.AddCommand(switchEditCmd)
	switchCmd.AddCommand(switchDetailCmd)
	switchCmd.AddCommand(switchReplaceCmd)

	switchUpdateCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.`)
	must(switchUpdateCmd.MarkFlagRequired("file"))

	switchDetailCmd.Flags().StringP("filter", "F", "", "filter for site, rack, ID")
	must(viper.BindPFlags(switchDetailCmd.Flags()))

	return switchCmd
}

func (c *config) switchList() error {
	resp, err := c.driver.SwitchList()
	if err != nil {
		return err
	}
	return output.New().Print(resp.Switch)
}

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

func (c *config) switchUpdate() error {
	surs, err := readSwitchUpdateRequests(viper.GetString("file"))
	if err != nil {
		return err
	}
	resp, err := c.driver.SwitchUpdate(surs[0])
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Switch)
}

func (c *config) switchEdit(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no switch ID given")
	}
	switchID := args[0]

	getFunc := func(id string) ([]byte, error) {
		resp, err := c.driver.SwitchGet(switchID)
		if err != nil {
			return nil, err
		}
		content, err := yaml.Marshal(resp.Switch)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		items, err := readSwitchUpdateRequests(filename)
		if err != nil {
			return err
		}
		uresp, err := c.driver.SwitchUpdate(items[0])
		if err != nil {
			return err
		}
		return output.NewDetailer().Detail(uresp.Switch)
	}

	return edit(switchID, getFunc, updateFunc)
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

func readSwitchUpdateRequests(filename string) ([]metalgo.SwitchUpdateRequest, error) {
	var items []metalgo.SwitchUpdateRequest
	var item metalgo.SwitchUpdateRequest
	err := readFrom(filename, &item, func(data interface{}) {
		doc := data.(*metalgo.SwitchUpdateRequest)
		items = append(items, *doc)
	})
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, fmt.Errorf("switch update error more or less than one switch given:%d", len(items))
	}
	return items, nil
}
