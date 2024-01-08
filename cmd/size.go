package cmd

import (
	"errors"
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/client/partition"
	"github.com/metal-stack/metal-go/api/client/project"
	"github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeCmd struct {
	*config
}

func newSizeCmd(c *config) *cobra.Command {
	w := sizeCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse](w).WithFS(c.fs),
		Singular:        "size",
		Plural:          "sizes",
		Description:     "a size matches a machine in terms of cpu cores, ram and storage.",
		Sorter:          sorters.SizeSorter(),
		ValidArgsFn:     c.comp.SizeListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		CreateRequestFromCLI: func() (*models.V1SizeCreateRequest, error) {
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
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "", "", "ID of the size. [required]")
			cmd.Flags().StringP("name", "n", "", "Name of the size. [optional]")
			cmd.Flags().StringP("description", "d", "", "Description of the size. [required]")
			// FIXME constraints must be given in a slice
			cmd.Flags().Int64P("min", "", 0, "min value of given size constraint type. [required]")
			cmd.Flags().Int64P("max", "", 0, "min value of given size constraint type. [required]")
			cmd.Flags().StringP("type", "", "", "type of constraints. [required]")
		},
	}

	tryCmd := &cobra.Command{
		Use:   "try",
		Short: "try a specific hardware spec and give the chosen size back",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.try()
		},
	}

	tryCmd.Flags().Int32P("cores", "C", 0, "Cores of the hardware to try")
	tryCmd.Flags().StringP("memory", "M", "", "Memory of the hardware to try, can be given in bytes or any human readable size spec")
	tryCmd.Flags().StringP("storagesize", "S", "", "Total storagesize of the hardware to try, can be given in bytes or any human readable size spec")

	reservationsCmd := &cobra.Command{
		Use:   "reservations",
		Short: "manage size reservations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.listReverations()
		},
	}

	listReservationsCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list size reservations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.listReverations()
		},
	}
	checkReservationsCmd := &cobra.Command{
		Use:   "check",
		Short: "check if there are size reservations that are ineffective, e.g. because a project was deleted",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.checkReverations()
		},
	}

	genericcli.AddSortFlag(listReservationsCmd, sorters.SizeReservationsSorter())

	reservationsCmd.AddCommand(listReservationsCmd, checkReservationsCmd)

	return genericcli.NewCmds(cmdsConfig, newSizeImageConstraintCmd(c), tryCmd, reservationsCmd)
}

