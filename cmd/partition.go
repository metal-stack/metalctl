package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/metal-stack/metal-go/api/client/partition"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type partitionCmd struct {
	*config
}

func newPartitionCmd(c *config) *cobra.Command {
	w := partitionCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse](w).WithFS(c.fs),
		Singular:        "partition",
		Plural:          "partitions",
		Description:     "a partition is a failure domain in the data center.",
		ValidArgsFn:     c.comp.PartitionListCompletion,
		Sorter:          sorters.PartitionSorter(),
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		CreateRequestFromCLI: func() (*models.V1PartitionCreateRequest, error) {
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
				Waitingpoolmaxsize: viper.GetString("waiting-pool-max-size"),
				Waitingpoolminsize: viper.GetString("waiting-pool-min-size"),
			}, nil
		},
		UpdateRequestFromCLI: func(args []string) (*models.V1PartitionUpdateRequest, error) {
			id, err := genericcli.GetExactlyOneArg(args)
			if err != nil {
				return nil, err
			}

			return &models.V1PartitionUpdateRequest{
				Description:        viper.GetString("description"),
				ID:                 pointer.Pointer(id),
				Mgmtserviceaddress: viper.GetString("mgmtserver"),
				Name:               viper.GetString("name"),
				Waitingpoolmaxsize: viper.GetString("waiting-pool-max-size"),
				Waitingpoolminsize: viper.GetString("waiting-pool-min-size"),
			}, nil
		},
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "", "", "ID of the partition. [required]")
			cmd.Flags().StringP("name", "n", "", "Name of the partition. [optional]")
			cmd.Flags().StringP("description", "d", "", "Description of the partition. [required]")
			cmd.Flags().StringP("mgmtserver", "", "", "management server address in the partition. [required]")
			cmd.Flags().StringP("cmdline", "", "", "kernel commandline for the metal-hammer in the partition. [required]")
			cmd.Flags().StringP("imageurl", "", "", "initrd for the metal-hammer in the partition. [required]")
			cmd.Flags().StringP("kernelurl", "", "", "kernel url for the metal-hammer in the partition. [required]")
			cmd.Flags().String("waiting-pool-min-size", "", "The minimum size of the waiting machine pool inside the partition (can be a number or percentage, e.g. 50% of the machines should be waiting, the rest will be shutdown). [optional]")
			cmd.Flags().String("waiting-pool-max-size", "", "The maximum size of the waiting machine pool inside the partition (can be a number or percentage, e.g. 70% of the machines should be waiting, the rest will be shutdown). [optional]")
		},
		UpdateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("name", "n", "", "Name of the partition. [optional]")
			cmd.Flags().StringP("description", "d", "", "Description of the partition. [required]")
			cmd.Flags().StringP("mgmtserver", "", "", "management server address in the partition. [required]")
			cmd.Flags().String("waiting-pool-min-size", "", "The minimum size of the waiting machine pool inside the partition (can be a number or percentage, e.g. 50% of the machines should be waiting, the rest will be shutdown). [optional]")
			cmd.Flags().String("waiting-pool-max-size", "", "The maximum size of the waiting machine pool inside the partition (can be a number or percentage, e.g. 70% of the machines should be waiting, the rest will be shutdown). [optional]")
		},
	}

	partitionCapacityCmd := &cobra.Command{
		Use:   "capacity",
		Short: "show partition capacity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.partitionCapacity()
		},
	}

	partitionCapacityCmd.Flags().StringP("id", "", "", "filter on partition id. [optional]")
	partitionCapacityCmd.Flags().StringP("size", "", "", "filter on size id. [optional]")
	partitionCapacityCmd.Flags().StringSlice("sort-by", []string{}, fmt.Sprintf("order by (comma separated) column(s), sort direction can be changed by appending :asc or :desc behind the column identifier. possible values: %s", strings.Join(sorters.PartitionCapacitySorter().AvailableKeys(), "|")))
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("id", c.comp.PartitionListCompletion))
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("sort-by", cobra.FixedCompletions(sorters.PartitionCapacitySorter().AvailableKeys(), cobra.ShellCompDirectiveNoFileComp)))

	return genericcli.NewCmds(cmdsConfig, partitionCapacityCmd)
}

func (c partitionCmd) Get(id string) (*models.V1PartitionResponse, error) {
	resp, err := c.client.Partition().FindPartition(partition.NewFindPartitionParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCmd) List() ([]*models.V1PartitionResponse, error) {
	resp, err := c.client.Partition().ListPartitions(partition.NewListPartitionsParams(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCmd) Delete(id string) (*models.V1PartitionResponse, error) {
	resp, err := c.client.Partition().DeletePartition(partition.NewDeletePartitionParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCmd) Create(rq *models.V1PartitionCreateRequest) (*models.V1PartitionResponse, error) {
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

func (c partitionCmd) Update(rq *models.V1PartitionUpdateRequest) (*models.V1PartitionResponse, error) {
	resp, err := c.client.Partition().UpdatePartition(partition.NewUpdatePartitionParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c partitionCmd) ToCreate(r *models.V1PartitionResponse) (*models.V1PartitionCreateRequest, error) {
	return partitionResponseToCreate(r), nil
}

func (c partitionCmd) ToUpdate(r *models.V1PartitionResponse) (*models.V1PartitionUpdateRequest, error) {
	return partitionResponseToUpdate(r), nil
}

func partitionResponseToCreate(r *models.V1PartitionResponse) *models.V1PartitionCreateRequest {
	return &models.V1PartitionCreateRequest{
		Bootconfig: &models.V1PartitionBootConfiguration{
			Commandline: r.Bootconfig.Commandline,
			Imageurl:    r.Bootconfig.Imageurl,
			Kernelurl:   r.Bootconfig.Kernelurl,
		},
		Description:                r.Description,
		ID:                         r.ID,
		Mgmtserviceaddress:         r.Mgmtserviceaddress,
		Name:                       r.Name,
		Privatenetworkprefixlength: r.Privatenetworkprefixlength,
		Waitingpoolmaxsize:         r.Waitingpoolmaxsize,
		Waitingpoolminsize:         r.Waitingpoolminsize,
	}
}

func partitionResponseToUpdate(r *models.V1PartitionResponse) *models.V1PartitionUpdateRequest {
	return &models.V1PartitionUpdateRequest{
		Bootconfig: &models.V1PartitionBootConfiguration{
			Commandline: r.Bootconfig.Commandline,
			Imageurl:    r.Bootconfig.Imageurl,
			Kernelurl:   r.Bootconfig.Kernelurl,
		},
		Description:        r.Description,
		ID:                 r.ID,
		Mgmtserviceaddress: r.Mgmtserviceaddress,
		Name:               r.Name,
		Waitingpoolmaxsize: r.Waitingpoolmaxsize,
		Waitingpoolminsize: r.Waitingpoolminsize,
	}
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

	err = sorters.PartitionCapacitySorter().SortBy(resp.Payload)
	if err != nil {
		return err
	}

	for _, pc := range resp.Payload {
		pc := pc
		sort.SliceStable(pc.Servers, func(i, j int) bool {
			return pointer.SafeDeref(pointer.SafeDeref(pc.Servers[i]).Size) < pointer.SafeDeref(pointer.SafeDeref(pc.Servers[j]).Size)
		})
	}

	return c.listPrinter.Print(resp.Payload)
}
