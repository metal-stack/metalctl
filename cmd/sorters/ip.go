package sorters

import (
	"net/netip"
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func IPSorter() *multisort.Sorter[*models.V1IPResponse] {
	return multisort.New(multisort.FieldMap[*models.V1IPResponse]{
		"ipaddress": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
			aIP, _ := netip.ParseAddr(p.SafeDeref(a.Ipaddress))
			bIP, _ := netip.ParseAddr(p.SafeDeref(b.Ipaddress))
			return multisort.WithCompareFunc(func() int {
				return aIP.Compare(bIP)
			}, descending)
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
		"age": func(a, b *models.V1IPResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(time.Time(a.Created).Unix(), time.Time(b.Created).Unix(), descending)
		},
	}, multisort.Keys{{ID: "ipaddress"}})
}
