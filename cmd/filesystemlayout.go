package cmd

import (
	"errors"
	"fmt"
	"net/http"

	fsmodel "github.com/metal-stack/metal-go/api/client/filesystemlayout"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metalctl/cmd/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newFilesystemLayoutCmd(c *config) *cobra.Command {
	filesystemLayoutCmd := &cobra.Command{
		Use:     "filesystemlayout",
		Aliases: []string{"fsl"},
		Short:   "manage filesystemlayouts",
		Long:    "a filesystemlayout is a specification how the disks in a machine are partitioned, formatted and mounted.",
	}

	filesystemListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all filesystems",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.filesystemList()
		},
		PreRun: bindPFlags,
	}
	filesystemDescribeCmd := &cobra.Command{
		Use:   "describe <filesystemID>",
		Short: "describe a filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.filesystemDescribe(args)
		},
		ValidArgsFunction: c.comp.FilesystemLayoutListCompletion,
	}
	filesystemApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a filesystem",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.filesystemApply()
		},
		PreRun: bindPFlags,
	}
	filesystemDeleteCmd := &cobra.Command{
		Use:     "delete <filesystemID>",
		Short:   "delete a filesystem",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.filesystemDelete(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.FilesystemLayoutListCompletion,
	}
	filesystemTryCmd := &cobra.Command{
		Use:   "try",
		Short: "try to detect a filesystem by given size and image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.filesystemTry()
		},
		PreRun: bindPFlags,
	}
	filesystemMatchCmd := &cobra.Command{
		Use:   "match",
		Short: "check if a machine satisfies all disk requirements of a given filesystemlayout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.filesystemMatch()
		},
		PreRun: bindPFlags,
	}
	filesystemApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl filesystem describe default > default.yaml
# vi default.yaml
## either via stdin
# cat default.yaml | metalctl filesystem apply -f -
## or via file
# metalctl filesystem apply -f default.yaml`)
	must(filesystemApplyCmd.MarkFlagRequired("file"))

	filesystemTryCmd.Flags().StringP("size", "", "", "size to try")
	filesystemTryCmd.Flags().StringP("image", "", "", "image to try")
	must(filesystemTryCmd.MarkFlagRequired("size"))
	must(filesystemTryCmd.MarkFlagRequired("image"))
	must(filesystemTryCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(filesystemTryCmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))

	filesystemMatchCmd.Flags().StringP("machine", "", "", "machine id to check for match [required]")
	filesystemMatchCmd.Flags().StringP("filesystemlayout", "", "", "filesystemlayout id to check against [required]")
	must(filesystemMatchCmd.MarkFlagRequired("machine"))
	must(filesystemMatchCmd.MarkFlagRequired("filesystemlayout"))
	must(filesystemMatchCmd.RegisterFlagCompletionFunc("machine", c.comp.MachineListCompletion))
	must(filesystemMatchCmd.RegisterFlagCompletionFunc("filesystemlayout", c.comp.FilesystemLayoutListCompletion))

	filesystemLayoutCmd.AddCommand(filesystemListCmd)
	filesystemLayoutCmd.AddCommand(filesystemDescribeCmd)
	filesystemLayoutCmd.AddCommand(filesystemDeleteCmd)
	filesystemLayoutCmd.AddCommand(filesystemApplyCmd)
	filesystemLayoutCmd.AddCommand(filesystemTryCmd)
	filesystemLayoutCmd.AddCommand(filesystemMatchCmd)

	return filesystemLayoutCmd
}

func (c *config) filesystemList() error {
	resp, err := c.driver.FilesystemLayoutList()
	if err != nil {
		return err
	}
	return output.New().Print(resp)
}

func (c *config) filesystemDescribe(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no filesystem ID given")
	}
	filesystemID := args[0]
	resp, err := c.driver.FilesystemLayoutGet(filesystemID)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp)
}

// TODO: General apply method would be useful as these are quite a lot of lines and it's getting erroneous
func (c *config) filesystemApply() error {
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
		p, err := c.driver.FilesystemLayoutGet(*iar.ID)
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
			resp, err := c.driver.FilesystemLayoutCreate(iar)
			if err != nil {
				return err
			}
			response = append(response, resp)
			continue
		}

		resp, err := c.driver.FilesystemLayoutUpdate(models.V1FilesystemLayoutUpdateRequest(iar))
		if err != nil {
			return err
		}
		response = append(response, resp)
	}
	return output.New().Print(response)
}

func (c *config) filesystemDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no filesystem ID given")
	}
	filesystemID := args[0]
	resp, err := c.driver.FilesystemLayoutDelete(filesystemID)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp)
}

func (c *config) filesystemTry() error {
	size := viper.GetString("size")
	image := viper.GetString("image")
	try := models.V1FilesystemLayoutTryRequest{
		Size:  &size,
		Image: &image,
	}

	resp, err := c.driver.FilesystemLayoutTry(try)
	if err != nil {
		return err
	}
	return output.New().Print(resp)
}
func (c *config) filesystemMatch() error {
	machine := viper.GetString("machine")
	fsl := viper.GetString("filesystemlayout")
	match := models.V1FilesystemLayoutMatchRequest{
		Machine:          &machine,
		Filesystemlayout: &fsl,
	}

	resp, err := c.driver.FilesystemLayoutMatch(match)
	if err != nil {
		return err
	}
	return output.New().Print(resp)
}
