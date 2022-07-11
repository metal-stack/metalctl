package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func ProjectSorter() *multisort.Sorter[*models.V1ProjectResponse] {
	return multisort.New(multisort.FieldMap[*models.V1ProjectResponse]{
		"id": func(a, b *models.V1ProjectResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.Deref(a.Meta).ID, p.Deref(b.Meta).ID, descending)
		},
		"name": func(a, b *models.V1ProjectResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1ProjectResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
		"tenant": func(a, b *models.V1ProjectResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.TenantID, b.TenantID, descending)
		},
	})
}

func ProjectSort(data []*models.V1ProjectResponse) error {
	return ProjectSorter().SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}