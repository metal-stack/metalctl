package cmd

import (
	"fmt"
	"log"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
)

func init() {
	switchCmd.AddCommand(switchListCmd)

	switchDetailCmd.Flags().StringP("filter", "F", "", "filter for site, rack, ID")
	switchCmd.AddCommand(switchDetailCmd)
	viper.BindPFlags(switchDetailCmd.Flags())
}

func switchList(driver *metalgo.Driver) error {
	resp, err := driver.SwitchList()
	if err != nil {
		return fmt.Errorf("switch list error:%v", err)
	}
	return printer.Print(resp.Switch)
}

func switchDetail(driver *metalgo.Driver) error {
	resp, err := driver.SwitchList()
	if err != nil {
		return fmt.Errorf("switch detail error:%v", err)
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
