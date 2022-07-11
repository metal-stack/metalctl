package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func NetworkSorter() *multisort.Sorter[*models.V1NetworkResponse] {
	return multisort.New(multisort.FieldMap[*models.V1NetworkResponse]{
		"id": func(a, b *models.V1NetworkResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.Deref(a.ID), p.Deref(b.ID), descending)
		},
		"name": func(a, b *models.V1NetworkResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1NetworkResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
		"partition": func(a, b *models.V1NetworkResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Partitionid, b.Partitionid, descending)
		},
		"project": func(a, b *models.V1NetworkResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Projectid, b.Projectid, descending)
		},
	})
}

func NetworkSort(data []*models.V1NetworkResponse) error {
	return NetworkSorter().SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "partition"}, {ID: "id"}})...)
}
