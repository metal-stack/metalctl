package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/metal-stack/metal-go/api/client/partition"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type partitionCmd struct {
	*config
	*genericcli.GenericCLI[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse]
}

func newPartitionCmd(c *config) *cobra.Command {
	w := partitionCmd{
		config:     c,
		GenericCLI: genericcli.NewGenericCLI[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse](partitionCRUD{config: c}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse]{
		gcli:              w.GenericCLI,
		singular:          "partition",
		plural:            "partitions",
		description:       "a partition is a group of machines and network which is logically separated from other partitions. Machines have no direct network connections between partitions.",
		validArgsFunc:     c.comp.PartitionListCompletion,
		availableSortKeys: sorters.PartitionSortKeys(),
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
			return w.partitionCapacity()
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
	partitionCapacityCmd.Flags().StringSlice("order", []string{}, fmt.Sprintf("order by (comma separated) column(s), sort direction can be changed by appending :asc or :desc behind the column identifier. possible values: %s", strings.Join(sorters.PartitionCapacitySortKeys(), "|")))
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("id", c.comp.PartitionListCompletion))
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("order", cobra.FixedCompletions(sorters.PartitionCapacitySortKeys(), cobra.ShellCompDirectiveNoFileComp)))

	return cmds.buildRootCmd(partitionCapacityCmd)
}

type partitionCRUD struct {
	*config
}

func (c partitionCRUD) Get(id string) (*models.V1PartitionResponse, error) {
	resp, err := c.client.Partition().FindPartition(partition.NewFindPartitionParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCRUD) List() ([]*models.V1PartitionResponse, error) {
	resp, err := c.client.Partition().ListPartitions(partition.NewListPartitionsParams(), nil)
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
	resp, err := c.client.Partition().DeletePartition(partition.NewDeletePartitionParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCRUD) Create(rq *models.V1PartitionCreateRequest) (*models.V1PartitionResponse, error) {
	resp, err := c.client.Partition().CreatePartition(partition.NewCreatePartitionParams().WithBody(rq), nil)
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
	resp, err := c.client.Partition().UpdatePartition(partition.NewUpdatePartitionParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *partitionCmd) partitionCapacity() error {
	resp, err := c.client.Partition().PartitionCapacity(partition.NewPartitionCapacityParams().WithBody(&models.V1PartitionCapacityRequest{
		ID:     viper.GetString("id"),
		Sizeid: viper.GetString("size"),
	}), nil)
	if err != nil {
		return err
	}

	err = sorters.PartitionCapacitySort(resp.Payload)
	if err != nil {
		return err
	}

	return newPrinterFromCLI().Print(resp.Payload)
}
