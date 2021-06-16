package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	metalgo "github.com/metal-stack/metal-go"
	partitionmodel "github.com/metal-stack/metal-go/api/client/partition"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	partitionCmd = &cobra.Command{
		Use:   "partition",
		Short: "manage partitions",
		Long:  "a partition is a group of machines and network which is logically separated from other partitions. Machines have no direct network connections between partitions.",
	}

	partitionListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all partitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return partitionList(driver)
		},
	}
	partitionCapacityCmd = &cobra.Command{
		Use:   "capacity",
		Short: "show partition capacity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return partitionCapacity(driver)
		},
	}
	partitionDescribeCmd = &cobra.Command{
		Use:   "describe <partitionID>",
		Short: "describe a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return partitionDescribe(driver, args)
		},
	}
	partitionCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return partitionCreate(driver)
		},
		PreRun: bindPFlags,
	}
	partitionUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "update a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return partitionUpdate(driver)
		},
		PreRun: bindPFlags,
	}
	partitionApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return partitionApply(driver)
		},
		PreRun: bindPFlags,
	}
	partitionDeleteCmd = &cobra.Command{
		Use:   "delete <partitionID>",
		Short: "delete a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return partitionDelete(driver, args)
		},
		PreRun: bindPFlags,
	}
	partitionEditCmd = &cobra.Command{
		Use:   "edit <partitionID>",
		Short: "edit a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return partitionEdit(driver, args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	partitionCreateCmd.Flags().StringP("id", "", "", "ID of the partition. [required]")
	partitionCreateCmd.Flags().StringP("name", "n", "", "Name of the partition. [optional]")
	partitionCreateCmd.Flags().StringP("description", "d", "", "Description of the partition. [required]")
	partitionCreateCmd.Flags().StringP("mgmtserver", "", "", "management server address in the partition. [required]")
	partitionCreateCmd.Flags().StringP("cmdline", "", "", "kernel commandline for the metal-hammer in the partition. [required]")
	partitionCreateCmd.Flags().StringP("imageurl", "", "", "initrd for the metal-hammer in the partition. [required]")
	partitionCreateCmd.Flags().StringP("kernelurl", "", "", "kernel url for the metal-hammer in the partition. [required]")

	partitionApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl partition describe partition-a > a.yaml
# vi a.yaml
## either via stdin
# cat a.yaml | metalctl partition apply -f -
## or via file
# metalctl partition apply -f a.yaml`)
	err := partitionApplyCmd.MarkFlagRequired("file")
	if err != nil {
		log.Fatal(err.Error())
	}

	partitionUpdateCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl partition describe partition-a > a.yaml
# vi a.yaml
## either via stdin
# cat a.yaml | metalctl partition update -f -
## or via file
# metalctl partition update -f a.yaml`)
	err = partitionUpdateCmd.MarkFlagRequired("file")
	if err != nil {
		log.Fatal(err.Error())
	}

	partitionCmd.AddCommand(partitionListCmd)
	partitionCmd.AddCommand(partitionCapacityCmd)
	partitionCmd.AddCommand(partitionDescribeCmd)
	partitionCmd.AddCommand(partitionCreateCmd)
	partitionCmd.AddCommand(partitionUpdateCmd)
	partitionCmd.AddCommand(partitionApplyCmd)
	partitionCmd.AddCommand(partitionDeleteCmd)
	partitionCmd.AddCommand(partitionEditCmd)
}

func partitionList(driver *metalgo.Driver) error {
	resp, err := driver.PartitionList()
	if err != nil {
		return err
	}
	return printer.Print(resp.Partition)
}

