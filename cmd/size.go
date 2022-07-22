package cmd

import (
	"errors"

	"github.com/dustin/go-humanize"
	"github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeCmd struct {
	*config
	*genericcli.GenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]
}

func newSizeCmd(c *config) *cobra.Command {
	w := sizeCmd{
		config:     c,
		GenericCLI: genericcli.NewGenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse](sizeCRUD{config: c}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]{
		gcli:              w.GenericCLI,
		singular:          "size",
		plural:            "sizes",
		description:       "a size is a distinct hardware equipment in terms of cpu cores, ram and storage of a machine.",
		availableSortKeys: sorters.SizeSortKeys(),
		validArgsFunc:     c.comp.SizeListCompletion,
		createRequestFromCLI: func() (*models.V1SizeCreateRequest, error) {
			return &models.V1SizeCreateRequest{
				ID:          pointer.Pointer(viper.GetString("id")),
				Name:        viper.GetString("name"),
				Description: viper.GetString("description"),
				Constraints: []*models.V1SizeConstraint{
					{
						Max:  pointer.Pointer(viper.GetInt64("max")),
						Min:  pointer.Pointer(viper.GetInt64("min")),
						Type: pointer.Pointer(viper.GetString("type")),
					},
				},
			}, nil
		},
	})

	tryCmd := &cobra.Command{
		Use:   "try",
		Short: "try a specific hardware spec and give the chosen size back",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.try()
		},
		PreRun: bindPFlags,
	}

	cmds.createCmd.Flags().StringP("id", "", "", "ID of the size. [required]")
	cmds.createCmd.Flags().StringP("name", "n", "", "Name of the size. [optional]")
	cmds.createCmd.Flags().StringP("description", "d", "", "Description of the size. [required]")
	// FIXME constraints must be given in a slice
	cmds.createCmd.Flags().Int64P("min", "", 0, "min value of given size constraint type. [required]")
	cmds.createCmd.Flags().Int64P("max", "", 0, "min value of given size constraint type. [required]")
	cmds.createCmd.Flags().StringP("type", "", "", "type of constraints. [required]")

	tryCmd.Flags().Int32P("cores", "C", 0, "Cores of the hardware to try")
	tryCmd.Flags().StringP("memory", "M", "", "Memory of the hardware to try, can be given in bytes or any human readable size spec")
	tryCmd.Flags().StringP("storagesize", "S", "", "Total storagesize of the hardware to try, can be given in bytes or any human readable size spec")

	return cmds.buildRootCmd(newSizeImageConstraintCmd(c), tryCmd)
}

type sizeCRUD struct {
	*config
}

func (c sizeCRUD) Get(id string) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().FindSize(size.NewFindSizeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeCRUD) List() ([]*models.V1SizeResponse, error) {
	resp, err := c.client.Size().ListSizes(size.NewListSizesParams(), nil)
	if err != nil {
		return nil, err
	}

	err = sorters.SizeSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeCRUD) Delete(id string) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().DeleteSize(size.NewDeleteSizeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeCRUD) Create(rq *models.V1SizeCreateRequest) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().CreateSize(size.NewCreateSizeParams().WithBody(rq), nil)
	if err != nil {
		var r *size.CreateSizeConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeCRUD) Update(rq *models.V1SizeUpdateRequest) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().UpdateSize(size.NewUpdateSizeParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *sizeCmd) try() error {
	var (
		memory int64
		disks  []*models.V1MachineBlockDevice
	)

	if viper.IsSet("memory") {
		m, err := humanize.ParseBytes(viper.GetString("memory"))
		if err != nil {
			return err
		}
		memory = int64(m)
	}

	if viper.IsSet("storagesize") {
		s, err := humanize.ParseBytes(viper.GetString("storagesize"))
		if err != nil {
			return err
		}
		disks = append(disks, &models.V1MachineBlockDevice{
			Name: pointer.Pointer("/dev/trydisk"),
			Size: pointer.Pointer(int64(s)),
		})
	}

	resp, err := c.client.Size().FromHardware(size.NewFromHardwareParams().WithBody(&models.V1MachineHardware{
		CPUCores: pointer.Pointer(viper.GetInt32("cores")),
		Memory:   &memory,
		Disks:    disks,
	}), nil)
	if err != nil {
		return err
	}

	return newPrinterFromCLI().Print(resp.Payload)
}
