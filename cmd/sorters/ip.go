package sorters

import (
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

var ipSorter = multisort.New(multisort.FieldMap[*models.V1IPResponse]{
	"ipaddress": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.Deref(a.Ipaddress), p.Deref(b.Ipaddress), descending)
	},
	"id": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.Deref(a.Allocationuuid), p.Deref(b.Allocationuuid), descending)
	},
	"name": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Name, b.Name, descending)
	},
	"description": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(a.Description, b.Description, descending)
	},
	"network": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.Deref(a.Networkid), p.Deref(b.Networkid), descending)
	},
	"type": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.Deref(a.Type), p.Deref(b.Type), descending)
	},
})

func IPSortKeys() []string {
	return ipSorter.AvailableKeys()
}

func IPSort(data []*models.V1IPResponse) error {
	return ipSorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "ipaddress"}})...)
}
