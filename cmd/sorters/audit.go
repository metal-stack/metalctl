package sorters

import (
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/multisort"
)

func AuditSorter() *multisort.Sorter[*models.V1AuditResponse] {
	return multisort.New(multisort.FieldMap[*models.V1AuditResponse]{
		"timestamp": func(a, b *models.V1AuditResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(time.Time(a.Timestamp).Unix(), time.Time(b.Timestamp).Unix(), descending)
		},
		"user": func(a, b *models.V1AuditResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.User, b.User, descending)
		},
		"tenant": func(a, b *models.V1AuditResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Tenant, b.Tenant, descending)
		},
		"path": func(a, b *models.V1AuditResponse, descending bool) multisort.CompareResult {
			return multisort.Compare(a.Path, b.Path, descending)
		},
	}, multisort.Keys{{ID: "timestamp", Descending: true}})
}
