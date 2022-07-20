package sorters

import (
	"sort"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

var partitionSorter = multisort.New(multisort.FieldMap[*models.V1PartitionResponse]{
	"id": func(a, b *models.V1PartitionResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
	},
	"name": func(a, b *models.V1PartitionResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Name, b.Name, descending)
	},
	"description": func(a, b *models.V1PartitionResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Description, b.Description, descending)
	},
})

var partitionCapacitySorter = multisort.New(multisort.FieldMap[*models.V1PartitionCapacity]{
	"id": func(a, b *models.V1PartitionCapacity, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
	},
	"name": func(a, b *models.V1PartitionCapacity, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Name, b.Name, descending)
	},
	"description": func(a, b *models.V1PartitionCapacity, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Description, b.Description, descending)
	},
})

func PartitionSortKeys() []string {
	return partitionSorter.AvailableKeys()
}

func PartitionSort(data []*models.V1PartitionResponse) error {
	return partitionSorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}

func PartitionCapacitySortKeys() []string {
	return partitionCapacitySorter.AvailableKeys()
}

func PartitionCapacitySort(data []*models.V1PartitionCapacity) error {
	for _, pc := range data {
		pc := pc
		sort.SliceStable(pc.Servers, func(i, j int) bool {
			return pointer.SafeDeref(pointer.SafeDeref(pc.Servers[i]).Size) < pointer.SafeDeref(pointer.SafeDeref(pc.Servers[j]).Size)
		})
	}

	return partitionCapacitySorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
