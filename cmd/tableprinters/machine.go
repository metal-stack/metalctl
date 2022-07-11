package tableprinters

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/pkg/api"
)

func (t *TablePrinter) MachineTable(data []*models.V1MachineResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "", "", "Last Event", "When", "Age", "Hostname", "Project", "Size", "Image", "Partition"}
	if wide {
		header = []string{"ID", "", "Last Event", "When", "Age", "Description", "Name", "Hostname", "Project", "IPs", "Size", "Image", "Partition", "Started", "Tags", "Lock/Reserve"}
	}

	for _, machine := range data {
		machineID := *machine.ID
		if machine.Ledstate != nil && *machine.Ledstate.Value == "LED-ON" {
			blue := color.New(color.FgBlue).SprintFunc()
			machineID = blue(machineID)
		}

		status := pointer.Deref(machine.Liveliness)
		statusEmoji := ""
		switch status {
		case "Alive":
			statusEmoji = nbr
		case "Dead":
			statusEmoji = skull
		case "Unknown":
			statusEmoji = question
		default:
			statusEmoji = question
		}

		alloc := pointer.Deref(machine.Allocation)
		sizeID := pointer.Deref(pointer.Deref(machine.Size).ID)
		partitionID := pointer.Deref(pointer.Deref(machine.Partition).ID)
		project := pointer.Deref(alloc.Project)
		name := pointer.Deref(alloc.Name)
		desc := alloc.Description
		hostname := pointer.Deref(alloc.Hostname)
		image := pointer.Deref(alloc.Image).Name

		truncatedHostname := truncate(hostname, 30)

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
		lockEmoji := ""
		if *machine.State.Value != "" {
			reserved = fmt.Sprintf("%s:%s", *machine.State.Value, *machine.State.Description)
			if *machine.State.Value == "LOCKED" {
				lockEmoji = lock
			}
			if *machine.State.Value == "RESERVED" {
				lockEmoji = bark
			}
		}
		lastEvent := ""
		lastEventEmoji := ""
		when := ""
		if len(machine.Events.Log) > 0 {
			since := time.Since(time.Time(machine.Events.LastEventTime))
			when = humanizeDuration(since)
			lastEvent = *machine.Events.Log[0].Event
			lastEventEmoji = lastEvent
		}

		if machine.Events.IncompleteProvisioningCycles != nil {
			if *machine.Events.IncompleteProvisioningCycles != "" && *machine.Events.IncompleteProvisioningCycles != "0" {
				lastEvent += " (!)"
				lastEventEmoji += nbr + circle
			}
		}

		if wide {
			rows = append(rows, []string{machineID, status, lastEvent, when, age, desc, name, hostname, project, ips, sizeID, image, partitionID, started, tags, reserved})
		} else {
			rows = append(rows, []string{machineID, lockEmoji, statusEmoji, lastEventEmoji, when, age, truncatedHostname, project, sizeID, image, partitionID})
		}
	}

	return header, rows, nil
}

func (t *TablePrinter) MachineIPMITable(data []*models.V1MachineIPMIResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Status", "Power", "IP", "Mac", "Board Part Number", "Bios Version", "BMC Version", "Size", "Partition"}
	if wide {
		header = []string{"ID", "Status", "Power", "IP", "Mac", "Board Part Number", "Chassis Serial", "Product Serial", "Bios Version", "BMC Version", "Size", "Partition"}
	}

	for _, machine := range data {
		id := pointer.Deref(machine.ID)
		partition := pointer.Deref(pointer.Deref(machine.Partition).ID)
		size := pointer.Deref(pointer.Deref(machine.Size).ID)

		statusEmoji := ""
		if machine.Liveliness != nil {
			switch *machine.Liveliness {
			case "Alive":
				statusEmoji = nbr
			case "Dead":
				statusEmoji = skull
			case "Unknown":
				statusEmoji = question
			default:
				statusEmoji = question
			}
		}

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
			ipAddress = pointer.Deref(ipmi.Address)
			mac = pointer.Deref(ipmi.Mac)
			bmcVersion = pointer.Deref(ipmi.Bmcversion)
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
			biosVersion = pointer.Deref(bios.Version)
		}

		if wide {
			rows = append(rows, []string{id, statusEmoji, powerText, ipAddress, mac, bpn, cs, ps, biosVersion, bmcVersion, size, partition})
		} else {
			rows = append(rows, []string{id, statusEmoji, power, ipAddress, mac, bpn, biosVersion, bmcVersion, size, partition})
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
		rows = append(rows, []string{i.Time.String(), pointer.Deref(i.Event), i.Message})
	}

	t.t.GetTable().SetAutoWrapText(false)

	return header, rows, nil
}

func (t *TablePrinter) MachineIssuesTable(data api.MachineIssues, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Power", "Allocated", "Lock", "Lock Reason", "Status", "Last Event", "When", "Issues"}
	if wide {
		header = []string{"ID", "Name", "Partition", "Project", "Power", "Status", "State", "Lock Reason", "Last Event", "When", "Issues"}
	}

	for id, machineWithIssues := range data {
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

		status := pointer.Deref(machine.Liveliness)
		statusEmoji := ""
		switch status {
		case "Alive":
			statusEmoji = nbr
		case "Dead":
			statusEmoji = skull
		case "Unknown":
			statusEmoji = question
		default:
			statusEmoji = question
		}

		lockEmoji := ""
		lockText := ""
		lockDesc := ""
		lockDescWide := ""
		if machine.State != nil && machine.State.Value != nil && *machine.State.Value != "" {
			if *machine.State.Value == "LOCKED" {
				lockEmoji = lock
			}
			if *machine.State.Value == "RESERVED" {
				lockEmoji = bark
			}
			lockText = *machine.State.Value
		}
		if machine.State != nil && machine.State.Value != nil && *machine.State.Description != "" {
			lockDescWide = *machine.State.Description
			lockDesc = truncateEnd(lockDescWide, 30)
		}

		power, powerText := extractPowerState(machine.Ipmi)

		when := ""
		lastEvent := ""
		lastEventEmoji := ""
		if len(machine.Events.Log) > 0 {
			since := time.Since(time.Time(machine.Events.LastEventTime))
			when = humanizeDuration(since)
			lastEvent = *machine.Events.Log[0].Event
			lastEventEmoji = lastEvent
		}

		var issues []string
		for _, issue := range machineWithIssues.Issues {
			text := fmt.Sprintf("- %s (%s)", issue.Description, issue.ShortName)
			if wide && issue.RefURL != "" {
				text += " (" + issue.RefURL + ")"
			}
			issues = append(issues, text)
		}

		if wide {
			rows = append(rows, []string{id, widename, partition, project, powerText, status, lockText, lockDescWide, lastEvent, when, strings.Join(issues, "\n")})
		} else {
			rows = append(rows, []string{id, power, allocated, lockEmoji, lockDesc, statusEmoji, lastEventEmoji, when, strings.Join(issues, "\n")})
		}
	}

	t.t.GetTable().SetAutoWrapText(false)

	return header, rows, nil
}
