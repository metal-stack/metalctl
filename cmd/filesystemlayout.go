package cmd

import (
	"errors"

	fsmodel "github.com/metal-stack/metal-go/api/client/filesystemlayout"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type fslCmd struct {
	*config
}

func newFilesystemLayoutCmd(c *config) *cobra.Command {
	w := &fslCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse]{
		BinaryName:        binaryName,
		GenericCLI:        genericcli.NewGenericCLI[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse](w),
		Singular:          "filesystemlayout",
		Plural:            "filesystemlayouts",
		Description:       "a filesystemlayout is a specification how the disks in a machine are partitioned, formatted and mounted.",
		Aliases:           []string{"fsl"},
		AvailableSortKeys: sorters.FilesystemLayoutSortKeys(),
		ValidArgsFn:       c.comp.FilesystemLayoutListCompletion,
		DescribePrinter:   DefaultToYAMLPrinter(),
		ListPrinter:       NewPrinterFromCLI(),
	}

	tryCmd := &cobra.Command{
		Use:   "try",
		Short: "try to detect a filesystem by given size and image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.filesystemTry()
		},
		PreRun: bindPFlags,
	}

	tryCmd.Flags().StringP("size", "", "", "size to try")
	tryCmd.Flags().StringP("image", "", "", "image to try")
	must(tryCmd.MarkFlagRequired("size"))
	must(tryCmd.MarkFlagRequired("image"))
	must(tryCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(tryCmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))

	matchCmd := &cobra.Command{
		Use:   "match",
		Short: "check if a machine satisfies all disk requirements of a given filesystemlayout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.filesystemMatch()
		},
		PreRun: bindPFlags,
	}

	matchCmd.Flags().StringP("machine", "", "", "machine id to check for match [required]")
	matchCmd.Flags().StringP("filesystemlayout", "", "", "filesystemlayout id to check against [required]")
	must(matchCmd.MarkFlagRequired("machine"))
	must(matchCmd.MarkFlagRequired("filesystemlayout"))
	must(matchCmd.RegisterFlagCompletionFunc("machine", c.comp.MachineListCompletion))
	must(matchCmd.RegisterFlagCompletionFunc("filesystemlayout", c.comp.FilesystemLayoutListCompletion))

	return genericcli.NewCmds(cmdsConfig, tryCmd, matchCmd)
}

func (c fslCmd) Get(id string) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.client.Filesystemlayout().GetFilesystemLayout(fsmodel.NewGetFilesystemLayoutParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCmd) List() ([]*models.V1FilesystemLayoutResponse, error) {
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

func (c fslCmd) Delete(id string) (*models.V1FilesystemLayoutResponse, error) {
	resp, err := c.client.Filesystemlayout().DeleteFilesystemLayout(fsmodel.NewDeleteFilesystemLayoutParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c fslCmd) Create(rq *models.V1FilesystemLayoutCreateRequest) (*models.V1FilesystemLayoutResponse, error) {
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

func (c fslCmd) Update(rq *models.V1FilesystemLayoutUpdateRequest) (*models.V1FilesystemLayoutResponse, error) {
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

	return NewPrinterFromCLI().Print(resp.Payload)
}

func (c *fslCmd) filesystemMatch() error {
	match := models.V1FilesystemLayoutMatchRequest{
		Machine:          pointer.Pointer(viper.GetString("machine")),
		Filesystemlayout: pointer.Pointer(viper.GetString("filesystemlayout")),
	}

	resp, err := c.client.Filesystemlayout().MatchFilesystemLayout(fsmodel.NewMatchFilesystemLayoutParams().WithBody(&match), nil)
	if err != nil {
		return err
	}

	return NewPrinterFromCLI().Print(resp.Payload)
}
