package sorters

import (
	"sort"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	"github.com/metal-stack/metal-lib/pkg/pointer"
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
	})
}

func SwitchSort(data []*models.V1SwitchResponse) error {
	for _, s := range data {
		s := s
		sort.SliceStable(s.Connections, func(i, j int) bool {
			return pointer.Deref(pointer.Deref((pointer.Deref(s.Connections[i])).Nic).Name) < pointer.Deref(pointer.Deref((pointer.Deref(s.Connections[j])).Nic).Name)
		})
	}
	return SwitchSorter().SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
