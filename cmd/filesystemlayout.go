package cmd

import (
	"errors"
	"fmt"

	fsmodel "github.com/metal-stack/metal-go/api/client/filesystemlayout"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
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
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI[*models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, *models.V1FilesystemLayoutResponse](w).WithFS(c.fs),
		Singular:        "filesystemlayout",
		Plural:          "filesystemlayouts",
		Description:     "a filesystemlayout is a specification how the disks in a machine are partitioned, formatted and mounted.",
		Aliases:         []string{"fsl"},
		Sorter:          sorters.FilesystemLayoutSorter(),
		ValidArgsFn:     c.comp.FilesystemLayoutListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
	}

	tryCmd := &cobra.Command{
		Use:   "try",
		Short: "try to detect a filesystem by given size and image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.filesystemTry()
		},
	}

	tryCmd.Flags().StringP("size", "", "", "size to try")
	tryCmd.Flags().StringP("image", "", "", "image to try")
	genericcli.Must(tryCmd.MarkFlagRequired("size"))
	genericcli.Must(tryCmd.MarkFlagRequired("image"))
	genericcli.Must(tryCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	genericcli.Must(tryCmd.RegisterFlagCompletionFunc("image", c.comp.ImageListCompletion))

	matchCmd := &cobra.Command{
		Use:   "match",
		Short: "check if a machine satisfies all disk requirements of a given filesystemlayout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.filesystemMatch()
		},
	}

	matchCmd.Flags().StringP("machine", "", "", "machine id to check for match [required]")
	matchCmd.Flags().StringP("filesystemlayout", "", "", "filesystemlayout id to check against [required]")
	genericcli.Must(matchCmd.MarkFlagRequired("machine"))
	genericcli.Must(matchCmd.MarkFlagRequired("filesystemlayout"))
	genericcli.Must(matchCmd.RegisterFlagCompletionFunc("machine", c.comp.MachineListCompletion))
	genericcli.Must(matchCmd.RegisterFlagCompletionFunc("filesystemlayout", c.comp.FilesystemLayoutListCompletion))

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

func (c fslCmd) Convert(r *models.V1FilesystemLayoutResponse) (string, *models.V1FilesystemLayoutCreateRequest, *models.V1FilesystemLayoutUpdateRequest, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}

	return *r.ID, filesystemLayoutResponseToCreate(r), filesystemLayoutResponseToUpdate(r), nil
}

func filesystemLayoutResponseToCreate(r *models.V1FilesystemLayoutResponse) *models.V1FilesystemLayoutCreateRequest {
	return &models.V1FilesystemLayoutCreateRequest{
		Constraints:    r.Constraints,
		Description:    r.Description,
		Disks:          r.Disks,
		Filesystems:    r.Filesystems,
		ID:             r.ID,
		Logicalvolumes: r.Logicalvolumes,
		Name:           r.Name,
		Raid:           r.Raid,
		Volumegroups:   r.Volumegroups,
	}
}

func filesystemLayoutResponseToUpdate(r *models.V1FilesystemLayoutResponse) *models.V1FilesystemLayoutUpdateRequest {
	return &models.V1FilesystemLayoutUpdateRequest{
		Constraints:    r.Constraints,
		Description:    r.Description,
		Disks:          r.Disks,
		Filesystems:    r.Filesystems,
		ID:             r.ID,
		Logicalvolumes: r.Logicalvolumes,
		Name:           r.Name,
		Raid:           r.Raid,
		Volumegroups:   r.Volumegroups,
	}
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

	return c.listPrinter.Print(resp.Payload)
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

	return c.listPrinter.Print(resp.Payload)
}
