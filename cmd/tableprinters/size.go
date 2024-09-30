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
		header = []string{"ID", "Name", "Description", "CPU Range", "Memory Range", "Storage Range", "GPU Range"}
		rows   [][]string
	)

	if wide {
		header = []string{"ID", "Name", "Description", "CPU Range", "Memory Range", "Storage Range", "GPU Range", "Labels"}
	}

	for _, size := range data {
		var cpu, memory, storage, gpu string
		for _, c := range size.Constraints {
			switch *c.Type {
			case models.V1SizeConstraintTypeCores:
				cpu = fmt.Sprintf("%d - %d", c.Min, c.Max)
			case models.V1SizeConstraintTypeMemory:
				memory = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(c.Min)), humanize.Bytes(uint64(c.Max))) //nolint:gosec
			case models.V1SizeConstraintTypeStorage:
				storage = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(c.Min)), humanize.Bytes(uint64(c.Max))) //nolint:gosec
			case models.V1SizeConstraintTypeGpu:
				gpu = fmt.Sprintf("%s: %d - %d", c.Identifier, c.Min, c.Max)
			}

		}

		row := []string{pointer.SafeDeref(size.ID), size.Name, size.Description, cpu, memory, storage, gpu}

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
		header = []string{"ID", "Size", "Project", "Partitions", "Description", "Amount"}
		rows   [][]string
	)

	if wide {
		header = append(header, "Labels")
	}

	for _, d := range data {
		d := d

		desc := d.Description
		if !wide {
			desc = genericcli.TruncateEnd(d.Description, 50)
		}

		row := []string{
			pointer.SafeDeref(d.ID),
			pointer.SafeDeref(d.Sizeid),
			pointer.SafeDeref(d.Projectid),
			strings.Join(d.Partitionids, ", "),
			desc,
			fmt.Sprintf("%d", pointer.SafeDeref(d.Amount)),
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

func (t *TablePrinter) SizeReservationUsageTable(data []*models.V1SizeReservationUsageResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Size", "Project", "Partition", "Used/Amount"}
		rows   [][]string
	)

	if wide {
		header = append(header, "Allocated", "Labels")
	}

	for _, d := range data {
		d := d

		row := []string{
			pointer.SafeDeref(d.ID),
			pointer.SafeDeref(d.Sizeid),
			pointer.SafeDeref(d.Projectid),
			pointer.SafeDeref(d.Partitionid),
			fmt.Sprintf("%d/%d", pointer.SafeDeref(d.Usedamount), pointer.SafeDeref(d.Amount)),
		}

		if wide {
			labels := genericcli.MapToLabels(d.Labels)
			sort.Strings(labels)
			row = append(row,
				strconv.Itoa(int(pointer.SafeDeref(d.Projectallocations))),
				strings.Join(labels, "\n"),
			)
		}

		rows = append(rows, row)
	}

	return header, rows, nil
}
