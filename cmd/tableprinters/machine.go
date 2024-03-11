package tableprinters

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/olekukonko/tablewriter"
)

// MachinesAndIssues is used for combining issues with more data on machines.
type MachinesAndIssues struct {
	EvaluationResult []*models.V1MachineIssueResponse `json:"evaluation_result" yaml:"evaluation_result"`
	Machines         []*models.V1MachineIPMIResponse  `json:"machines" yaml:"machines"`
	Issues           []*models.V1MachineIssue         `json:"issues" yaml:"issues"`
}

func (t *TablePrinter) MachineTable(data []*models.V1MachineResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "", "Last Event", "When", "Age", "Hostname", "Project", "Size", "Image", "Partition", "Rack"}
	if wide {
		header = []string{"ID", "Last Event", "When", "Age", "Description", "Name", "Hostname", "Project", "IPs", "Size", "Image", "Partition", "Rack", "Started", "Tags", "Lock/Reserve"}
	}

	for _, machine := range data {
		machineID := *machine.ID
		if machine.Ledstate != nil && *machine.Ledstate.Value == "LED-ON" {
			blue := color.New(color.FgBlue).SprintFunc()
			machineID = blue(machineID)
		}

		alloc := pointer.SafeDeref(machine.Allocation)
		sizeID := pointer.SafeDeref(pointer.SafeDeref(machine.Size).ID)
		partitionID := pointer.SafeDeref(pointer.SafeDeref(machine.Partition).ID)
		project := pointer.SafeDeref(alloc.Project)
		name := pointer.SafeDeref(alloc.Name)
		desc := alloc.Description
		hostname := pointer.SafeDeref(alloc.Hostname)
		image := pointer.SafeDeref(alloc.Image).Name

		rack := machine.Rackid

		truncatedHostname := genericcli.TruncateEnd(hostname, 30)

		var nwIPs []string
		for _, nw := range alloc.Networks {
			nwIPs = append(nwIPs, nw.Ips...)
		}
		ips := strings.Join(nwIPs, "\n")

		started := ""
		age := ""
		if alloc.Created != nil && !time.Time(*alloc.Created).IsZero() {
			started = time.Time(*alloc.Created).Format(time.RFC3339)
			age = humanizeDuration(time.Since(time.Time(*alloc.Created)))
		}
		tags := ""
		if len(machine.Tags) > 0 {
			tags = strings.Join(machine.Tags, ",")
		}

		reserved := ""
		if *machine.State.Value != "" {
			reserved = fmt.Sprintf("%s:%s", *machine.State.Value, *machine.State.Description)
		}

		lastEvent := ""
		when := ""
		if len(machine.Events.Log) > 0 {
			since := time.Since(time.Time(machine.Events.LastEventTime))
			when = humanizeDuration(since)
			lastEvent = *machine.Events.Log[0].Event
		}

		emojis, _ := t.getMachineStatusEmojis(machine.Liveliness, machine.Events, machine.State, alloc.Vpn)

		if wide {
			rows = append(rows, []string{machineID, lastEvent, when, age, desc, name, hostname, project, ips, sizeID, image, partitionID, rack, started, tags, reserved})
		} else {
			rows = append(rows, []string{machineID, emojis, lastEvent, when, age, truncatedHostname, project, sizeID, image, partitionID, rack})
		}
	}

	return header, rows, nil
}

func (t *TablePrinter) getMachineStatusEmojis(liveliness *string, events *models.V1MachineRecentProvisioningEvents, state *models.V1MachineState, vpn *models.V1MachineVPN) (string, string) {
	var (
		emojis []string
		wide   []string
	)

	switch l := pointer.SafeDeref(liveliness); l {
	case "Alive":
		// noop
	case "Dead":
		emojis = append(emojis, api.Skull)
		wide = append(wide, l)
	case "Unknown":
		emojis = append(emojis, api.Question)
		wide = append(wide, l)
	default:
		emojis = append(emojis, api.Question)
		wide = append(wide, l)
	}

	if state != nil {
		switch pointer.SafeDeref(state.Value) {
		case "":
			// noop
		case "LOCKED":
			emojis = append(emojis, api.Lock)
			wide = append(wide, "Locked")
		case "RESERVED":
			emojis = append(emojis, api.Bark)
			wide = append(wide, "Reserved")
		}
	}

	if events != nil {
		if pointer.SafeDeref(events.FailedMachineReclaim) {
			emojis = append(emojis, api.Ambulance)
			wide = append(wide, "FailedReclaim")
		}

		if events.LastErrorEvent != nil && time.Since(time.Time(events.LastErrorEvent.Time)) < t.lastEventErrorThreshold {
			emojis = append(emojis, api.Exclamation)
			wide = append(wide, "LastEventErrors")
		}

		if pointer.SafeDeref(events.CrashLoop) {
			emojis = append(emojis, api.Loop)
			wide = append(wide, "CrashLoop")
		}
	}

	if vpn != nil && *vpn.Connected {
		emojis = append(emojis, api.VPN)
		wide = append(wide, "VPN")
	}

	return strings.Join(emojis, nbr), strings.Join(wide, ", ")
}

