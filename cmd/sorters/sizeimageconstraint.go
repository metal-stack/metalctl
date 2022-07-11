package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func SizeImageConstraintSorter() *multisort.Sorter[*models.V1SizeImageConstraintResponse] {
	return multisort.New(multisort.FieldMap[*models.V1SizeImageConstraintResponse]{
		"id": func(a, b *models.V1SizeImageConstraintResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.Deref(a.ID), p.Deref(b.ID), descending)
		},
		"name": func(a, b *models.V1SizeImageConstraintResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1SizeImageConstraintResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
	})
}

func SizeImageConstraintSort(data []*models.V1SizeImageConstraintResponse) error {
	return SizeImageConstraintSorter().SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
