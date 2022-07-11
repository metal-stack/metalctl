package tableprinters

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metalctl/pkg/api"
)

func ToHeaderAndRows(data interface{}, wide bool) ([]string, [][]string, error) {
	switch d := data.(type) {
	case []*models.V1MachineResponse:
		return MachineTable(d, wide)
	case *models.V1MachineResponse:
		return MachineTable(toArray(d), wide)
	case api.MachineIssues:
		return MachineIssuesTable(d, wide)
	case []*models.V1FirewallResponse:
		return FirewallTable(d, wide)
	case *models.V1FirewallResponse:
		return FirewallTable(toArray(d), wide)
	case []*models.V1ImageResponse:
		return ImageTable(d, wide)
	case []*models.V1PartitionResponse:
		return PartitionTable(d, wide)
	case []*models.V1PartitionCapacity:
		return PartitionCapacityTable(d, wide)
	case []*models.V1SwitchResponse:
		return SwitchTable(d, wide)
	case []*SwitchDetail:
		return SwitchDetailTable(d, wide)
	case *models.V1NetworkResponse:
		return NetworkTable(toArray(d), wide)
	case []*models.V1NetworkResponse:
		return NetworkTable(d, wide)
	case *models.V1IPResponse:
		return IPTable(toArray(d), wide)
	case []*models.V1IPResponse:
		return IPTable(d, wide)
	case *models.V1ProjectResponse:
		return ProjectTable(toArray(d), wide)
	case []*models.V1ProjectResponse:
		return ProjectTable(d, wide)
	case []*models.V1MachineIPMIResponse:
		return MachineIPMITable(d, wide)
	case *models.V1MachineIPMIResponse:
		return MachineIPMITable(toArray(d), wide)
	case []*models.V1MachineProvisioningEvent:
		return MachineLogsTable(d, wide)
	case *models.V1FirmwaresResponse:
		return FirmwareTable(d, wide)
	case *models.V1FilesystemLayoutResponse:
		return FSLTable(toArray(d), wide)
	case []*models.V1FilesystemLayoutResponse:
		return FSLTable(d, wide)
	case *api.Contexts:
		return ContextTable(d, wide)
	case *models.V1SizeImageConstraintResponse:
		return SizeImageConstraintTable(toArray(d), wide)
	case []*models.V1SizeImageConstraintResponse:
		return SizeImageConstraintTable(d, wide)
	case *models.V1SizeResponse:
		return SizeTable(toArray(d), wide)
	case []*models.V1SizeResponse:
		return SizeTable(d, wide)
	case []*models.V1SizeMatchingLog:
		return SizeMatchingLogTable(d, wide)
	default:
		return nil, nil, fmt.Errorf("unknown table printer for type: %T", d)
	}
}

func toArray[E any](e E) []E {
	return []E{e}
}
