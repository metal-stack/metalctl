package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

var sizeSorter = multisort.New(multisort.FieldMap[*models.V1SizeResponse]{
	"id": func(a, b *models.V1SizeResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
	},
	"name": func(a, b *models.V1SizeResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Name, b.Name, descending)
	},
	"description": func(a, b *models.V1SizeResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Description, b.Description, descending)
	},
})

func SizeSortKeys() []string {
	return sizeSorter.AvailableKeys()
}

func SizeSort(data []*models.V1SizeResponse) error {
	return sizeSorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
