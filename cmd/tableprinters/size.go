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
		header = []string{"ID", "Name", "Description", "Reservations", "CPU Range", "Memory Range", "Storage Range", "GPU Range"}
		rows   [][]string
	)

	if wide {
		header = []string{"ID", "Name", "Description", "Reservations", "CPU Range", "Memory Range", "Storage Range", "GPU Range", "Labels"}
	}

	for _, size := range data {
		var cpu, memory, storage, gpu string
		for _, c := range size.Constraints {
			switch *c.Type {
			case models.V1SizeConstraintTypeCores:
				cpu = fmt.Sprintf("%d - %d", c.Min, c.Max)
			case models.V1SizeConstraintTypeMemory:
				memory = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(c.Min)), humanize.Bytes(uint64(c.Max)))
			case models.V1SizeConstraintTypeStorage:
				storage = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(c.Min)), humanize.Bytes(uint64(c.Max)))
			case models.V1SizeConstraintTypeGpu:
				gpu = fmt.Sprintf("%s: %d - %d", c.Identifier, c.Min, c.Max)
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

func (t *TablePrinter) SizeReservationTable(data []*models.V1SizeReservationResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Partition", "Size", "Tenant", "Project", "Project Name", "Used/Amount", "Project Allocations"}
		rows   [][]string
	)

	if wide {
		header = append(header, "Labels")
	}

	for _, d := range data {
		d := d

		row := []string{
			pointer.SafeDeref(d.Partitionid),
			pointer.SafeDeref(d.Sizeid),
			pointer.SafeDeref(d.Tenant),
			pointer.SafeDeref(d.Projectid),
			pointer.SafeDeref(d.Projectname),
			fmt.Sprintf("%d/%d", pointer.SafeDeref(d.Usedreservations), pointer.SafeDeref(d.Reservations)),
			strconv.Itoa(int(pointer.SafeDeref(d.Projectallocations))),
		}

		if wide {
			labels := genericcli.MapToLabels(d.Labels)
			sort.Strings(labels)
			row = append(row, strings.Join(labels, "\n"))
		}

		rows = append(rows, row)
	}

	return header, rows, nil
}
