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
)

func (t *TablePrinter) MachineTable(data []*models.V1MachineResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "", "Last Event", "When", "Age", "Hostname", "Project", "Size", "Image", "Partition"}
	if wide {
		header = []string{"ID", "Last Event", "When", "Age", "Description", "Name", "Hostname", "Project", "IPs", "Size", "Image", "Partition", "Started", "Tags", "Lock/Reserve"}
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

		emojis, _ := getMachineStatusEmojis(machine.Liveliness, machine.Events, machine.State)

		if wide {
			rows = append(rows, []string{machineID, lastEvent, when, age, desc, name, hostname, project, ips, sizeID, image, partitionID, started, tags, reserved})
		} else {
			rows = append(rows, []string{machineID, emojis, lastEvent, when, age, truncatedHostname, project, sizeID, image, partitionID})
		}
	}

	return header, rows, nil
}

func getMachineStatusEmojis(liveliness *string, events *models.V1MachineRecentProvisioningEvents, state *models.V1MachineState) (string, string) {
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

		if events.LastErrorEvent != nil && time.Since(time.Time(events.LastErrorEvent.Time)) < LastErrorEventRelevant {
			emojis = append(emojis, api.Exclamation)
			wide = append(wide, "LastEventErrors")
		}

		if pointer.SafeDeref(events.CrashLoop) {
			emojis = append(emojis, api.Loop)
			wide = append(wide, "CrashLoop")
		}
	}

	return strings.Join(emojis, nbr), strings.Join(wide, ", ")
}

func (t *TablePrinter) MachineIPMITable(data []*models.V1MachineIPMIResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "", "Power", "IP", "Mac", "Board Part Number", "Bios Version", "BMC Version", "Size", "Partition"}
	if wide {
		header = []string{"ID", "Status", "Power", "IP", "Mac", "Board Part Number", "Chassis Serial", "Product Serial", "Bios Version", "BMC Version", "Size", "Partition"}
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
		}

		biosVersion := ""
		bios := machine.Bios
		if bios != nil {
			biosVersion = pointer.SafeDeref(bios.Version)
		}

		emojis, wideEmojis := getMachineStatusEmojis(machine.Liveliness, machine.Events, machine.State)

		if wide {
			rows = append(rows, []string{id, wideEmojis, powerText, ipAddress, mac, bpn, cs, ps, biosVersion, bmcVersion, size, partition})
		} else {
			rows = append(rows, []string{id, emojis, power, ipAddress, mac, bpn, biosVersion, bmcVersion, size, partition})
		}
	}

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

	t.t.GetTable().SetAutoWrapText(false)

	return header, rows, nil
}

func (t *TablePrinter) MachineIssuesTable(data api.MachineIssues, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Power", "Allocated", "", "Lock Reason", "Last Event", "When", "Issues"}
	if wide {
		header = []string{"ID", "Name", "Partition", "Project", "Power", "State", "Lock Reason", "Last Event", "When", "Issues"}
	}

	for _, machineWithIssues := range data {
		machine := machineWithIssues.Machine

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

		var issues []string
		for _, issue := range machineWithIssues.Issues {
			text := fmt.Sprintf("- %s (%s)", issue.Description, issue.ShortName)
			if wide && issue.RefURL != "" {
				text += " (" + issue.RefURL + ")"
			}
			issues = append(issues, text)
		}

		emojis, _ := getMachineStatusEmojis(machine.Liveliness, machine.Events, machine.State)

		if wide {
			rows = append(rows, []string{pointer.SafeDeref(machine.ID), widename, partition, project, powerText, lockText, lockDescWide, lastEvent, when, strings.Join(issues, "\n")})
		} else {
			rows = append(rows, []string{pointer.SafeDeref(machine.ID), power, allocated, emojis, lockDesc, lastEvent, when, strings.Join(issues, "\n")})
		}
	}

	t.t.GetTable().SetAutoWrapText(false)

	return header, rows, nil
}
