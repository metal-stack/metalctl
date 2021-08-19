package cmd

import (
	"errors"
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
		Use:     "filesystemlayout",
		Aliases: []string{"fsl"},
		Short:   "manage filesystemlayouts",
		Long:    "a filesystemlayout is a specification how the disks in a machine are partitioned, formatted and mounted.",
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
		ValidArgsFunction: filesystemLayoutListCompletionFunc,
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
		PreRun:            bindPFlags,
		ValidArgsFunction: filesystemLayoutListCompletionFunc,
	}
	filesystemTryCmd = &cobra.Command{
		Use:   "try",
		Short: "try to detect a filesystem by given size and image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filesystemTry(driver)
		},
		PreRun: bindPFlags,
	}
	filesystemMatchCmd = &cobra.Command{
		Use:   "match",
		Short: "check if a machine satisfies all disk requirements of a given filesystemlayout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filesystemMatch(driver)
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

	filesystemTryCmd.Flags().StringP("size", "", "", "size to try")
	filesystemTryCmd.Flags().StringP("image", "", "", "image to try")
	err = filesystemTryCmd.MarkFlagRequired("size")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = filesystemTryCmd.MarkFlagRequired("image")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = filesystemTryCmd.RegisterFlagCompletionFunc("size", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return sizeListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = filesystemTryCmd.RegisterFlagCompletionFunc("image", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return imageListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	filesystemMatchCmd.Flags().StringP("machine", "", "", "machine id to check for match [required]")
	filesystemMatchCmd.Flags().StringP("filesystemlayout", "", "", "filesystemlayout id to check against [required]")
	err = filesystemMatchCmd.MarkFlagRequired("machine")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = filesystemMatchCmd.MarkFlagRequired("filesystemlayout")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = filesystemMatchCmd.RegisterFlagCompletionFunc("machine", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return machineListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = filesystemMatchCmd.RegisterFlagCompletionFunc("filesystemlayout", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return filesystemLayoutListCompletion(driver)
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	filesystemLayoutCmd.AddCommand(filesystemListCmd)
	filesystemLayoutCmd.AddCommand(filesystemDescribeCmd)
	filesystemLayoutCmd.AddCommand(filesystemDeleteCmd)
	filesystemLayoutCmd.AddCommand(filesystemApplyCmd)
	filesystemLayoutCmd.AddCommand(filesystemTryCmd)
	filesystemLayoutCmd.AddCommand(filesystemMatchCmd)
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
			var r *fsmodel.GetFilesystemLayoutDefault
			if !errors.As(err, &r) {
				return err
			}
			if r.Code() != http.StatusNotFound {
				return err
			}
		}
		if p == nil {
			resp, err := driver.FilesystemLayoutCreate(iar)
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

func filesystemTry(driver *metalgo.Driver) error {
	size := viper.GetString("size")
	image := viper.GetString("image")
	try := models.V1FilesystemLayoutTryRequest{
		Size:  &size,
		Image: &image,
	}

	resp, err := driver.FilesystemLayoutTry(try)
	if err != nil {
		return err
	}
	return printer.Print(resp)
}
func filesystemMatch(driver *metalgo.Driver) error {
	machine := viper.GetString("machine")
	fsl := viper.GetString("filesystemlayout")
	match := models.V1FilesystemLayoutMatchRequest{
		Machine:          &machine,
		Filesystemlayout: &fsl,
	}

	resp, err := driver.FilesystemLayoutMatch(match)
	if err != nil {
		return err
	}
	return printer.Print(resp)
}
