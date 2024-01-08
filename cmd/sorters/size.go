package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
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

func SizeReservationsSorter() *multisort.Sorter[*tableprinters.SizeReservation] {
	return multisort.New(multisort.FieldMap[*tableprinters.SizeReservation]{
		"partition": func(a, b *tableprinters.SizeReservation, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Partition, b.Partition, descending)
		},
		"size": func(a, b *tableprinters.SizeReservation, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Size, b.Size, descending)
		},
		"tenant": func(a, b *tableprinters.SizeReservation, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Tenant, b.Tenant, descending)
		},
		"project": func(a, b *tableprinters.SizeReservation, descending bool) multisort.CompareResult {
			return multisort.Compare(a.ProjectID, b.ProjectID, descending)
		},
	}, multisort.Keys{{ID: "partition"}, {ID: "size"}, {ID: "tenant"}, {ID: "project"}})
}
