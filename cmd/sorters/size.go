package sorters

import (
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
			return multisort.Compare(pointer.SafeDeref(a.Partitionid), pointer.SafeDeref(b.Partitionid), descending)
		},
		"size": func(a, b *models.V1SizeReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Sizeid), pointer.SafeDeref(b.Sizeid), descending)
		},
		"tenant": func(a, b *models.V1SizeReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Tenant), pointer.SafeDeref(b.Tenant), descending)
		},
		"project": func(a, b *models.V1SizeReservationResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(pointer.SafeDeref(a.Projectid), pointer.SafeDeref(b.Projectid), descending)
		},
	}, multisort.Keys{{ID: "partition"}, {ID: "size"}, {ID: "tenant"}, {ID: "project"}})
}
