package tableprinters

import (
	"fmt"
	"sort"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func (t *TablePrinter) PartitionTable(data []*models.V1PartitionResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Name", "Description"}
		rows   [][]string
	)

	if wide {
		header = []string{"ID", "Name", "Description", "Labels"}
	}

	for _, p := range data {
		row := []string{pointer.SafeDeref(p.ID), p.Name, p.Description}

		if wide {
			labels := genericcli.MapToLabels(p.Labels)
			sort.Strings(labels)
			row = append(row, strings.Join(labels, "\n"))
		}

		rows = append(rows, row)
	}

	return header, rows, nil
}

func (t *TablePrinter) PartitionCapacityTable(data []*models.V1PartitionCapacity, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Partition", "Size", "Total", "Free", "Allocated", "Reserved", "Other", "Faulty"}
		rows   [][]string

		totalCount           int32
		freeCount            int32
		allocatedCount       int32
		faultyCount          int32
		otherCount           int32
		reservationCount     int32
		usedReservationCount int32
	)

	for _, pc := range data {
		pc := pc

		for _, c := range pc.Servers {
			id := pointer.SafeDeref(c.Size)

			var (
				allocated    = fmt.Sprintf("%d", pointer.SafeDeref(c.Allocated))
				total        = fmt.Sprintf("%d", pointer.SafeDeref(c.Total))
				free         = fmt.Sprintf("%d", pointer.SafeDeref(c.Free))
				faulty       = fmt.Sprintf("%d", pointer.SafeDeref(c.Faulty))
				other        = fmt.Sprintf("%d", pointer.SafeDeref(c.Other))
				reservations = fmt.Sprintf("%d/%d", pointer.SafeDeref(c.Usedreservations), pointer.SafeDeref(c.Reservations))
			)

			if wide {
				if len(c.Faultymachines) > 0 {
					faulty = strings.Join(c.Faultymachines, "\n")
				}
				if len(c.Othermachines) > 0 {
					other = strings.Join(c.Othermachines, "\n")
				}
			}

			totalCount += pointer.SafeDeref(c.Total)
			freeCount += pointer.SafeDeref(c.Free)
			allocatedCount += pointer.SafeDeref(c.Allocated)
			otherCount += pointer.SafeDeref(c.Other)
			faultyCount += pointer.SafeDeref(c.Faulty)
			reservationCount += pointer.SafeDeref(c.Reservations)
			usedReservationCount += pointer.SafeDeref(c.Usedreservations)

			rows = append(rows, []string{*pc.ID, id, total, free, allocated, reservations, other, faulty})
		}
	}

	footerRow := ([]string{
		"Total",
		"",
		fmt.Sprintf("%d", totalCount),
		fmt.Sprintf("%d", freeCount),
		fmt.Sprintf("%d", allocatedCount),
		fmt.Sprintf("%d/%d", usedReservationCount, reservationCount),
		fmt.Sprintf("%d", otherCount),
		fmt.Sprintf("%d", faultyCount),
	})
	rows = append(rows, footerRow)

	return header, rows, nil
}
