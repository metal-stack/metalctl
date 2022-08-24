package sorters

import (
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func ImageSorter() *multisort.Sorter[*models.V1ImageResponse] {
	return multisort.New(multisort.FieldMap[*models.V1ImageResponse]{
		"id": func(a, b *models.V1ImageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
		},
		"name": func(a, b *models.V1ImageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1ImageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
		"classification": func(a, b *models.V1ImageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Classification, b.Classification, descending)
		},
		"expiration": func(a, b *models.V1ImageResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(time.Time(p.SafeDeref(a.ExpirationDate)).Unix(), time.Time(p.SafeDeref(b.ExpirationDate)).Unix(), descending)
		},
	}, multisort.Keys{{ID: "id"}})
}