func (c sizeCmd) Get(id string) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().FindSize(size.NewFindSizeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeCmd) List() ([]*models.V1SizeResponse, error) {
	resp, err := c.client.Size().ListSizes(size.NewListSizesParams(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeCmd) Delete(id string) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().DeleteSize(size.NewDeleteSizeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeCmd) Create(rq *models.V1SizeCreateRequest) (*models.V1SizeResponse, error) {
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

func (c sizeCmd) Update(rq *models.V1SizeUpdateRequest) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().UpdateSize(size.NewUpdateSizeParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeCmd) Convert(r *models.V1SizeResponse) (string, *models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.ID, sizeResponseToCreate(r), sizeResponseToUpdate(r), nil
}

func sizeResponseToCreate(r *models.V1SizeResponse) *models.V1SizeCreateRequest {
	var constraints []*models.V1SizeConstraint
	for i := range r.Constraints {
		constraints = append(constraints, &models.V1SizeConstraint{
			Max:  r.Constraints[i].Max,
			Min:  r.Constraints[i].Min,
			Type: r.Constraints[i].Type,
		})
	}
	return &models.V1SizeCreateRequest{
		Constraints: constraints,
		Description: r.Description,
		ID:          r.ID,
		Name:        r.Name,
	}
}

func sizeResponseToUpdate(r *models.V1SizeResponse) *models.V1SizeUpdateRequest {
	var constraints []*models.V1SizeConstraint
	for i := range r.Constraints {
		constraints = append(constraints, &models.V1SizeConstraint{
			Max:  r.Constraints[i].Max,
			Min:  r.Constraints[i].Min,
			Type: r.Constraints[i].Type,
		})
	}
	return &models.V1SizeUpdateRequest{
		Constraints:  constraints,
		Description:  r.Description,
		ID:           r.ID,
		Name:         r.Name,
		Labels:       r.Labels,
		Reservations: r.Reservations,
	}
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

	return c.listPrinter.Print(resp.Payload)
}

func (c sizeCmd) listReverations() error {
	sizes, err := c.client.Size().ListSizes(size.NewListSizesParams(), nil)
	if err != nil {
		return err
	}

	projects, err := c.client.Project().ListProjects(project.NewListProjectsParams(), nil)
	if err != nil {
		return err
	}

	machines, err := c.client.Machine().ListMachines(machine.NewListMachinesParams(), nil)
	if err != nil {
		return err
	}

	var (
		sizeReservations        []*tableprinters.SizeReservation
		projectsByID            = projectsByID(projects.Payload)
		machinesByProjectBySize = map[string]map[string][]*models.V1MachineResponse{}
	)

	for _, m := range machines.Payload {
		m := m
		if m.Allocation == nil || m.Allocation.Project == nil || m.Size == nil || m.Size.ID == nil {
			continue
		}

		byProject, ok := machinesByProjectBySize[*m.Allocation.Project]
		if !ok {
			byProject = map[string][]*models.V1MachineResponse{}
		}

		byProject[*m.Size.ID] = append(byProject[*m.Size.ID], m)

		machinesByProjectBySize[*m.Allocation.Project] = byProject
	}

	for _, size := range sizes.Payload {
		size := size
		for _, reservation := range size.Reservations {
			if reservation.Projectid == nil {
				continue
			}

			for _, partitionID := range reservation.Partitionids {
				var (
					projectName string
					tenant      string
				)

				project, ok := projectsByID[*reservation.Projectid]
				if ok {
					projectName = project.Name
					tenant = project.TenantID
				}

				projectMachineCount := len(machinesByProjectBySize[*reservation.Projectid][*size.ID])
				maxReservationCount := int(pointer.SafeDeref(reservation.Amount))

				sizeReservations = append(sizeReservations, &tableprinters.SizeReservation{
					Size:               *size.ID,
					Partition:          partitionID,
					Tenant:             tenant,
					ProjectID:          *reservation.Projectid,
					ProjectName:        projectName,
					Reservations:       maxReservationCount,
					UsedReservations:   min(maxReservationCount, projectMachineCount),
					ProjectAllocations: projectMachineCount,
				})
			}
		}
	}

	err = sorters.SizeReservationsSorter().SortBy(sizeReservations)
	if err != nil {
		return err
	}

	return c.listPrinter.Print(sizeReservations)
}

func (c sizeCmd) checkReverations() error {
	sizes, err := c.client.Size().ListSizes(size.NewListSizesParams(), nil)
	if err != nil {
		return err
	}

	projects, err := c.client.Project().ListProjects(project.NewListProjectsParams(), nil)
	if err != nil {
		return err
	}

	partitions, err := c.client.Partition().ListPartitions(partition.NewListPartitionsParams(), nil)
	if err != nil {
		return err
	}

	var (
		errs []error

		projectsByID   = projectsByID(projects.Payload)
		partitionsByID = map[string]*models.V1PartitionResponse{}
	)

	for _, p := range partitions.Payload {
		p := p
		partitionsByID[*p.ID] = p
	}

	for _, size := range sizes.Payload {
		size := size

		for _, reservation := range size.Reservations {
			reservation := reservation

			var (
				sizeID    = pointer.SafeDeref(size.ID)
				projectID = pointer.SafeDeref(reservation.Projectid)
			)

			for _, partition := range reservation.Partitionids {
				_, ok := partitionsByID[partition]
				if !ok {
					errs = append(errs, fmt.Errorf("size reservation for size %q and project %q references a non-existing partition %q", sizeID, projectID, partition))
				}
			}

			_, ok := projectsByID[projectID]
			if !ok {
				errs = append(errs, fmt.Errorf("size reservation for size %q references a non-existing project %q", sizeID, projectID))
			}
		}
	}

	if len(errs) == 0 {
		fmt.Fprintln(c.out, "all size reservations are effective")
		return nil
	} else {
		for _, err := range errs {
			fmt.Fprintln(c.out, "found ineffective size reservations:")
			fmt.Fprintln(c.out, err.Error())
		}
	}

	return nil
}

func projectsByID(projects []*models.V1ProjectResponse) map[string]*models.V1ProjectResponse {
	projectsByID := map[string]*models.V1ProjectResponse{}

	for _, project := range projects {
		project := project

		if project.Meta == nil {
			continue
		}

		projectsByID[project.Meta.ID] = project
	}

	return projectsByID
}
