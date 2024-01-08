package tableprinters

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/olekukonko/tablewriter"
)

type SizeReservations struct {
	Projects []*models.V1ProjectResponse
	Sizes    []*models.V1SizeResponse
	Machines []*models.V1MachineResponse
}

func (t *TablePrinter) SizeTable(data []*models.V1SizeResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Name", "Description", "Reservations", "CPU Range", "Memory Range", "Storage Range"}
		rows   [][]string
	)

	if wide {
		header = []string{"ID", "Name", "Description", "Reservations", "CPU Range", "Memory Range", "Storage Range", "Labels"}
	}

	for _, size := range data {
		var cpu, memory, storage string
		for _, c := range size.Constraints {
			switch *c.Type {
			case "cores":
				cpu = fmt.Sprintf("%d - %d", *c.Min, *c.Max)
			case "memory":
				memory = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)))
			case "storage":
				storage = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)))
			}
		}

		reservationCount := 0
		for _, r := range size.Reservations {
			r := r
			reservationCount += int(pointer.SafeDeref(r.Amount))
		}

		row := []string{pointer.SafeDeref(size.ID), size.Name, size.Description, strconv.Itoa(reservationCount), cpu, memory, storage}

		if wide {
			labels := genericcli.MapToLabels(size.Labels)
			sort.Strings(labels)
			row = append(row, strings.Join(labels, "\n"))
		}

		rows = append(rows, row)
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
	})

	return header, rows, nil
}

func (t *TablePrinter) SizeMatchingLogTable(data []*models.V1SizeMatchingLog, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Name", "Match", "CPU Constraint", "Memory Constraint", "Storage Constraint"}
		rows   [][]string
	)

	for _, d := range data {
		var cpu, memory, storage string
		for _, cs := range d.Constraints {
			c := cs.Constraint
			switch *c.Type {
			case "cores": // TODO: should be enums in spec
				cpu = fmt.Sprintf("%d - %d\n%s\nmatches: %v", *c.Min, *c.Max, *cs.Log, *cs.Match)
			case "memory":
				memory = fmt.Sprintf("%s - %s\n%s\nmatches: %v", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)), *cs.Log, *cs.Match)
			case "storage":
				storage = fmt.Sprintf("%s - %s\n%s\nmatches: %v", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)), *cs.Log, *cs.Match)
			}
		}
		sizeMatch := fmt.Sprintf("%v", *d.Match)

		rows = append(rows, []string{*d.Name, sizeMatch, cpu, memory, storage})
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
		table.SetColMinWidth(3, 40)
	})

	return header, rows, nil
}

func (t *TablePrinter) SizeReservationTable(data *SizeReservations, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Partition", "Tenant", "Project", "Project Name", "Used/Amount", "Total Allocations"}
		rows   [][]string
	)

	projectsByID := ProjectsByID(data.Projects)
	machinesByProject := map[string][]*models.V1MachineResponse{}
	for _, m := range data.Machines {
		m := m
		if m.Allocation == nil || m.Allocation.Project == nil {
			continue
		}

		machinesByProject[*m.Allocation.Project] = append(machinesByProject[*m.Allocation.Project], m)
	}

	for _, d := range data.Sizes {
		d := d
		for _, reservation := range d.Reservations {
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

				projectMachineCount := len(machinesByProject[*reservation.Projectid])
				maxReservationCount := int(pointer.SafeDeref(reservation.Amount))

				rows = append(rows, []string{
					partitionID,
					projectName,
					*reservation.Projectid,
					tenant,
					fmt.Sprintf("%d/%d", min(maxReservationCount, projectMachineCount), maxReservationCount),
					strconv.Itoa(projectMachineCount),
				})
			}
		}
	}

	return header, rows, nil
}

func ProjectsByID(projects []*models.V1ProjectResponse) map[string]*models.V1ProjectResponse {
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
