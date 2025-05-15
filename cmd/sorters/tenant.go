package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func TenantSorter() *multisort.Sorter[*models.V1TenantResponse] {
	return multisort.New(multisort.FieldMap[*models.V1TenantResponse]{
		"id": func(a, b *models.V1TenantResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Meta).ID, p.SafeDeref(b.Meta).ID, descending)
		},
		"name": func(a, b *models.V1TenantResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Name, b.Name, descending)
		},
		"description": func(a, b *models.V1TenantResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Description, b.Description, descending)
		},
	}, multisort.Keys{{ID: "id"}})
}
