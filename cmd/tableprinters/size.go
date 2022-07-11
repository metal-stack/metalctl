package tableprinters

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func SizeTable(data []*models.V1SizeResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Name", "Description", "CPU Range", "Memory Range", "Storage Range"}
		rows   [][]string
	)

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

		rows = append(rows, []string{pointer.Deref(size.ID), size.Name, size.Description, cpu, memory, storage})
	}

	return header, rows, nil
}
