package cmd

import (
	"fmt"
	"log"
	"net/http"

	metalgo "github.com/metal-stack/metal-go"
	fsmodel "github.com/metal-stack/metal-go/api/client/filesystemlayout"
	"github.com/metal-stack/metal-go/api/models"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	filesystemLayoutCmd = &cobra.Command{
		Use:   "filesystemlayout",
		Short: "manage filesystemlayouts",
		Long:  "a filesystemlayout is a specification how the disks in a machine are partitioned, formatted and mounted.",
	}

	filesystemListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all filesystems",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filesystemList(driver)
		},
		PreRun: bindPFlags,
	}
	filesystemDescribeCmd = &cobra.Command{
		Use:   "describe <filesystemID>",
		Short: "describe a filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filesystemDescribe(driver, args)
		},
	}
	filesystemApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filesystemApply(driver)
		},
		PreRun: bindPFlags,
	}
	filesystemDeleteCmd = &cobra.Command{
		Use:   "delete <filesystemID>",
		Short: "delete a filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filesystemDelete(driver, args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	filesystemApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl filesystem describe default > default.yaml
# vi default.yaml
## either via stdin
# cat default.yaml | metalctl filesystem apply -f -
## or via file
# metalctl filesystem apply -f default.yaml`)
	err := filesystemApplyCmd.MarkFlagRequired("file")
	if err != nil {
		log.Fatal(err.Error())
	}

	filesystemLayoutCmd.AddCommand(filesystemListCmd)
	filesystemLayoutCmd.AddCommand(filesystemDescribeCmd)
	filesystemLayoutCmd.AddCommand(filesystemDeleteCmd)
	filesystemLayoutCmd.AddCommand(filesystemApplyCmd)
}

func filesystemList(driver *metalgo.Driver) error {
	resp, err := driver.FilesystemLayoutList()
	if err != nil {
		return err
	}
	return printer.Print(resp)
}

func filesystemDescribe(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no filesystem ID given")
	}
	filesystemID := args[0]
	resp, err := driver.FilesystemLayoutGet(filesystemID)
	if err != nil {
		return err
	}
	return detailer.Detail(resp)
}

// TODO: General apply method would be useful as these are quite a lot of lines and it's getting erroneous
func filesystemApply(driver *metalgo.Driver) error {
	var iars []models.V1FilesystemLayoutCreateRequest
	var iar models.V1FilesystemLayoutCreateRequest
	err := readFrom(viper.GetString("file"), &iar, func(data interface{}) {
		doc := data.(*models.V1FilesystemLayoutCreateRequest)
		iars = append(iars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		iar = models.V1FilesystemLayoutCreateRequest{}
	})
	if err != nil {
		return err
	}
	var response []*models.V1FilesystemLayoutResponse
	for _, iar := range iars {
		p, err := driver.FilesystemLayoutGet(*iar.ID)
		if err != nil {
			switch e := err.(type) {
			case *fsmodel.GetFilesystemLayoutDefault:
				if e.Code() != http.StatusNotFound {
					return err
				}
			default:
				return err
			}
		}
		if p == nil {
			resp, err := driver.FilesystemLayoutCreate(iar)
			fmt.Printf("error:%v\n", err)
			if err != nil {
				return err
			}
			response = append(response, resp)
			continue
		}

		resp, err := driver.FilesystemLayoutUpdate(models.V1FilesystemLayoutUpdateRequest(iar))
		if err != nil {
			return err
		}
		response = append(response, resp)
	}
	return printer.Print(response)
}

func filesystemDelete(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no filesystem ID given")
	}
	filesystemID := args[0]
	resp, err := driver.FilesystemLayoutDelete(filesystemID)
	if err != nil {
		return err
	}
	return detailer.Detail(resp)
}
