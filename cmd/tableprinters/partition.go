package tableprinters

import (
	"fmt"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func (t *TablePrinter) PartitionTable(data []*models.V1PartitionResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Name", "Description", "MinWait", "MaxWait"}
		rows   [][]string
	)

	for _, p := range data {
		rows = append(rows, []string{pointer.SafeDeref(p.ID), p.Name, p.Description, p.Waitingpoolminsize, p.Waitingpoolmaxsize})
	}

	return header, rows, nil
}

func (t *TablePrinter) PartitionCapacityTable(data []*models.V1PartitionCapacity, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Partition", "Size", "Total", "Free", "Allocated", "Other", "Faulty"}
		rows   [][]string

		totalCount     int32
		freeCount      int32
		allocatedCount int32
		faultyCount    int32
		otherCount     int32
	)

	for _, pc := range data {
		pc := pc

		for _, c := range pc.Servers {
			id := pointer.SafeDeref(c.Size)

			allocated := fmt.Sprintf("%d", *c.Allocated)
			total := fmt.Sprintf("%d", *c.Total)
			free := fmt.Sprintf("%d", *c.Free)
			faulty := fmt.Sprintf("%d", *c.Faulty)
			other := fmt.Sprintf("%d", *c.Other)

			if wide {
				if len(c.Faultymachines) > 0 {
					faulty = strings.Join(c.Faultymachines, "\n")
				}
				if len(c.Othermachines) > 0 {
					other = strings.Join(c.Othermachines, "\n")
				}
			}

			totalCount += *c.Total
			freeCount += *c.Free
			allocatedCount += *c.Allocated
			otherCount += *c.Other
			faultyCount += *c.Faulty

			rows = append(rows, []string{*pc.ID, id, total, free, allocated, other, faulty})
		}
	}

	footerRow := ([]string{"Total", "", fmt.Sprintf("%d", totalCount), fmt.Sprintf("%d", freeCount), fmt.Sprintf("%d", allocatedCount), fmt.Sprintf("%d", otherCount), fmt.Sprintf("%d", faultyCount)})
	rows = append(rows, footerRow)

	return header, rows, nil
}
