package tableprinters

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func (t *TablePrinter) SwitchTable(data []*models.V1SwitchResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Partition", "Rack", "Status"}
	if wide {
		header = []string{"ID", "Partition", "Rack", "Mode", "Last Sync", "Sync Duration", "Last Sync Error"}
	}

	for _, s := range data {
		id := pointer.Deref(s.ID)
		partition := pointer.Deref(pointer.Deref(s.Partition).ID)
		rack := pointer.Deref(s.RackID)

		syncAgeStr := ""
		syncDurStr := ""
		syncError := ""
		shortStatus := nbr
		var syncTime time.Time
		if s.LastSync != nil {
			syncTime = time.Time(*s.LastSync.Time)
			syncAge := time.Since(syncTime)
			syncDur := time.Duration(*s.LastSync.Duration).Round(time.Millisecond)
			if syncAge >= time.Minute*10 || syncDur >= 30*time.Second {
				shortStatus += color.RedString(dot)
			} else if syncAge >= time.Minute*1 || syncDur >= 20*time.Second {
				shortStatus += color.YellowString(dot)
			} else {
				shortStatus += color.GreenString(dot)
			}

			syncAgeStr = humanizeDuration(syncAge)
			syncDurStr = fmt.Sprintf("%v", syncDur)
		}

		if s.LastSyncError != nil {
			errorTime := time.Time(*s.LastSyncError.Time)
			syncError = fmt.Sprintf("%s ago: %s", humanizeDuration(time.Since(errorTime)), s.LastSyncError.Error)
		}

		var mode string
		switch s.Mode {
		case "replace":
			shortStatus = nbr + color.RedString(dot)
			mode = "replace"
		default:
			mode = "operational"
		}

		if wide {
			rows = append(rows, []string{id, partition, rack, mode, syncAgeStr, syncDurStr, syncError})
		} else {
			rows = append(rows, []string{id, partition, rack, shortStatus})
		}
	}

	return header, rows, nil
}

type SwitchDetail struct {
	*models.V1SwitchResponse
}

func (t *TablePrinter) SwitchDetailTable(data []*SwitchDetail, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Partition", "Rack", "Switch", "Port", "Machine", "VNI-Filter", "CIDR-Filter"}
		rows   [][]string
	)

	for _, sw := range data {
		sw := sw
		filterBySwp := map[string]models.V1BGPFilter{}
		for _, n := range sw.Nics {
			swp := *(n.Name)
			if n.Filter != nil {
				filterBySwp[swp] = *(n.Filter)
			}
		}

		for _, conn := range sw.Connections {
			swp := *conn.Nic.Name
			partitionID := ""
			if sw.Partition != nil {
				partitionID = *sw.Partition.ID
			}

			f := filterBySwp[swp]
			row := []string{partitionID, *sw.RackID, *sw.ID, swp, conn.MachineID}
			row = append(row, filterColumns(f, 0)...)
			max := len(f.Vnis)
			if len(f.Cidrs) > max {
				max = len(f.Cidrs)
			}
			rows = append(rows, row)
			for i := 1; i < max; i++ {
				row = append([]string{"", "", "", "", ""}, filterColumns(f, i)...)
				rows = append(rows, row)
			}
		}
	}

	return header, rows, nil
}

func filterColumns(f models.V1BGPFilter, i int) []string {
	v := ""
	if len(f.Vnis) > i {
		v = f.Vnis[i]
	}
	c := ""
	if len(f.Cidrs) > i {
		c = f.Cidrs[i]
	}
	return []string{v, c}
}
