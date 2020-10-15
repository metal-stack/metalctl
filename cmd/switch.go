package cmd

import (
	"fmt"
	"log"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	switchCmd = &cobra.Command{
		Use:   "switch",
		Short: "manage switches",
	}

	switchListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all switches",
		RunE: func(cmd *cobra.Command, args []string) error {
			return switchList(driver)
		},
	}

	switchDetailCmd = &cobra.Command{
		Use:   "detail",
		Short: "switch details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return switchDetail(driver)
		},
	}

	switchUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "update a switch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return switchUpdate(driver)
		},
		PreRun: bindPFlags,
	}

	switchEditCmd = &cobra.Command{
		Use:   "edit <switchID>",
		Short: "edit a switch",
		RunE: func(cmd *cobra.Command, args []string) error {
			return switchEdit(driver, args)
		},
		PreRun: bindPFlags,
	}

	switchReplaceCmd = &cobra.Command{
		Use:   "replace <switchID>",
		Short: "puts a switch in replace mode in preparation for physical replacement",
		RunE: func(cmd *cobra.Command, args []string) error {
			return switchReplace(driver, args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	switchCmd.AddCommand(switchListCmd)
	switchCmd.AddCommand(switchUpdateCmd)
	switchCmd.AddCommand(switchEditCmd)
	switchCmd.AddCommand(switchDetailCmd)
	switchCmd.AddCommand(switchReplaceCmd)

	switchUpdateCmd.MarkFlagRequired("file")
	switchDetailCmd.Flags().StringP("filter", "F", "", "filter for site, rack, ID")
	viper.BindPFlags(switchDetailCmd.Flags())
}

func switchList(driver *metalgo.Driver) error {
	resp, err := driver.SwitchList()
	if err != nil {
		return formatSwaggerError(err)
	}
	return printer.Print(resp.Switch)
}

func switchDetail(driver *metalgo.Driver) error {
	resp, err := driver.SwitchList()
	if err != nil {
		return formatSwaggerError(err)
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
	return detailer.Detail(result)
}

func switchUpdate(driver *metalgo.Driver) error {
	surs, err := readSwitchUpdateRequests(viper.GetString("file"))
	if err != nil {
		return err
	}
	resp, err := driver.SwitchUpdate(surs[0])
	if err != nil {
		return formatSwaggerError(err)
	}
	return detailer.Detail(resp.Switch)
}

func switchEdit(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no switch ID given")
	}
	switchID := args[0]

	getFunc := func(id string) ([]byte, error) {
		resp, err := driver.SwitchGet(switchID)
		if err != nil {
			return nil, formatSwaggerError(err)
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
		uresp, err := driver.SwitchUpdate(items[0])
		if err != nil {
			return formatSwaggerError(err)
		}
		return detailer.Detail(uresp.Switch)
	}

	return edit(switchID, getFunc, updateFunc)
}

func switchReplace(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no switch ID given")
	}
	switchID := args[0]

	resp, err := driver.SwitchGet(switchID)
	if err != nil {
		return formatSwaggerError(err)
	}
	s := resp.Switch
	sur := metalgo.SwitchUpdateRequest{
		ID:          *s.ID,
		Name:        s.Name,
		Description: s.Description,
		RackID:      *s.RackID,
		Mode:        "replace",
	}
	uresp, err := driver.SwitchUpdate(sur)
	if err != nil {
		return formatSwaggerError(err)
	}
	return detailer.Detail(uresp.Switch)
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
