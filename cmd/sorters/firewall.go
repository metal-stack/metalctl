package sorters

import (
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

func FirewallSorter() *multisort.Sorter[*models.V1FirewallResponse] {
	return multisort.New(multisort.FieldMap[*models.V1FirewallResponse]{
		"id": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.ID), p.SafeDeref(b.ID), descending)
		},
		"size": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			aID := p.SafeDeref(p.SafeDeref(a.Size).ID)
			bID := p.SafeDeref(p.SafeDeref(b.Size).ID)
			return multisort.Compare(aID, bID, descending)
		},
		"image": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			aID := p.SafeDeref(p.SafeDeref(p.SafeDeref(a.Allocation).Image).ID)
			bID := p.SafeDeref(p.SafeDeref(p.SafeDeref(b.Allocation).Image).ID)
			return multisort.Compare(aID, bID, descending)
		},
		"partition": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			aID := p.SafeDeref(p.SafeDeref(a.Partition).ID)
			bID := p.SafeDeref(p.SafeDeref(b.Partition).ID)
			return multisort.Compare(aID, bID, descending)
		},
		"project": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			aID := p.SafeDeref(p.SafeDeref(a.Allocation).Project)
			bID := p.SafeDeref(p.SafeDeref(b.Allocation).Project)
			return multisort.Compare(aID, bID, descending)
		},
		"liveliness": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(p.SafeDeref(a.Liveliness), p.SafeDeref(b.Liveliness), descending)
		},
		"when": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			aTime := time.Time(p.SafeDeref(a.Events).LastEventTime)
			bTime := time.Time(p.SafeDeref(b.Events).LastEventTime)
			return multisort.Compare(bTime.Unix(), aTime.Unix(), descending)
		},
		"age": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			aTime := time.Time(p.SafeDeref(p.SafeDeref(a.Allocation).Created))
			bTime := time.Time(p.SafeDeref(p.SafeDeref(b.Allocation).Created))
			return multisort.Compare(bTime.Unix(), aTime.Unix(), descending)
		},
		"event": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
			aEvent := p.SafeDeref(p.SafeDeref(p.FirstOrZero(p.SafeDeref(a.Events).Log)).Event)
			bEvent := p.SafeDeref(p.SafeDeref(p.FirstOrZero(p.SafeDeref(b.Events).Log)).Event)
			return multisort.Compare(aEvent, bEvent, descending)
		},
	}, multisort.Keys{{ID: "id"}})
}
