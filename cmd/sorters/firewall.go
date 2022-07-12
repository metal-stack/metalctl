package sorters

import (
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
	p "github.com/metal-stack/metal-lib/pkg/pointer"
)

var firewallSorter = multisort.New(multisort.FieldMap[*models.V1FirewallResponse]{
	"id": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.Deref(a.ID), p.Deref(b.ID), descending)
	},
	"size": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		aID := p.Deref(p.Deref(a.Size).ID)
		bID := p.Deref(p.Deref(b.Size).ID)
		return multisort.Compare(aID, bID, descending)
	},
	"image": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		aID := p.Deref(p.Deref(p.Deref(a.Allocation).Image).ID)
		bID := p.Deref(p.Deref(p.Deref(b.Allocation).Image).ID)
		return multisort.Compare(aID, bID, descending)
	},
	"partition": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		aID := p.Deref(p.Deref(a.Partition).ID)
		bID := p.Deref(p.Deref(b.Partition).ID)
		return multisort.Compare(aID, bID, descending)
	},
	"project": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		aID := p.Deref(p.Deref(a.Allocation).Project)
		bID := p.Deref(p.Deref(b.Allocation).Project)
		return multisort.Compare(aID, bID, descending)
	},
	"liveliness": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		return multisort.Compare(p.Deref(a.Liveliness), p.Deref(b.Liveliness), descending)
	},
	"when": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		aTime := time.Time(p.Deref(a.Events).LastEventTime)
		bTime := time.Time(p.Deref(b.Events).LastEventTime)
		return multisort.Compare(bTime.Unix(), aTime.Unix(), descending)
	},
	"age": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		aTime := time.Time(p.Deref(p.Deref(a.Allocation).Created))
		bTime := time.Time(p.Deref(p.Deref(b.Allocation).Created))
		return multisort.Compare(bTime.Unix(), aTime.Unix(), descending)
	},
	"event": func(a, b *models.V1FirewallResponse, descending bool) multisort.CompareResult {
		aEvent := p.Deref(p.Deref(p.FirstOrZero(p.Deref(a.Events).Log)).Event)
		bEvent := p.Deref(p.Deref(p.FirstOrZero(p.Deref(b.Events).Log)).Event)
		return multisort.Compare(aEvent, bEvent, descending)
	},
})

func FirewallSortKeys() []string {
	return firewallSorter.AvailableKeys()
}

func FirewallSort(data []*models.V1FirewallResponse) error {
	return firewallSorter.SortBy(data, MustKeysFromCLIOrDefaults(multisort.Keys{{ID: "id"}})...)
}
