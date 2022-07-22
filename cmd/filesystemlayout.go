package cmd

import (
	"errors"

	fsmodel "github.com/metal-stack/metal-go/api/client/filesystemlayout"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/sorters"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type fslCmd struct {
	*config
	*genericcli.GenericCLI[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse]
}

func newFilesystemLayoutCmd(c *config) *cobra.Command {
	w := fslCmd{
		config:     c,
		GenericCLI: genericcli.NewGenericCLI[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse](fslCRUD{config: c}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse]{
		gcli:              w.GenericCLI,
		singular:          "filesystemlayout",
		plural:            "filesystemlayouts",
		description:       "a filesystemlayout is a specification how the disks in a machine are partitioned, formatted and mounted.",
		aliases:           []string{"fsl"},
		availableSortKeys: sorters.FilesystemLayoutSortKeys(),
		validArgsFunc:     c.comp.FilesystemLayoutListCompletion,
	})

	tryCmd := &cobra.Command{
		Use:   "try",
		Short: "try to detect a filesystem by given size and image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.filesystemTry()
		},
		PreRun: bindPFlags,
	}

	matchCmd := &cobra.Command{
		Use:   "match",
		Short: "check if a machine satisfies all disk requirements of a given filesystemlayout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.filesystemMatch()
		},
		PreRun: bindPFlags,
	}

	tryCmd.Flags().StringP("size", "", "", "size to try")
	tryCmd.Flags().StringP("image", "", "", "image to try")
	must(tryCmd.MarkFlagRequired("size"))
	must(tryCmd.MarkFlagRequired("image"))
	must(tryCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(tryCmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))

	matchCmd.Flags().StringP("machine", "", "", "machine id to check for match [required]")
	matchCmd.Flags().StringP("filesystemlayout", "", "", "filesystemlayout id to check against [required]")
	must(matchCmd.MarkFlagRequired("machine"))
	must(matchCmd.MarkFlagRequired("filesystemlayout"))
	must(matchCmd.RegisterFlagCompletionFunc("machine", c.comp.MachineListCompletion))
	must(matchCmd.RegisterFlagCompletionFunc("filesystemlayout", c.comp.FilesystemLayoutListCompletion))

	return cmds.buildRootCmd(matchCmd, tryCmd)
}

type fslCRUD struct {
	*config
}

func (c fslCRUD) Get(id string) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.client.Filesystemlayout().GetFilesystemLayout(fsmodel.NewGetFilesystemLayoutParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCRUD) List() ([]*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.client.Filesystemlayout().ListFilesystemLayouts(fsmodel.NewListFilesystemLayoutsParams(), nil)
	if err != nil {
		return nil, err
	}

	err = sorters.FilesystemLayoutSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCRUD) Delete(id string) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.client.Filesystemlayout().DeleteFilesystemLayout(fsmodel.NewDeleteFilesystemLayoutParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCRUD) Create(rq *models.V1FilesystemLayoutCreateRequest) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.client.Filesystemlayout().CreateFilesystemLayout(fsmodel.NewCreateFilesystemLayoutParams().WithBody(rq), nil)
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
	resp, err := c.client.Filesystemlayout().UpdateFilesystemLayout(fsmodel.NewUpdateFilesystemLayoutParams().WithBody(rq), nil)
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

	resp, err := c.client.Filesystemlayout().TryFilesystemLayout(fsmodel.NewTryFilesystemLayoutParams().WithBody(&try), nil)
	if err != nil {
		return err
	}

	return newPrinterFromCLI().Print(resp.Payload)
}

func (c *fslCmd) filesystemMatch() error {
	machine := viper.GetString("machine")
	fsl := viper.GetString("filesystemlayout")
	match := models.V1FilesystemLayoutMatchRequest{
		Machine:          &machine,
		Filesystemlayout: &fsl,
	}

	resp, err := c.client.Filesystemlayout().MatchFilesystemLayout(fsmodel.NewMatchFilesystemLayoutParams().WithBody(&match), nil)
	if err != nil {
		return err
	}

	return newPrinterFromCLI().Print(resp.Payload)
}
