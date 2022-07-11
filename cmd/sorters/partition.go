package sorters

import (
	"sort"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	"github.com/metal-stack/metal-lib/pkg/pointer"
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

func PartitionCapacitySorter() *multisort.Sorter[*models.V1PartitionCapacity] {
	return multisort.New(multisort.FieldMap[*models.V1PartitionCapacity]{
		"id": func(a, b *models.V1PartitionCapacity, descending bool) multisort.CompareResult {
			return multisort.Compare(p.Deref(a.ID), p.Deref(b.ID), descending)
		},
		"name": func(a, b *models.V1PartitionCapacity, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1PartitionCapacity, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
	})
}

func PartitionCapacitySort(data []*models.V1PartitionCapacity) error {
	for _, pc := range data {
		pc := pc
		sort.SliceStable(pc.Servers, func(i, j int) bool {
			return pointer.Deref(pointer.Deref(pc.Servers[i]).Size) < pointer.Deref(pointer.Deref(pc.Servers[j]).Size)
		})
	}

	return PartitionCapacitySorter().SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
