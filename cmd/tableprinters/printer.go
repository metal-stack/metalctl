package tableprinters

import (
	"fmt"
	"time"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/pkg/api"
)

type TablePrinter struct {
	t                       *printers.TablePrinter
	lastEventErrorThreshold time.Duration
	markdown                bool
}

func New() *TablePrinter {
	return &TablePrinter{}
}

func (t *TablePrinter) SetPrinter(printer *printers.TablePrinter) {
	t.t = printer
}

func (t *TablePrinter) SetLastEventErrorThreshold(threshold time.Duration) {
	t.lastEventErrorThreshold = threshold
}

func (t *TablePrinter) SetMarkdown(markdown bool) {
	t.markdown = markdown
}

func (t *TablePrinter) ToHeaderAndRows(data any, wide bool) ([]string, [][]string, error) {
	switch d := data.(type) {
	case []*models.V1AuditResponse:
		return t.AuditTable(d, wide)
	case *models.V1AuditResponse:
		return t.AuditTable(pointer.WrapInSlice(d), wide)
	case []*models.V1MachineResponse:
		return t.MachineTable(d, wide)
	case *models.V1MachineResponse:
		return t.MachineTable(pointer.WrapInSlice(d), wide)
	case *MachinesAndIssues:
		return t.MachineIssuesTable(d, wide)
	case []*models.V1MachineIssue:
		return t.MachineIssuesListTable(d, wide)
	case *models.V1MachineIssue:
		return t.MachineIssuesListTable(pointer.WrapInSlice(d), wide)
	case []*models.V1FirewallResponse:
		return t.FirewallTable(d, wide)
	case *models.V1FirewallResponse:
		return t.FirewallTable(pointer.WrapInSlice(d), wide)
	case []*models.V1ImageResponse:
		return t.ImageTable(d, wide)
	case *models.V1ImageResponse:
		return t.ImageTable(pointer.WrapInSlice(d), wide)
	case []*models.V1PartitionResponse:
		return t.PartitionTable(d, wide)
	case *models.V1PartitionResponse:
		return t.PartitionTable(pointer.WrapInSlice(d), wide)
	case []*models.V1PartitionCapacity:
		return t.PartitionCapacityTable(d, wide)
	case []*models.V1SwitchResponse:
		return t.SwitchTable(d, wide)
	case *models.V1SwitchResponse:
		return t.SwitchTable(pointer.WrapInSlice(d), wide)
	case []*SwitchDetail:
		return t.SwitchDetailTable(d, wide)
	case *SwitchesWithMachines:
		return t.SwitchWithConnectedMachinesTable(d, wide)
	case *models.V1NetworkResponse:
		return t.NetworkTable(pointer.WrapInSlice(d), wide)
	case []*models.V1NetworkResponse:
		return t.NetworkTable(d, wide)
	case *models.V1IPResponse:
		return t.IPTable(pointer.WrapInSlice(d), wide)
	case []*models.V1IPResponse:
		return t.IPTable(d, wide)
	case *models.V1ProjectResponse:
		return t.ProjectTable(pointer.WrapInSlice(d), wide)
	case []*models.V1ProjectResponse:
		return t.ProjectTable(d, wide)
	case *models.V1TenantResponse:
		return t.TenantTable(pointer.WrapInSlice(d), wide)
	case []*models.V1TenantResponse:
		return t.TenantTable(d, wide)
	case []*models.V1MachineIPMIResponse:
		return t.MachineIPMITable(d, wide)
	case *models.V1MachineIPMIResponse:
		return t.MachineIPMITable(pointer.WrapInSlice(d), wide)
	case []*models.V1MachineProvisioningEvent:
		return t.MachineLogsTable(d, wide)
	case *models.V1MachineProvisioningEvent:
		return t.MachineLogsTable(pointer.WrapInSlice(d), wide)
	case *models.V1FirmwaresResponse:
		return t.FirmwareTable(d, wide)
	case *models.V1FilesystemLayoutResponse:
		return t.FSLTable(pointer.WrapInSlice(d), wide)
	case []*models.V1FilesystemLayoutResponse:
		return t.FSLTable(d, wide)
	case *api.Contexts:
		return t.ContextTable(d, wide)

	case *models.V1SizeImageConstraintResponse:
		return t.SizeImageConstraintTable(pointer.WrapInSlice(d), wide)
	case []*models.V1SizeImageConstraintResponse:
		return t.SizeImageConstraintTable(d, wide)
	case *models.V1SizeResponse:
		return t.SizeTable(pointer.WrapInSlice(d), wide)
	case []*models.V1SizeResponse:
		return t.SizeTable(d, wide)
	case *models.V1SizeReservationResponse:
		return t.SizeReservationTable(pointer.WrapInSlice(d), wide)
	case []*models.V1SizeReservationResponse:
		return t.SizeReservationTable(d, wide)
	case *models.V1SizeReservationUsageResponse:
		return t.SizeReservationUsageTable(pointer.WrapInSlice(d), wide)
	case []*models.V1SizeReservationUsageResponse:
		return t.SizeReservationUsageTable(d, wide)

		// V2 Table printers

	case *apiv2.Image:
		return t.V2ImageTable(pointer.WrapInSlice(d), wide)
	case []*apiv2.Image:
		return t.V2ImageTable(d, wide)

	default:
		return nil, nil, fmt.Errorf("unknown table printer for type: %T", d)
	}
}
