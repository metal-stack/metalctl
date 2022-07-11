package tableprinters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func SizeImageConstraintTable(data []*models.V1SizeImageConstraintResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Name", "Description", "Image", "Constraint"}
		rows   [][]string
	)

	for _, size := range data {
		for i, c := range size.Constraints.Images {
			rows = append(rows, []string{pointer.Deref(size.ID), size.Name, size.Description, i, c})
		}
	}

	return header, rows, nil
}