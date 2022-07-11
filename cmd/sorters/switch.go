package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func SwitchSorter() *multisort.Sorter[*models.V1SwitchResponse] {
	return multisort.New(multisort.FieldMap[*models.V1SwitchResponse]{
		"id": func(a, b *models.V1SwitchResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.Deref(a.ID), p.Deref(b.ID), descending)
		},
		"name": func(a, b *models.V1SwitchResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1SwitchResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
		// TODO: make this work
		// sort.SliceStable(sw.Connections, func(i, j int) bool { return *((*sw.Connections[i]).Nic.Name) < *((*sw.Connections[j]).Nic.Name) })
	})
}

func SwitchSort(data []*models.V1SwitchResponse) error {
	return SwitchSorter().SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