func (t *TablePrinter) MachineIPMITable(data []*models.V1MachineIPMIResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "", "Power", "IP", "Mac", "Board Part Number", "Bios", "BMC", "Size", "Partition", "Rack", "Updated"}
	if wide {
		header = []string{"ID", "Status", "Power", "IP", "Mac", "Board Part Number", "Chassis Serial", "Product Serial", "Bios Version", "BMC Version", "Size", "Partition", "Rack", "Updated"}
	}

	for _, machine := range data {
		id := pointer.SafeDeref(machine.ID)
		partition := pointer.SafeDeref(pointer.SafeDeref(machine.Partition).ID)
		size := pointer.SafeDeref(pointer.SafeDeref(machine.Size).ID)

		ipAddress := ""
		mac := ""
		bpn := ""
		cs := ""
		ps := ""
		bmcVersion := ""
		power := color.WhiteString(dot)
		powerText := ""
		ipmi := machine.Ipmi
		rack := machine.Rackid
		lastUpdated := "never"

		if ipmi != nil {
			ipAddress = pointer.SafeDeref(ipmi.Address)
			mac = pointer.SafeDeref(ipmi.Mac)
			bmcVersion = pointer.SafeDeref(ipmi.Bmcversion)
			fru := ipmi.Fru

			if fru != nil {
				bpn = fru.BoardPartNumber
				cs = fru.ChassisPartSerial
				ps = fru.ProductSerial
			}

			power, powerText = extractPowerState(ipmi)

			if ipmi.LastUpdated != nil && !ipmi.LastUpdated.IsZero() {
				lastUpdated = fmt.Sprintf("%s ago", humanizeDuration(time.Since(time.Time(*ipmi.LastUpdated))))
			}
		}

		biosVersion := ""
		bios := machine.Bios
		if bios != nil {
			biosVersion = pointer.SafeDeref(bios.Version)
		}

		emojis, wideEmojis := t.getMachineStatusEmojis(machine.Liveliness, machine.Events, machine.State, nil)

		if wide {
			rows = append(rows, []string{id, wideEmojis, powerText, ipAddress, mac, bpn, cs, ps, biosVersion, bmcVersion, size, partition, rack, lastUpdated})
		} else {
			rows = append(rows, []string{id, emojis, power, ipAddress, mac, bpn, biosVersion, bmcVersion, size, partition, rack, lastUpdated})
		}
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
	})

	return header, rows, nil
}

func extractPowerState(ipmi *models.V1MachineIPMI) (short, wide string) {
	if ipmi == nil || ipmi.Powerstate == nil {
		return color.WhiteString(dot), wide
	}

	state := *ipmi.Powerstate
	switch state {
	case "ON":
		short = color.GreenString(dot)
	case "OFF":
		short = color.RedString(dot)
	default:
		short = color.WhiteString(dot)
	}

	wide = state

	if ipmi.Powermetric != nil {
		short = fmt.Sprintf("%s"+nbr+nbr+"(%.1fW)", short, pointer.SafeDeref(ipmi.Powermetric.Averageconsumedwatts))
		wide = fmt.Sprintf("%s %.2fW", wide, pointer.SafeDeref(ipmi.Powermetric.Averageconsumedwatts))
	}

	return short, wide
}

