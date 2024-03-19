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

func (t *TablePrinter) SizeTable(data []*models.V1SizeResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Name", "Description", "Reservations", "CPU Range", "Memory Range", "Storage Range", "GPUs"}
		rows   [][]string
	)

	if wide {
		header = []string{"ID", "Name", "Description", "Reservations", "CPU Range", "Memory Range", "Storage Range", "GPUs", "Labels"}
	}

	for _, size := range data {
		var cpu, memory, storage, gpu string
		for _, c := range size.Constraints {
			switch *c.Type {
			case models.V1SizeConstraintTypeCores:
				cpu = fmt.Sprintf("%d - %d", *c.Min, *c.Max)
			case models.V1SizeConstraintTypeMemory:
				memory = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)))
			case models.V1SizeConstraintTypeStorage:
				storage = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)))
			case models.V1SizeConstraintTypeGpu:
				gpu = fmt.Sprintf("%v", c.Gpus)
			}

		}

		reservationCount := 0
		for _, r := range size.Reservations {
			r := r
			reservationCount += int(pointer.SafeDeref(r.Amount))
		}

		row := []string{pointer.SafeDeref(size.ID), size.Name, size.Description, strconv.Itoa(reservationCount), cpu, memory, storage, gpu}

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

func (t *TablePrinter) SizeReservationTable(data []*models.V1SizeReservationResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Partition", "Size", "Tenant", "Project", "Project Name", "Used/Amount", "Project Allocations"}
		rows   [][]string
	)

	for _, d := range data {
		d := d

		rows = append(rows, []string{
			pointer.SafeDeref(d.Partitionid),
			pointer.SafeDeref(d.Sizeid),
			pointer.SafeDeref(d.Tenant),
			pointer.SafeDeref(d.Projectid),
			pointer.SafeDeref(d.Projectname),
			fmt.Sprintf("%d/%d", pointer.SafeDeref(d.Usedreservations), pointer.SafeDeref(d.Reservations)),
			strconv.Itoa(int(pointer.SafeDeref(d.Projectallocations))),
		})
	}

	return header, rows, nil
}
