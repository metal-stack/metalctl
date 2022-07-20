package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

var ipSorter = multisort.New(multisort.FieldMap[*models.V1IPResponse]{
	"ipaddress": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.Ipaddress), p.SafeDeref(b.Ipaddress), descending)
	},
	"id": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.Allocationuuid), p.SafeDeref(b.Allocationuuid), descending)
	},
	"name": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Name, b.Name, descending)
	},
	"description": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Description, b.Description, descending)
	},
	"network": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.Networkid), p.SafeDeref(b.Networkid), descending)
	},
	"type": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.SafeDeref(a.Type), p.SafeDeref(b.Type), descending)
	},
})

func IPSortKeys() []string {
	return ipSorter.AvailableKeys()
}

func IPSort(data []*models.V1IPResponse) error {
	return ipSorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "ipaddress"}})...)
}