func partitionDescribe(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no partition ID given")
	}
	partitionID := args[0]
	resp, err := driver.PartitionGet(partitionID)
	if err != nil {
		return err
	}
	return detailer.Detail(resp.Partition)
}
func partitionCapacity(driver *metalgo.Driver) error {
	resp, err := driver.PartitionCapacity()
	if err != nil {
		return err
	}
	return printer.Print(resp.Capacity)
}
func partitionCreate(driver *metalgo.Driver) error {
	var icrs []metalgo.PartitionCreateRequest
	var icr metalgo.PartitionCreateRequest
	if viper.GetString("file") != "" {
		err := readFrom(viper.GetString("file"), &icr, func(data interface{}) {
			doc := data.(*metalgo.PartitionCreateRequest)
			icrs = append(icrs, *doc)
		})
		if err != nil {
			return err
		}
		if len(icrs) != 1 {
			return fmt.Errorf("partition create error more or less than one partition given:%d", len(icrs))
		}
		icr = icrs[0]
	} else {
		icr = metalgo.PartitionCreateRequest{
			Description:        viper.GetString("description"),
			ID:                 viper.GetString("id"),
			Name:               viper.GetString("name"),
			Mgmtserviceaddress: viper.GetString("mgmtserver"),
			Bootconfig: metalgo.BootConfig{
				Commandline: viper.GetString("cmdline"),
				Imageurl:    viper.GetString("imageurl"),
				Kernelurl:   viper.GetString("kernelurl"),
			},
		}
	}

	resp, err := driver.PartitionCreate(icr)
	if err != nil {
		return err
	}
	return detailer.Detail(resp.Partition)
}

func partitionUpdate(driver *metalgo.Driver) error {
	icrs, err := readPartitionCreateRequests(viper.GetString("file"))
	if err != nil {
		return err
	}
	if len(icrs) != 1 {
		return fmt.Errorf("partition update error more or less than one partition given:%d", len(icrs))
	}
	resp, err := driver.PartitionUpdate(icrs[0])
	if err != nil {
		return err
	}
	return detailer.Detail(resp.Partition)
}

func readPartitionCreateRequests(filename string) ([]metalgo.PartitionCreateRequest, error) {
	var icrs []metalgo.PartitionCreateRequest
	var uir metalgo.PartitionCreateRequest
	err := readFrom(filename, &uir, func(data interface{}) {
		doc := data.(*metalgo.PartitionCreateRequest)
		icrs = append(icrs, *doc)
	})
	if err != nil {
		return nil, err
	}
	if len(icrs) != 1 {
		return nil, fmt.Errorf("partition update error more or less than one partition given:%d", len(icrs))
	}
	return icrs, nil
}

// TODO: General apply method would be useful as these are quite a lot of lines and it's getting erroneous
func partitionApply(driver *metalgo.Driver) error {
	var iars []metalgo.PartitionCreateRequest
	var iar metalgo.PartitionCreateRequest
	err := readFrom(viper.GetString("file"), &iar, func(data interface{}) {
		doc := data.(*metalgo.PartitionCreateRequest)
		iars = append(iars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		iar = metalgo.PartitionCreateRequest{}
	})
	if err != nil {
		return err
	}
	var response []*models.V1PartitionResponse
	for _, iar := range iars {
		resp, err := driver.PartitionGet(iar.ID)
		if err != nil {
			var fe *partitionmodel.FindPartitionDefault
			if errors.As(err, &fe) {
				if fe.Code() != http.StatusNotFound {
					return err
				}
			}
			return err
		}
		if resp.Partition == nil {
			resp, err := driver.PartitionCreate(iar)
			if err != nil {
				return err
			}
			response = append(response, resp.Partition)
			continue
		}

		updateResponse, err := driver.PartitionUpdate(iar)
		if err != nil {
			return err
		}
		response = append(response, updateResponse.Partition)
	}
	return detailer.Detail(response)
}

func partitionDelete(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no partition ID given")
	}
	partitionID := args[0]
	resp, err := driver.PartitionDelete(partitionID)
	if err != nil {
		return err
	}
	return detailer.Detail(resp.Partition)
}

func partitionEdit(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no partition ID given")
	}
	partitionID := args[0]

	getFunc := func(id string) ([]byte, error) {
		resp, err := driver.PartitionGet(partitionID)
		if err != nil {
			return nil, err
		}
		content, err := yaml.Marshal(resp.Partition)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		iars, err := readPartitionCreateRequests(filename)
		if err != nil {
			return err
		}
		if len(iars) != 1 {
			return fmt.Errorf("partition update error more or less than one partition given:%d", len(iars))
		}
		uresp, err := driver.PartitionUpdate(iars[0])
		if err != nil {
			return err
		}
		return detailer.Detail(uresp.Partition)
	}

	return edit(partitionID, getFunc, updateFunc)
}
