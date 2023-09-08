package tableprinters

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
)

func (t *TablePrinter) SwitchTable(data []*models.V1SwitchResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Partition", "Rack", "OS", "Status"}
	if wide {
		header = []string{"ID", "Partition", "Rack", "OS", "MetalCore", "IP", "Mode", "Last Sync", "Sync Duration", "Last Sync Error"}

		t.t.MutateTable(func(table *tablewriter.Table) {
			table.SetAutoWrapText(false)
		})
	}

	for _, s := range data {
		id := pointer.SafeDeref(s.ID)
		partition := pointer.SafeDeref(pointer.SafeDeref(s.Partition).ID)
		rack := pointer.SafeDeref(s.RackID)

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
				shortStatus = color.RedString(dot)
			} else if syncAge >= time.Minute*1 || syncDur >= 20*time.Second {
				shortStatus = color.YellowString(dot)
			} else {
				shortStatus = color.GreenString(dot)
			}

			syncAgeStr = humanizeDuration(syncAge)
			syncDurStr = fmt.Sprintf("%v", syncDur)
		}

		if s.LastSyncError != nil {
			errorTime := time.Time(*s.LastSyncError.Time)
			// after 7 days we do not show sync errors anymore
			if !errorTime.IsZero() && time.Since(errorTime) < 7*24*time.Hour {
				syncError = fmt.Sprintf("%s ago: %s", humanizeDuration(time.Since(errorTime)), s.LastSyncError.Error)

				if errorTime.After(time.Time(pointer.SafeDeref(s.LastSync.Time))) {
					shortStatus = color.RedString(dot)
				}
			}
		}

		var mode string
		switch s.Mode {
		case "replace":
			shortStatus = nbr + color.RedString(dot)
			mode = "replace"
		default:
			mode = "operational"
		}

		os := ""
		osIcon := ""
		metalCore := ""
		if s.Os != nil {
			switch strings.ToLower(s.Os.Vendor) {
			case "cumulus":
				osIcon = "🐢"
			case "sonic":
				osIcon = "🦔"
			default:
				osIcon = s.Os.Vendor
			}

			os = s.Os.Vendor
			if s.Os.Version != "" {
				os = fmt.Sprintf("%s (%s)", os, s.Os.Version)
			}
			// metal core version is very long: v0.9.1 (1d5e42ea), tags/v0.9.1-0-g1d5e42e, go1.20.5
			metalCore = strings.Split(s.Os.MetalCoreVersion, ",")[0]
		}

		if wide {
			rows = append(rows, []string{id, partition, rack, os, metalCore, s.ManagementIP, mode, syncAgeStr, syncDurStr, syncError})
		} else {
			rows = append(rows, []string{id, partition, rack, osIcon, shortStatus})
		}
	}

	return header, rows, nil
}

type SwitchesWithMachines struct {
	SS []*models.V1SwitchResponse               `json:"switches" yaml:"switches"`
	MS map[string]*models.V1MachineIPMIResponse `json:"machines" yaml:"machines"`
}

func (t *TablePrinter) SwitchWithConnectedMachinesTable(data *SwitchesWithMachines, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "NIC Name", "Identifier", "Partition", "Rack", "Size", "Product Serial"}

	for _, s := range data.SS {
		id := pointer.SafeDeref(s.ID)
		partition := pointer.SafeDeref(pointer.SafeDeref(s.Partition).ID)
		rack := pointer.SafeDeref(s.RackID)

		rows = append(rows, []string{id, "", "", partition, rack})

		conns := s.Connections
		if viper.IsSet("size") || viper.IsSet("machine-id") {
			filteredConns := []*models.V1SwitchConnection{}

			for _, conn := range s.Connections {
				conn := conn

				m, ok := data.MS[conn.MachineID]
				if !ok {
					continue
				}

				if viper.IsSet("machine-id") && pointer.SafeDeref(m.ID) == viper.GetString("machine-id") {
					filteredConns = append(filteredConns, conn)
				}

				if viper.IsSet("size") && pointer.SafeDeref(m.Size.ID) == viper.GetString("size") {
					filteredConns = append(filteredConns, conn)
				}
			}

			conns = filteredConns
		}

		sort.Slice(conns, switchInterfaceNameLessFunc(conns))

		for i, conn := range conns {
			prefix := "├"
			if i == len(conns)-1 {
				prefix = "└"
			}
			prefix += "─╴"

			m, ok := data.MS[conn.MachineID]
			if !ok {
				return nil, nil, fmt.Errorf("switch port %s is connected to a machine which does not exist: %q", pointer.SafeDeref(pointer.SafeDeref(conn.Nic).Name), conn.MachineID)
			}

			identifier := pointer.SafeDeref(conn.Nic.Identifier)
			if identifier == "" {
				identifier = pointer.SafeDeref(conn.Nic.Mac)
			}

			rows = append(rows, []string{
				fmt.Sprintf("%s%s", prefix, pointer.SafeDeref(m.ID)),
				pointer.SafeDeref(pointer.SafeDeref(conn.Nic).Name),
				identifier,
				pointer.SafeDeref(pointer.SafeDeref(m.Partition).ID),
				m.Rackid,
				pointer.SafeDeref(pointer.SafeDeref(m.Size).ID),
				pointer.SafeDeref(pointer.SafeDeref(m.Ipmi).Fru).ProductSerial,
			})
		}
	}

	return header, rows, nil
}

var numberRegex = regexp.MustCompile("([0-9]+)")

func switchInterfaceNameLessFunc(conns []*models.V1SwitchConnection) func(i, j int) bool {
	return func(i, j int) bool {
		var (
			a = pointer.SafeDeref(pointer.SafeDeref(conns[i]).Nic.Name)
			b = pointer.SafeDeref(pointer.SafeDeref(conns[j]).Nic.Name)

			aMatch = numberRegex.FindAllStringSubmatch(a, -1)
			bMatch = numberRegex.FindAllStringSubmatch(b, -1)
		)

		for i := range aMatch {
			if i >= len(bMatch) {
				return true
			}

			interfaceNumberA, aErr := strconv.Atoi(aMatch[i][0])
			interfaceNumberB, bErr := strconv.Atoi(bMatch[i][0])

			if aErr == nil && bErr == nil {
				if interfaceNumberA < interfaceNumberB {
					return true
				}
				if interfaceNumberA != interfaceNumberB {
					return false
				}
			}
		}

		return a < b
	}
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
