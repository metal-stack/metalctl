package sorters

import (
	"slices"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func SizeSorter() *multisort.Sorter[*models.V1SizeResponse] {
	return multisort.New(multisort.FieldMap[*models.V1SizeResponse]{
		"id": func(a, b *models.V1SizeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
		},
		"name": func(a, b *models.V1SizeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1SizeResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
	}, multisort.Keys{{ID: "id"}})
}

func SizeReservationsSorter() *multisort.Sorter[*models.V1SizeReservationResponse] {
	return multisort.New(multisort.FieldMap[*models.V1SizeReservationResponse]{
		"partition": func(a, b *models.V1SizeReservationResponse, descending bool) multisort.CompareResult {
			slices.Sort(a.Partitionid)
			slices.Sort(b.Partitionid)
			return multisort.Compare(strings.Join(a.Partitionid, " "), strings.Join(b.Partitionid, " "), descending)
		},
		"size": func(a, b *models.V1SizeReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Sizeid), pointer.SafeDeref(b.Sizeid), descending)
		},
		"project": func(a, b *models.V1SizeReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Projectid), pointer.SafeDeref(b.Projectid), descending)
		},
		"amount": func(a, b *models.V1SizeReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Amount), pointer.SafeDeref(b.Amount), descending)
		},
		"id": func(a, b *models.V1SizeReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.ID), pointer.SafeDeref(b.ID), descending)
		},
	}, multisort.Keys{{ID: "partition"}, {ID: "size"}, {ID: "project"}, {ID: "id"}})
}

func SizeReservationsUsageSorter() *multisort.Sorter[*models.V1SizeReservationUsageResponse] {
	return multisort.New(multisort.FieldMap[*models.V1SizeReservationUsageResponse]{
		"partition": func(a, b *models.V1SizeReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Partitionid), pointer.SafeDeref(b.Partitionid), descending)
		},
		"size": func(a, b *models.V1SizeReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Sizeid), pointer.SafeDeref(b.Sizeid), descending)
		},
		"project": func(a, b *models.V1SizeReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Projectid), pointer.SafeDeref(b.Projectid), descending)
		},
		"id": func(a, b *models.V1SizeReservationUsageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.ID), pointer.SafeDeref(b.ID), descending)
		},
	}, multisort.Keys{{ID: "partition"}, {ID: "size"}, {ID: "project"}, {ID: "id"}})
}
