package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

var filesystemLayoutSorter = multisort.New(multisort.FieldMap[*models.V1FilesystemLayoutResponse]{
	"id": func(a, b *models.V1FilesystemLayoutResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.Deref(a.ID), p.Deref(b.ID), descending)
	},
	"name": func(a, b *models.V1FilesystemLayoutResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Name, b.Name, descending)
	},
	"description": func(a, b *models.V1FilesystemLayoutResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Description, b.Description, descending)
	},
})

func FilesystemLayoutSortKeys() []string {
	return filesystemLayoutSorter.AvailableKeys()
}

func FilesystemLayoutSort(data []*models.V1FilesystemLayoutResponse) error {
	return filesystemLayoutSorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
