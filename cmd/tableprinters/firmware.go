package tableprinters

import (
	"sort"

	"github.com/metal-stack/metal-go/api/models"
)

func FirmwareTable(data *models.V1FirmwaresResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Description", "Filesystems", "Sizes", "Images"}
		rows   [][]string
	)

	for k, vv := range data.Revisions {
		for v, bb := range vv.VendorRevisions {
			for b, rr := range bb.BoardRevisions {
				sort.Strings(rr)
				for _, rev := range rr {
					rows = append(rows, []string{k, v, b, rev})
				}
			}
		}

	}

	return header, rows, nil
}
