package cmd

import (
	"errors"
	"fmt"

	sizemodel "github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeReservationsCmd struct {
	*config
}

func newSizeReservationsCmd(c *config) *cobra.Command {
	w := &sizeReservationsCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1SizeReservationCreateRequest, *models.V1SizeReservationUpdateRequest, *models.V1SizeReservationResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI(w).WithFS(c.fs),
		Singular:        "reservation",
		Plural:          "reservations",
		Description:     "manage size reservations",
		Aliases:         []string{"rs"},
		Sorter:          sorters.SizeReservationsSorter(),
		ValidArgsFn:     c.comp.SizeReservationsListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		CreateRequestFromCLI: func() (*models.V1SizeReservationCreateRequest, error) {
			labels, err := genericcli.LabelsToMap(viper.GetStringSlice("labels"))
			if err != nil {
				return nil, err
			}

			return &models.V1SizeReservationCreateRequest{
				Amount:       pointer.PointerOrNil(int32(viper.GetInt32("amount"))),
				Description:  viper.GetString("description"),
				ID:           pointer.PointerOrNil(viper.GetString("id")),
				Labels:       labels,
				Partitionids: viper.GetStringSlice("partitions"),
				Projectid:    pointer.PointerOrNil(viper.GetString("project")),
				Sizeid:       pointer.PointerOrNil(viper.GetString("size")),
			}, nil
		},
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().Int32("amount", 0, "the amount to associate with this reservation")
			cmd.Flags().String("id", "", "the id to associate with this reservation")
			cmd.Flags().String("size", "", "the size id to associate with this reservation")
			cmd.Flags().String("project", "", "the project id to associate with this reservation")
			cmd.Flags().StringSlice("partitions", nil, "the partition ids to associate with this reservation")
			cmd.Flags().StringSlice("labels", nil, "the labels to associate with this reservation")
			cmd.Flags().String("description", "", "the description to associate with this reservation")

			genericcli.Must(cmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("partitions", c.comp.PartitionListCompletion))
		},
		UpdateRequestFromCLI: func(args []string) (*models.V1SizeReservationUpdateRequest, error) {
			id, err := genericcli.GetExactlyOneArg(args)
			if err != nil {
				return nil, err
			}

			labels, err := genericcli.LabelsToMap(viper.GetStringSlice("labels"))
			if err != nil {
				return nil, err
			}

			return &models.V1SizeReservationUpdateRequest{ //nolint:exhaustruct
				Amount:       pointer.PointerOrNil(int32(viper.GetInt32("amount"))),
				Description:  viper.GetString("description"),
				ID:           &id,
				Labels:       labels,
				Partitionids: viper.GetStringSlice("partitions"),
			}, nil
		},
		UpdateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().Int32("amount", 0, "the amount to associate with this reservation")
			cmd.Flags().StringSlice("partitions", nil, "the partition ids to associate with this reservation")
			cmd.Flags().StringSlice("labels", nil, "the labels to associate with this reservation")
			cmd.Flags().String("description", "", "the description to associate with this reservation")

			genericcli.Must(cmd.RegisterFlagCompletionFunc("partitions", c.comp.PartitionListCompletion))
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "the id to filter")
			cmd.Flags().String("size", "", "the size id to filter")
			cmd.Flags().String("project", "", "the project id to filter")
			cmd.Flags().String("partition", "", "the partition id to filter")

			genericcli.Must(cmd.RegisterFlagCompletionFunc("id", c.comp.SizeReservationsListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("size", c.comp.SizeListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))
		},
	}

	usageCmd := &cobra.Command{
		Use:   "usage",
		Short: "see current usage of size reservations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.usage()
		},
	}

	usageCmd.Flags().String("size-id", "", "the size-id to filter")
	usageCmd.Flags().String("project", "", "the project to filter")
	usageCmd.Flags().String("partition", "", "the partition to filter")

	genericcli.Must(usageCmd.RegisterFlagCompletionFunc("size-id", c.comp.SizeListCompletion))
	genericcli.Must(usageCmd.RegisterFlagCompletionFunc("project", c.comp.ProjectListCompletion))
	genericcli.Must(usageCmd.RegisterFlagCompletionFunc("partition", c.comp.PartitionListCompletion))

	genericcli.AddSortFlag(usageCmd, sorters.SizeReservationsUsageSorter())

	return genericcli.NewCmds(cmdsConfig, usageCmd)
}

func (c *sizeReservationsCmd) Get(id string) (*models.V1SizeReservationResponse, error) {
	resp, err := c.client.Size().GetSizeReservation(sizemodel.NewGetSizeReservationParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeReservationsCmd) List() ([]*models.V1SizeReservationResponse, error) {
	resp, err := c.client.Size().FindSizeReservations(sizemodel.NewFindSizeReservationsParams().WithBody(&models.V1SizeReservationListRequest{
		ID:          viper.GetString("id"),
		Partitionid: viper.GetString("partition"),
		Projectid:   viper.GetString("project"),
		Sizeid:      viper.GetString("size"),
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeReservationsCmd) Delete(id string) (*models.V1SizeReservationResponse, error) {
	resp, err := c.client.Size().DeleteSizeReservation(sizemodel.NewDeleteSizeReservationParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeReservationsCmd) Create(rq *models.V1SizeReservationCreateRequest) (*models.V1SizeReservationResponse, error) {
	resp, err := c.client.Size().CreateSizeReservation(sizemodel.NewCreateSizeReservationParams().WithBody(rq), nil)
	if err != nil {
		var r *sizemodel.CreateSizeReservationConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeReservationsCmd) Update(rq *models.V1SizeReservationUpdateRequest) (*models.V1SizeReservationResponse, error) {
	resp, err := c.client.Size().UpdateSizeReservation(sizemodel.NewUpdateSizeReservationParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeReservationsCmd) Convert(r *models.V1SizeReservationResponse) (string, *models.V1SizeReservationCreateRequest, *models.V1SizeReservationUpdateRequest, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.ID, sizeReservationResponseToCreate(r), sizeReservationResponseToUpdate(r), nil
}

func sizeReservationResponseToCreate(r *models.V1SizeReservationResponse) *models.V1SizeReservationCreateRequest {
	return &models.V1SizeReservationCreateRequest{
		Amount:       r.Amount,
		Description:  r.Description,
		ID:           r.ID,
		Labels:       r.Labels,
		Name:         r.Name,
		Partitionids: r.Partitionids,
		Projectid:    r.Projectid,
		Sizeid:       r.Sizeid,
	}
}

func sizeReservationResponseToUpdate(r *models.V1SizeReservationResponse) *models.V1SizeReservationUpdateRequest {
	return &models.V1SizeReservationUpdateRequest{
		Amount:       r.Amount,
		Description:  r.Description,
		ID:           r.ID,
		Labels:       r.Labels,
		Name:         r.Name,
		Partitionids: r.Partitionids,
	}
}

// non-generic command handling

func (c *sizeReservationsCmd) usage() error {
	sortKeys, err := genericcli.ParseSortFlags()
	if err != nil {
		return err
	}

	resp, err := c.client.Size().SizeReservationsUsage(sizemodel.NewSizeReservationsUsageParams().WithBody(&models.V1SizeReservationListRequest{
		Partitionid: viper.GetString("partition"),
		Projectid:   viper.GetString("project"),
		Sizeid:      viper.GetString("size"),
	}), nil)
	if err != nil {
		return err
	}

	err = sorters.SizeReservationsUsageSorter().SortBy(resp.Payload, sortKeys...)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(resp.Payload)
}
