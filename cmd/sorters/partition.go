package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func PartitionSorter() *multisort.Sorter[*models.V1PartitionResponse] {
	return multisort.New(multisort.FieldMap[*models.V1PartitionResponse]{
		"id": func(a, b *models.V1PartitionResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.Deref(a.ID), p.Deref(b.ID), descending)
		},
		"name": func(a, b *models.V1PartitionResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1PartitionResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
	})
}

func PartitionSort(data []*models.V1PartitionResponse) error {
	return PartitionSorter().SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
