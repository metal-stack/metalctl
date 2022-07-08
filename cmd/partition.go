package cmd

import (
	"errors"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/partition"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type partitionCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	gcli   *genericcli.GenericCLI[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse]
}

func newPartitionCmd(c *config) *cobra.Command {
	w := partitionCmd{
		c:      c.client,
		driver: c.driver,
		gcli:   genericcli.NewGenericCLI[*models.V1PartitionCreateRequest, *models.V1PartitionUpdateRequest, *models.V1PartitionResponse](partitionGeneric{c: c.client}),
	}

	partitionCmd := &cobra.Command{
		Use:   "partition",
		Short: "manage partitions",
		Long:  "a partition is a group of machines and network which is logically separated from other partitions. Machines have no direct network connections between partitions.",
	}

	partitionListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all partitions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.partitionList()
		},
		PreRun: bindPFlags,
	}
	partitionCapacityCmd := &cobra.Command{
		Use:   "capacity",
		Short: "show partition capacity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.partitionCapacity()
		},
		PreRun: bindPFlags,
	}
	partitionDescribeCmd := &cobra.Command{
		Use:   "describe <partitionID>",
		Short: "describe a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.DescribeAndPrint(args, genericcli.NewYAMLPrinter())
		},
		ValidArgsFunction: c.comp.PartitionListCompletion,
	}
	partitionCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.IsSet("file") {
				return w.gcli.CreateFromFileAndPrint(viper.GetString("file"), genericcli.NewYAMLPrinter())
			}

			return w.gcli.CreateAndPrint(&models.V1PartitionCreateRequest{
				ID:                 pointer.Pointer(viper.GetString("id")),
				Description:        viper.GetString("description"),
				Name:               viper.GetString("name"),
				Mgmtserviceaddress: viper.GetString("mgmtserver"),
				Bootconfig: &models.V1PartitionBootConfiguration{
					Commandline: viper.GetString("cmdline"),
					Imageurl:    viper.GetString("imageurl"),
					Kernelurl:   viper.GetString("kernelurl"),
				},
			}, genericcli.NewYAMLPrinter())
		},
		PreRun: bindPFlags,
	}
	partitionUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.UpdateFromFileAndPrint(viper.GetString("file"), genericcli.NewYAMLPrinter())
		},
		PreRun: bindPFlags,
	}
	partitionApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.ApplyFromFileAndPrint(viper.GetString("file"), output.New())
		},
		PreRun: bindPFlags,
	}
	partitionDeleteCmd := &cobra.Command{
		Use:     "delete <partitionID>",
		Short:   "delete a partition",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.DeleteAndPrint(args, genericcli.NewYAMLPrinter())
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.PartitionListCompletion,
	}
	partitionEditCmd := &cobra.Command{
		Use:   "edit <partitionID>",
		Short: "edit a partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.EditAndPrint(args, genericcli.NewYAMLPrinter())
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.PartitionListCompletion,
	}

	partitionCreateCmd.Flags().StringP("id", "", "", "ID of the partition. [required]")
	partitionCreateCmd.Flags().StringP("name", "n", "", "Name of the partition. [optional]")
	partitionCreateCmd.Flags().StringP("description", "d", "", "Description of the partition. [required]")
	partitionCreateCmd.Flags().StringP("mgmtserver", "", "", "management server address in the partition. [required]")
	partitionCreateCmd.Flags().StringP("cmdline", "", "", "kernel commandline for the metal-hammer in the partition. [required]")
	partitionCreateCmd.Flags().StringP("imageurl", "", "", "initrd for the metal-hammer in the partition. [required]")
	partitionCreateCmd.Flags().StringP("kernelurl", "", "", "kernel url for the metal-hammer in the partition. [required]")

	partitionApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl partition describe partition-a > a.yaml
# vi a.yaml
## either via stdin
# cat a.yaml | metalctl partition apply -f -
## or via file
# metalctl partition apply -f a.yaml`)
	must(partitionApplyCmd.MarkFlagRequired("file"))

	partitionUpdateCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl partition describe partition-a > a.yaml
# vi a.yaml
## either via stdin
# cat a.yaml | metalctl partition update -f -
## or via file
# metalctl partition update -f a.yaml`)
	must(partitionUpdateCmd.MarkFlagRequired("file"))

	partitionCapacityCmd.Flags().StringP("id", "", "", "filter on partition id. [optional]")
	partitionCapacityCmd.Flags().StringP("size", "", "", "filter on size id. [optional]")
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("id", c.comp.PartitionListCompletion))
	must(partitionCapacityCmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))

	partitionCmd.AddCommand(partitionListCmd)
	partitionCmd.AddCommand(partitionCapacityCmd)
	partitionCmd.AddCommand(partitionDescribeCmd)
	partitionCmd.AddCommand(partitionCreateCmd)
	partitionCmd.AddCommand(partitionUpdateCmd)
	partitionCmd.AddCommand(partitionApplyCmd)
	partitionCmd.AddCommand(partitionDeleteCmd)
	partitionCmd.AddCommand(partitionEditCmd)

	return partitionCmd
}

type partitionGeneric struct {
	c metalgo.Client
}

func (g partitionGeneric) Get(id string) (*models.V1PartitionResponse, error) {
	resp, err := g.c.Partition().FindPartition(partition.NewFindPartitionParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (g partitionGeneric) Delete(id string) (*models.V1PartitionResponse, error) {
	resp, err := g.c.Partition().DeletePartition(partition.NewDeletePartitionParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (g partitionGeneric) Create(rq *models.V1PartitionCreateRequest) (*models.V1PartitionResponse, error) {
	resp, err := g.c.Partition().CreatePartition(partition.NewCreatePartitionParams().WithBody(rq), nil)
	if err != nil {
		var r *partition.CreatePartitionConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (g partitionGeneric) Update(rq *models.V1PartitionUpdateRequest) (*models.V1PartitionResponse, error) {
	resp, err := g.c.Partition().UpdatePartition(partition.NewUpdatePartitionParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *config) partitionList() error {
	resp, err := c.driver.PartitionList()
	if err != nil {
		return err
	}
	return output.New().Print(resp.Partition)
}

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