func (t *TablePrinter) MachineLogsTable(data []*models.V1MachineProvisioningEvent, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Time", "Event", "Message"}
		rows   [][]string
	)

	for _, i := range data {
		msg := i.Message
		if !wide {
			split := strings.Split(msg, "\n")
			if len(split) > 1 {
				msg = split[0] + " " + genericcli.TruncateElipsis
			}
			msg = genericcli.TruncateEnd(msg, 120)
		}
		rows = append(rows, []string{time.Time(i.Time).Format(time.RFC1123), pointer.SafeDeref(i.Event), msg})
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
	})

	return header, rows, nil
}

func (t *TablePrinter) MachineIssuesListTable(data []*models.V1MachineIssue, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Severity", "Description", "Reference URL"}
		rows   [][]string
	)

	for _, issue := range data {
		rows = append(rows, []string{
			pointer.SafeDeref(issue.ID),
			pointer.SafeDeref(issue.Severity),
			pointer.SafeDeref(issue.Description),
			pointer.SafeDeref(issue.RefURL),
		})
	}

	return header, rows, nil
}

func (t *TablePrinter) MachineIssuesTable(data *MachinesAndIssues, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Power", "Allocated", "", "Lock Reason", "Last Event", "When", "Issues"}
	if wide {
		header = []string{"ID", "Name", "Partition", "Project", "Power", "State", "Lock Reason", "Last Event", "When", "Issues", "Ref URL", "Details"}
	}

	machinesByID := map[string]*models.V1MachineIPMIResponse{}
	for _, m := range data.Machines {
		m := m

		if m.ID == nil {
			continue
		}

		machinesByID[*m.ID] = m
	}

	issuesByID := map[string]*models.V1MachineIssue{}
	for _, issue := range data.Issues {
		issue := issue

		if issue.ID == nil {
			continue
		}

		issuesByID[*issue.ID] = issue
	}

	for _, issue := range data.EvaluationResult {
		if issue.Machineid == nil {
			continue
		}

		machine := machinesByID[*issue.Machineid]
		if machine == nil {
			continue
		}

		widename := ""
		if machine.Allocation != nil && machine.Allocation.Name != nil {
			widename = *machine.Allocation.Name
		}
		partition := ""
		if machine.Partition != nil && machine.Partition.ID != nil {
			partition = *machine.Partition.ID
		}
		project := ""
		if machine.Allocation != nil && machine.Allocation.Project != nil {
			project = *machine.Allocation.Project
		}

		allocated := "no"
		if machine.Allocation != nil {
			allocated = "yes"
		}

		lockText := ""
		lockDesc := ""
		lockDescWide := ""
		if machine.State != nil && machine.State.Value != nil && *machine.State.Value != "" {
			lockText = *machine.State.Value
		}
		if machine.State != nil && machine.State.Value != nil && *machine.State.Description != "" {
			lockDescWide = *machine.State.Description
			lockDesc = genericcli.TruncateEnd(lockDescWide, 30)
		}

		power, powerText := extractPowerState(machine.Ipmi)

		when := ""
		lastEvent := ""
		if len(machine.Events.Log) > 0 {
			since := time.Since(time.Time(machine.Events.LastEventTime))
			when = humanizeDuration(since)
			lastEvent = *machine.Events.Log[0].Event
		}

		emojis, _ := t.getMachineStatusEmojis(machine.Liveliness, machine.Events, machine.State, nil)

		for i, id := range issue.Issues {
			iss, ok := issuesByID[id]
			if !ok {
				continue
			}

			text := fmt.Sprintf("%s (%s)", pointer.SafeDeref(iss.Description), pointer.SafeDeref(iss.ID))
			ref := pointer.SafeDeref(iss.RefURL)
			details := pointer.SafeDeref(iss.Details)

			if i != 0 {
				if wide {
					rows = append(rows, []string{"", "", "", "", "", "", "", "", "", text, ref, details})
				} else {
					rows = append(rows, []string{"", "", "", "", "", "", "", text})
				}
				continue
			}

			if wide {
				rows = append(rows, []string{pointer.SafeDeref(machine.ID), widename, partition, project, powerText, lockText, lockDescWide, lastEvent, when, text, ref, details})
			} else {
				rows = append(rows, []string{pointer.SafeDeref(machine.ID), power, allocated, emojis, lockDesc, lastEvent, when, text})
			}
		}
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
	})

	return header, rows, nil
}
