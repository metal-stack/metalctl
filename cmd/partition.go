package cmd

import (
	"errors"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/partition"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type partitionCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	*genericcli.GenericCLI[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse]
}

func newPartitionCmd(c *config) *cobra.Command {
	w := partitionCmd{
		c:          c.client,
		driver:     c.driver,
		GenericCLI: genericcli.NewGenericCLI[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse](partitionCRUD{c: c.client}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse]{
		gcli:              w.GenericCLI,
		singular:          "partition",
		plural:            "partitions",
		description:       "a partition is a group of machines and network which is logically separated from other partitions. Machines have no direct network connections between partitions.",
		validArgsFunc:     c.comp.PartitionListCompletion,
		availableSortKeys: sorters.PartitionSorter().AvailableKeys(),
		createRequestFromCLI: func() (*models.V1PartitionCreateRequest, error) {
			return &models.V1PartitionCreateRequest{
				ID:                 pointer.Pointer(viper.GetString("id")),
				Description:        viper.GetString("description"),
				Name:               viper.GetString("name"),
				Mgmtserviceaddress: viper.GetString("mgmtserver"),
				Bootconfig: &models.V1PartitionBootConfiguration{
					Commandline: viper.GetString("cmdline"),
					Imageurl:    viper.GetString("imageurl"),
					Kernelurl:   viper.GetString("kernelurl"),
				},
			}, nil
		},
	})

	partitionCapacityCmd := &cobra.Command{
		Use:   "capacity",
		Short: "show partition capacity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.partitionCapacity()
		},
		PreRun: bindPFlags,
	}

	cmds.createCmd.Flags().StringP("id", "", "", "ID of the partition. [required]")
	cmds.createCmd.Flags().StringP("name", "n", "", "Name of the partition. [optional]")
	cmds.createCmd.Flags().StringP("description", "d", "", "Description of the partition. [required]")
	cmds.createCmd.Flags().StringP("mgmtserver", "", "", "management server address in the partition. [required]")
	cmds.createCmd.Flags().StringP("cmdline", "", "", "kernel commandline for the metal-hammer in the partition. [required]")
	cmds.createCmd.Flags().StringP("imageurl", "", "", "initrd for the metal-hammer in the partition. [required]")
	cmds.createCmd.Flags().StringP("kernelurl", "", "", "kernel url for the metal-hammer in the partition. [required]")

	partitionCapacityCmd.Flags().StringP("id", "", "", "filter on partition id. [optional]")
	partitionCapacityCmd.Flags().StringP("size", "", "", "filter on size id. [optional]")
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("id", c.comp.PartitionListCompletion))
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))

	return cmds.buildRootCmd(partitionCapacityCmd)
}

type partitionCRUD struct {
	c metalgo.Client
}

func (c partitionCRUD) Get(id string) (*models.V1PartitionResponse, error) {
	resp, err := c.c.Partition().FindPartition(partition.NewFindPartitionParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCRUD) List() ([]*models.V1PartitionResponse, error) {
	resp, err := c.c.Partition().ListPartitions(partition.NewListPartitionsParams(), nil)
	if err != nil {
		return nil, err
	}

	err = sorters.PartitionSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCRUD) Delete(id string) (*models.V1PartitionResponse, error) {
	resp, err := c.c.Partition().DeletePartition(partition.NewDeletePartitionParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCRUD) Create(rq *models.V1PartitionCreateRequest) (*models.V1PartitionResponse, error) {
	resp, err := c.c.Partition().CreatePartition(partition.NewCreatePartitionParams().WithBody(rq), nil)
	if err != nil {
		var r *partition.CreatePartitionConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCRUD) Update(rq *models.V1PartitionUpdateRequest) (*models.V1PartitionResponse, error) {
	resp, err := c.c.Partition().UpdatePartition(partition.NewUpdatePartitionParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *config) partitionCapacity() error {
	var (
		pcr  = metalgo.PartitionCapacityRequest{}
		id   = viper.GetString("id")
		size = viper.GetString("size")
	)

	if id != "" {
		pcr.ID = &id
	}
	if size != "" {
		pcr.Size = &size
	}

	resp, err := c.driver.PartitionCapacity(pcr)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Capacity)
}
