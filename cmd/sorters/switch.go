package sorters

import (
	"sort"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

var switchSorter = multisort.New(multisort.FieldMap[*models.V1SwitchResponse]{
	"id": func(a, b *models.V1SwitchResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
	},
	"name": func(a, b *models.V1SwitchResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Name, b.Name, descending)
	},
	"description": func(a, b *models.V1SwitchResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Description, b.Description, descending)
	},
})

func SwitchSortKeys() []string {
	return switchSorter.AvailableKeys()
}

func SwitchSort(data []*models.V1SwitchResponse) error {
	for _, s := range data {
		s := s
		sort.SliceStable(s.Connections, func(i, j int) bool {
			return pointer.SafeDeref(pointer.SafeDeref((pointer.SafeDeref(s.Connections[i])).Nic).Name) < pointer.SafeDeref(pointer.SafeDeref((pointer.SafeDeref(s.Connections[j])).Nic).Name)
		})
	}
	return switchSorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
