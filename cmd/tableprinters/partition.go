package tableprinters

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func (t *TablePrinter) PartitionTable(data []*models.V1PartitionResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Name", "Description", "MinWait", "MaxWait"}
		rows   [][]string
	)

	if wide {
		header = []string{"ID", "Name", "Description", "Labels"}
	}

	for _, p := range data {
		row := []string{pointer.SafeDeref(p.ID), p.Name, p.Description, p.Waitingpoolminsize, p.Waitingpoolmaxsize}

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
		header = []string{"Partition", "Size", "Allocated", "Free", "Unavailable", "Reservations", "|", "Total", "|", "Faulty"}
		rows   [][]string

		allocatedCount       int32
		faultyCount          int32
		freeCount            int32
		otherCount           int32
		phonedHomeCount      int32
		reservationCount     int32
		totalCount           int32
		unavailableCount     int32
		usedReservationCount int32
		waitingCount         int32
	)

	if wide {
		header = append(header, "Phoned Home", "Waiting", "Other")
	}

	for _, pc := range data {
		pc := pc

		for _, c := range pc.Servers {
			id := pointer.SafeDeref(c.Size)

			var (
				allocated    = fmt.Sprintf("%d", c.Allocated)
				faulty       = fmt.Sprintf("%d", c.Faulty)
				free         = fmt.Sprintf("%d", c.Free)
				other        = fmt.Sprintf("%d", c.Other)
				phonedHome   = fmt.Sprintf("%d", c.PhonedHome)
				reservations = "0"
				total        = fmt.Sprintf("%d", c.Total)
				unavailable  = fmt.Sprintf("%d", c.Unavailable)
				waiting      = fmt.Sprintf("%d", c.Waiting)
			)

			if c.Reservations > 0 {
				reservations = fmt.Sprintf("%d (%d/%d used)", c.Reservations-c.Usedreservations, c.Usedreservations, c.Reservations)
			}

			allocatedCount += c.Allocated
			faultyCount += c.Faulty
			freeCount += c.Free
			otherCount += c.Other
			phonedHomeCount += c.PhonedHome
			reservationCount += c.Reservations
			totalCount += c.Total
			unavailableCount += c.Unavailable
			usedReservationCount += c.Usedreservations
			waitingCount += c.Waiting

			row := []string{*pc.ID, id, allocated, free, unavailable, reservations, "|", total, "|", faulty}
			if wide {
				row = append(row, phonedHome, waiting, other)
			}

			rows = append(rows, row)
		}
	}

	footerRow := ([]string{
		"Total",
		"",
		fmt.Sprintf("%d", allocatedCount),
		fmt.Sprintf("%d", freeCount),
		fmt.Sprintf("%d", unavailableCount),
		fmt.Sprintf("%d", reservationCount-usedReservationCount),
		"|",
		fmt.Sprintf("%d", totalCount),
		"|",
		fmt.Sprintf("%d", faultyCount),
	})

	if wide {
		footerRow = append(footerRow, []string{
			fmt.Sprintf("%d", phonedHomeCount),
			fmt.Sprintf("%d", waitingCount),
			fmt.Sprintf("%d", otherCount),
		}...)
	}

	if t.markdown {
		// for markdown we already have enough dividers, remove them
		removeDivider := func(e string) bool {
			return e == "|"
		}
		header = slices.DeleteFunc(header, removeDivider)
		footerRow = slices.DeleteFunc(footerRow, removeDivider)
		for i, row := range rows {
			rows[i] = slices.DeleteFunc(row, removeDivider)
		}
	}

	rows = append(rows, footerRow)

	return header, rows, nil
}
