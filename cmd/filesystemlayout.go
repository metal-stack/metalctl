package cmd

import (
	"errors"

	metalgo "github.com/metal-stack/metal-go"
	fsmodel "github.com/metal-stack/metal-go/api/client/filesystemlayout"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type fslCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	*genericcli.GenericCLI[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse]
}

func newFilesystemLayoutCmd(c *config) *cobra.Command {
	w := fslCmd{
		c:          c.client,
		driver:     c.driver,
		GenericCLI: genericcli.NewGenericCLI[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse](fslCRUD{Client: c.client}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse]{
		gcli:          w.GenericCLI,
		singular:      "filesystemlayout",
		plural:        "filesystemlayouts",
		description:   "a filesystemlayout is a specification how the disks in a machine are partitioned, formatted and mounted.",
		aliases:       []string{"fsl"},
		validArgsFunc: c.comp.FilesystemLayoutListCompletion,
	})

	filesystemTryCmd := &cobra.Command{
		Use:   "try",
		Short: "try to detect a filesystem by given size and image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.filesystemTry()
		},
		PreRun: bindPFlags,
	}

	filesystemMatchCmd := &cobra.Command{
		Use:   "match",
		Short: "check if a machine satisfies all disk requirements of a given filesystemlayout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.filesystemMatch()
		},
		PreRun: bindPFlags,
	}

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

	root := cmds.BuildRootCmd()

	root.AddCommand(filesystemTryCmd)
	root.AddCommand(filesystemMatchCmd)

	return root
}

type fslCRUD struct {
	metalgo.Client
}

func (c fslCRUD) Get(id string) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.Filesystemlayout().GetFilesystemLayout(fsmodel.NewGetFilesystemLayoutParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCRUD) List() ([]*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.Filesystemlayout().ListFilesystemLayouts(fsmodel.NewListFilesystemLayoutsParams(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCRUD) Delete(id string) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.Filesystemlayout().DeleteFilesystemLayout(fsmodel.NewDeleteFilesystemLayoutParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCRUD) Create(rq *models.V1FilesystemLayoutCreateRequest) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.Filesystemlayout().CreateFilesystemLayout(fsmodel.NewCreateFilesystemLayoutParams().WithBody(rq), nil)
	if err != nil {
		var r *fsmodel.CreateFilesystemLayoutConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCRUD) Update(rq *models.V1FilesystemLayoutUpdateRequest) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.Filesystemlayout().UpdateFilesystemLayout(fsmodel.NewUpdateFilesystemLayoutParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *fslCmd) filesystemTry() error {
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

func (c *fslCmd) filesystemMatch() error {
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
