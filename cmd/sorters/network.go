package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

var networkSorter = multisort.New(multisort.FieldMap[*models.V1NetworkResponse]{
	"id": func(a, b *models.V1NetworkResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
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

func NetworkSortKeys() []string {
	return networkSorter.AvailableKeys()
}

func NetworkSort(data []*models.V1NetworkResponse) error {
	return networkSorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "partition"}, {ID: "id"}})...)
}
