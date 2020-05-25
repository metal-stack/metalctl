package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"
	"time"

	"bytes"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/tag"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

const (
	bark            = "\U0001F6A7"
	blueDiamond     = "\U0001F539"
	circle          = "\U000021BB"
	dot             = "\U000025CF"
	exclamationMark = "\U00002757"
	lock            = "\U0001F512"
	nbr             = "\U00002007"
	question        = "\U00002753"
	skull           = "\U0001F480"
)

type (
	// Printer main Interface for implementations which spits out to stdout
	Printer interface {
		Print(data interface{}) error
	}
	// JSONPrinter returns the model in json format
	JSONPrinter struct{}
	// YAMLPrinter returns the model in yaml format
	YAMLPrinter struct{}
	// TablePrinter produces a human readable model representation
	TablePrinter struct {
		table       *tablewriter.Table
		wide        bool
		order       string
		noHeaders   bool
		template    *template.Template
		shortHeader []string
		wideHeader  []string
		shortData   [][]string
		wideData    [][]string
	}

	// MetalMachineTablePrinter is a table printer for a MetalMachine
	MetalMachineTablePrinter struct {
		TablePrinter
	}
	// MetalMachineLogsPrinter prints machine logs
	MetalMachineLogsPrinter struct {
		TablePrinter
	}

	// MetalFirewallTablePrinter is a table printer for a MetalFirewall
	MetalFirewallTablePrinter struct {
		TablePrinter
	}
	// MetalMachineAllocationTablePrinter is a table printer for a MetalMachine
	MetalMachineAllocationTablePrinter struct {
		TablePrinter
	}
	// MetalSizeTablePrinter is a table printer for a MetalSize
	MetalSizeTablePrinter struct {
		TablePrinter
	}
	// MetalSizeMatchingLogTablePrinter is a table printer for a MetalSizeMatchingLog
	MetalSizeMatchingLogTablePrinter struct {
		TablePrinter
	}
	// MetalImageTablePrinter is a table printer for a MetalImage
	MetalImageTablePrinter struct {
		TablePrinter
	}
	// MetalPartitionTablePrinter is a table printer for a MetalPartition
	MetalPartitionTablePrinter struct {
		TablePrinter
	}
	// MetalPartitionCapacityTablePrinter is a table printer for a MetalPartition
	MetalPartitionCapacityTablePrinter struct {
		TablePrinter
	}
	// MetalSwitchTablePrinter is a table printer for a MetalSwitch
	MetalSwitchTablePrinter struct {
		TablePrinter
	}
	// MetalNetworkTablePrinter is a table printer for a MetalNetwork
	MetalNetworkTablePrinter struct {
		TablePrinter
	}
	// MetalIPTablePrinter is a table printer for a MetalIP
	MetalIPTablePrinter struct {
		TablePrinter
	}
	// MetalProjectTablePrinter is a table printer for a MetalProject
	MetalProjectTablePrinter struct {
		TablePrinter
	}
	// MachineWithIPMIPrinter is a table printer with Machine IPMI data
	MachineWithIPMIPrinter struct {
		TablePrinter
	}

	// ContextPrinter is a table printer with context
	ContextPrinter struct {
		TablePrinter
	}
)

// render the table shortHeader and shortData are always expected.
func (t *TablePrinter) render() {
	if t.template == nil {
		if !t.noHeaders {
			if t.wide {
				t.table.SetHeader(t.wideHeader)
			} else {
				t.table.SetHeader(t.shortHeader)
			}
		}
		if t.wide {
			t.table.AppendBulk(t.wideData)
		} else {
			t.table.AppendBulk(t.shortData)
		}
		t.table.Render()
	} else {
		rows := t.shortData
		if t.wide {
			rows = t.wideData
		}
		for _, row := range rows {
			if len(row) < 1 {
				continue
			}
			fmt.Println(row[0])
		}
	}
}

func (t *TablePrinter) addShortData(row []string, data interface{}) {
	if t.wide {
		return
	}
	t.shortData = append(t.shortData, t.rowOrTemplate(row, data))
}
func (t *TablePrinter) addWideData(row []string, data interface{}) {
	if !t.wide {
		return
	}
	t.wideData = append(t.wideData, t.rowOrTemplate(row, data))
}

// rowOrTemplate return either given row or the data rendered with the given template, depending if template is set.
func (t *TablePrinter) rowOrTemplate(row []string, data interface{}) []string {
	tpl := t.template
	if tpl != nil {
		var buf bytes.Buffer
		err := tpl.Execute(&buf, genericObject(data))
		if err != nil {
			fmt.Printf("unable to parse template:%v", err)
			os.Exit(1)
		}
		return []string{buf.String()}
	}
	return row
}

// NewPrinter returns a suitable stdout printer for the given format
func NewPrinter(format, order, tpl string, noHeaders bool) (Printer, error) {
	var printer Printer
	switch format {
	case "yaml":
		printer = &YAMLPrinter{}
	case "json":
		printer = &JSONPrinter{}
	case "table", "wide", "markdown":
		printer = newTablePrinter(format, order, noHeaders, nil)
	case "template":
		tmpl, err := template.New("").Parse(tpl)
		if err != nil {
			return nil, fmt.Errorf("template invalid:%v", err)
		}
		printer = newTablePrinter(format, order, true, tmpl)
	default:
		return nil, fmt.Errorf("unknown format:%s", format)
	}
	return printer, nil
}

func newTablePrinter(format, order string, noHeaders bool, template *template.Template) TablePrinter {
	table := tablewriter.NewWriter(os.Stdout)
	wide := false
	if format == "wide" {
		wide = true
	}
	switch format {
	case "markdown":
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
	default:
		table.SetHeaderLine(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetBorder(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetRowLine(false)
		table.SetTablePadding("\t") // pad with tabs
		table.SetNoWhiteSpace(true) // no whitespace in front of every line
	}
	return TablePrinter{
		table:     table,
		wide:      wide,
		order:     order,
		noHeaders: noHeaders,
		template:  template,
	}
}

// Print a model in json format
func (j JSONPrinter) Print(data interface{}) error {
	json, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("unable to marshal to json:%v", err)
	}
	fmt.Printf("%s\n", string(json))
	return nil
}

// Print a model in yaml format
func (y YAMLPrinter) Print(data interface{}) error {
	yml, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal to yaml:%v", err)
	}
	fmt.Printf("%s\n", string(yml))
	return nil
}

// Print a model in yaml format
func (m ContextPrinter) Print(data *Contexts) error {
	for name, c := range data.Contexts {
		if name == data.CurrentContext {
			name = name + " [*]"
		}
		row := []string{name, c.ApiURL, c.IssuerURL}
		m.addShortData(row, c)
	}
	m.shortHeader = []string{"Name", "URL", "DEX"}
	m.render()
	return nil
}

// Print a model in a human readable table
func (t TablePrinter) Print(data interface{}) error {
	switch d := data.(type) {
	case []*models.V1MachineResponse:
		MetalMachineTablePrinter{t}.Print(d)
	case *models.V1MachineResponse:
		MetalMachineTablePrinter{t}.Print([]*models.V1MachineResponse{d})
	case []*models.V1FirewallResponse:
		MetalFirewallTablePrinter{t}.Print(d)
	case *models.V1FirewallResponse:
		MetalFirewallTablePrinter{t}.Print([]*models.V1FirewallResponse{d})
	case []*models.V1SizeResponse:
		MetalSizeTablePrinter{t}.Print(d)
	case []*models.V1SizeMatchingLog:
		MetalSizeMatchingLogTablePrinter{t}.Print(d)
	case []*models.V1ImageResponse:
		MetalImageTablePrinter{t}.Print(d)
	case []*models.V1PartitionResponse:
		MetalPartitionTablePrinter{t}.Print(d)
	case []*models.V1PartitionCapacity:
		MetalPartitionCapacityTablePrinter{t}.Print(d)
	case []*models.V1SwitchResponse:
		MetalSwitchTablePrinter{t}.Print(d)
	case []*models.V1NetworkResponse:
		MetalNetworkTablePrinter{t}.Print(d)
	case []*models.V1IPResponse:
		MetalIPTablePrinter{t}.Print(d)
	case *models.V1ProjectResponse:
		MetalProjectTablePrinter{t}.Print([]*models.V1ProjectResponse{d})
	case []*models.V1ProjectResponse:
		MetalProjectTablePrinter{t}.Print(d)
	case []*models.V1MachineIPMIResponse:
		MachineWithIPMIPrinter{t}.Print(d)
	case *models.V1MachineIPMIResponse:
		MachineWithIPMIPrinter{t}.Print([]*models.V1MachineIPMIResponse{d})
	case []*models.V1MachineProvisioningEvent:
		MetalMachineLogsPrinter{t}.Print(d)
	case *Contexts:
		ContextPrinter{t}.Print(d)
	default:
		return fmt.Errorf("unknown table printer for type: %T", d)
	}
	return nil
}

// Order machines
func (m MetalMachineTablePrinter) Order(data []*models.V1MachineResponse) {
	cols := strings.Split(m.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "size":
					if A.Size == nil || A.Size.ID == nil {
						return true
					}
					if B.Size == nil || B.Size.ID == nil {
						return false
					}
					if *A.Size.ID < *B.Size.ID {
						return true
					}
					if *A.Size.ID != *B.Size.ID {
						return false
					}
				case "id":
					if A.ID == nil {
						return true
					}
					if B.ID == nil {
						return false
					}
					if *A.ID < *B.ID {
						return true
					}
					if *A.ID != *B.ID {
						return false
					}
				case "status":
					if A.Liveliness == nil {
						return true
					}
					if B.Liveliness == nil {
						return false
					}
					if *A.Liveliness < *B.Liveliness {
						return true
					}
					if *A.Liveliness != *B.Liveliness {
						return false
					}
				case "event":
					if A.Events == nil || len(A.Events.Log) == 0 || A.Events.Log[0].Event == nil {
						return true
					}
					if B.Events == nil || len(B.Events.Log) == 0 || B.Events.Log[0].Event == nil {
						return false
					}
					if *A.Events.Log[0].Event < *B.Events.Log[0].Event {
						return true
					}
					if *A.Events.Log[0].Event != *B.Events.Log[0].Event {
						return false
					}
				case "when":
					if A.Events == nil || A.Events.LastEventTime == nil {
						return true
					}
					if B.Events == nil || B.Events.LastEventTime == nil {
						return false
					}
					if time.Time(*A.Events.LastEventTime).After(time.Time(*B.Events.LastEventTime)) {
						return true
					}
					if *A.Events.LastEventTime != *B.Events.LastEventTime {
						return false
					}
				case "partition":
					if A.Partition == nil || A.Partition.ID == nil {
						return true
					}
					if B.Partition == nil || B.Partition.ID == nil {
						return true
					}
					if *A.Partition.ID < *B.Partition.ID {
						return true
					}
					if *A.Partition.ID != *B.Partition.ID {
						return false
					}
				case "project":
					if A.Allocation == nil || B.Allocation == nil {
						return true
					}
					if *A.Allocation.Project < *B.Allocation.Project {
						return true
					}
					if *A.Allocation.Project != *B.Allocation.Project {
						return false
					}
				}
			}

			return false
		})
	}
}

// Order ipmi data from machines
func (m MachineWithIPMIPrinter) Order(data []*models.V1MachineIPMIResponse) {
	sort.SliceStable(data, func(i, j int) bool {
		if data[i].IPMI.Address != nil && data[j].IPMI.Address != nil {
			return *data[i].IPMI.Address < *data[j].IPMI.Address
		}
		return false
	})
	sort.SliceStable(data, func(i, j int) bool {
		if data[i].Partition != nil && data[j].Partition != nil {
			return *data[i].Partition.ID < *data[j].Partition.ID
		}
		return false
	})

	cols := strings.Split(m.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "size":
					if A.Size == nil || A.Size.ID == nil {
						return true
					}
					if B.Size == nil || B.Size.ID == nil {
						return false
					}
					if *A.Size.ID < *B.Size.ID {
						return true
					}
					if *A.Size.ID != *B.Size.ID {
						return false
					}
				case "id":
					if A.ID == nil {
						return true
					}
					if B.ID == nil {
						return false
					}
					if *A.ID < *B.ID {
						return true
					}
					if *A.ID != *B.ID {
						return false
					}
				case "status":
					if A.Liveliness == nil {
						return true
					}
					if B.Liveliness == nil {
						return false
					}
					if *A.Liveliness < *B.Liveliness {
						return true
					}
					if *A.Liveliness != *B.Liveliness {
						return false
					}
				case "event":
					if A.Events == nil || len(A.Events.Log) == 0 || A.Events.Log[0].Event == nil {
						return true
					}
					if B.Events == nil || len(B.Events.Log) == 0 || B.Events.Log[0].Event == nil {
						return false
					}
					if *A.Events.Log[0].Event < *B.Events.Log[0].Event {
						return true
					}
					if *A.Events.Log[0].Event != *B.Events.Log[0].Event {
						return false
					}
				case "when":
					if A.Events == nil || A.Events.LastEventTime == nil {
						return true
					}
					if B.Events == nil || B.Events.LastEventTime == nil {
						return false
					}
					if time.Time(*A.Events.LastEventTime).After(time.Time(*B.Events.LastEventTime)) {
						return true
					}
					if *A.Events.LastEventTime != *B.Events.LastEventTime {
						return false
					}
				case "partition":
					if A.Partition == nil || A.Partition.ID == nil {
						return true
					}
					if B.Partition == nil || B.Partition.ID == nil {
						return true
					}
					if *A.Partition.ID < *B.Partition.ID {
						return true
					}
					if *A.Partition.ID != *B.Partition.ID {
						return false
					}
				case "project":
					if A.Allocation == nil || B.Allocation == nil {
						return true
					}
					if *A.Allocation.Project < *B.Allocation.Project {
						return true
					}
					if *A.Allocation.Project != *B.Allocation.Project {
						return false
					}
				}
			}

			return false
		})
	}
}

// Print a list of Machines in a table
func (m MetalMachineTablePrinter) Print(data []*models.V1MachineResponse) {
	m.Order(data)
	m.shortHeader = []string{"ID", "", "", "Last Event", "When", "Age", "Hostname", "Project", "Size", "Image", "Partition"}
	m.wideHeader = []string{"ID", "", "Last Event", "When", "Age", "Description", "Name", "Hostname", "Project", "IPs", "Size", "Image", "Partition", "Started", "Console Password", "Tags", "Lock/Reserve"}
	atLeastOneLEDOn := false
	for _, machine := range data {
		if machine.Ledstate != nil && *machine.Ledstate.Value == "LED-ON" {
			atLeastOneLEDOn = true
			break
		}
	}
	for _, machine := range data {
		machineID := *machine.ID
		led := ""
		if machine.Ledstate != nil && *machine.Ledstate.Value == "LED-ON" {
			led = blueDiamond
		}

		alloc := machine.Allocation
		if alloc == nil {
			alloc = &models.V1MachineAllocation{}
		}
		status := strValue(machine.Liveliness)
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
		var sizeID string
		if machine.Size != nil {
			sizeID = strValue(machine.Size.ID)
		}
		var partitionID string
		if machine.Partition != nil {
			partitionID = strValue(machine.Partition.ID)
		}
		project := strValue(alloc.Project)
		name := strValue(alloc.Name)
		hostname := strValue(alloc.Hostname)
		truncatedHostname := truncate(hostname, "...", 30)

		var nwIPs []string
		for _, nw := range alloc.Networks {
			nwIPs = append(nwIPs, nw.Ips...)
		}
		ips := strings.Join(nwIPs, "\n")
		image := ""
		if alloc.Image != nil {
			image = alloc.Image.Name
		}
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
			since := time.Since(time.Time(*machine.Events.LastEventTime))
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
		desc := alloc.Description
		if desc == "" {
			desc = machine.Description
		}

		row := []string{machineID, lockEmoji, statusEmoji, lastEventEmoji, when, age, truncatedHostname, project, sizeID, image, partitionID}
		wide := []string{machineID, status, lastEvent, when, age, desc, name, hostname, project, ips, sizeID, image, partitionID, started, alloc.ConsolePassword, tags, reserved}

		if atLeastOneLEDOn {
			m.shortHeader = append([]string{""}, m.shortHeader...)
			m.wideHeader = append([]string{""}, m.wideHeader...)
			row = append([]string{led}, row...)
			wide = append([]string{led}, wide...)
		}

		m.addShortData(row, machine)
		m.addWideData(wide, machine)
	}
	m.render()
}

// Print a MetalFirewall in a table
func (m MetalFirewallTablePrinter) Print(data []*models.V1FirewallResponse) {
	for _, firewall := range data {
		if firewall == nil {
			continue
		}
		alloc := firewall.Allocation
		partition := strValue(firewall.Partition.ID)
		project := strValue(alloc.Project)
		hostname := strValue(alloc.Hostname)

		var nwIPs []string
		var nws []string
		for _, nw := range alloc.Networks {
			nwIPs = append(nwIPs, nw.Ips...)
			nws = append(nws, *nw.Networkid)
		}
		ips := strings.Join(nwIPs, "\n")
		networks := strings.Join(nws, "\n")

		firewallID := *firewall.ID

		age := ""
		if alloc.Created != nil && !time.Time(*alloc.Created).IsZero() {
			age = humanizeDuration(time.Since(time.Time(*alloc.Created)))
		}

		row := []string{firewallID, age, hostname, project, networks, ips, partition}
		m.addShortData(row, firewall)
	}
	m.shortHeader = []string{"ID", "AGE", "Hostname", "Project", "Networks", "IPs", "Partition"}
	m.render()
}

// Print a MetalSize in a table
func (m MetalSizeTablePrinter) Print(data []*models.V1SizeResponse) {
	sort.SliceStable(data, func(i, j int) bool { return *data[i].ID < *data[j].ID })
	for _, size := range data {
		id := strValue(size.ID)
		cs := size.Constraints
		var cpu, memory, storage string
		for _, c := range cs {
			switch *c.Type {
			case "cores":
				cpu = fmt.Sprintf("%d - %d", *c.Min, *c.Max)
			case "memory":
				memory = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)))
			case "storage":
				storage = fmt.Sprintf("%s - %s", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)))
			}
		}
		row := []string{id, size.Name, size.Description, cpu, memory, storage}
		m.addShortData(row, size)
	}
	m.shortHeader = []string{"ID", "Name", "Description", "CPU Range", "Memory Range", "Storage Range"}
	m.render()
}

// Print a MetalSize in a table
func (m MetalSizeMatchingLogTablePrinter) Print(data []*models.V1SizeMatchingLog) {
	for _, d := range data {
		var cpu, memory, storage string
		for _, cs := range d.Constraints {
			c := cs.Constraint
			switch *c.Type {
			case "cores":
				cpu = fmt.Sprintf("%d - %d\n%s\nmatches: %v", *c.Min, *c.Max, *cs.Log, *cs.Match)
			case "memory":
				memory = fmt.Sprintf("%s - %s\n%s\nmatches: %v", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)), *cs.Log, *cs.Match)
			case "storage":
				storage = fmt.Sprintf("%s - %s\n%s\nmatches: %v", humanize.Bytes(uint64(*c.Min)), humanize.Bytes(uint64(*c.Max)), *cs.Log, *cs.Match)
			}
		}
		sizeMatch := fmt.Sprintf("%v", *d.Match)
		row := []string{*d.Name, sizeMatch, cpu, memory, storage}
		m.addShortData(row, d)
	}
	m.shortHeader = []string{"Name", "Match", "CPU Constraint", "Memory Constraint", "Storage Constraint"}
	m.table.SetAutoWrapText(false)
	m.table.SetColMinWidth(3, 40)
	m.render()
}

// Print a MetalImage in a table
func (m MetalImageTablePrinter) Print(data []*models.V1ImageResponse) {
	sort.SliceStable(data, func(i, j int) bool { return *data[i].ID < *data[j].ID })
	for _, image := range data {
		id := strValue(image.ID)
		features := strings.Join(image.Features, ",")
		name := image.Name
		description := image.Description
		expiration := ""
		if image.ExpirationDate != nil {
			expiration = humanizeDuration(time.Until(time.Time(*image.ExpirationDate)))
		}
		status := image.Classification
		row := []string{id, name, description, features, expiration, status}
		m.addShortData(row, image)
	}
	m.shortHeader = []string{"ID", "Name", "Description", "Features", "Expiration", "Status"}
	m.render()
}

// Print a MetalPartition in a table
func (m MetalPartitionTablePrinter) Print(data []*models.V1PartitionResponse) {
	for _, p := range data {
		id := strValue(p.ID)
		row := []string{id, p.Name, p.Description}
		m.addShortData(row, p)
	}
	m.shortHeader = []string{"ID", "Name", "Description"}
	m.render()
}

// Print a PartitionCapacity in a table
func (m MetalPartitionCapacityTablePrinter) Print(pcs []*models.V1PartitionCapacity) {
	sort.SliceStable(pcs, func(i, j int) bool { return *pcs[i].ID < *pcs[j].ID })
	total := int32(0)
	free := int32(0)
	allocated := int32(0)
	faulty := int32(0)
	other := int32(0)
	for _, pc := range pcs {
		sort.SliceStable(pc.Servers, func(i, j int) bool { return *pc.Servers[i].Size < *pc.Servers[j].Size })
		for _, c := range pc.Servers {
			id := strValue(c.Size)
			row := []string{*pc.ID, id, fmt.Sprintf("%d", *c.Total), fmt.Sprintf("%d", *c.Free), fmt.Sprintf("%d", *c.Allocated), fmt.Sprintf("%d", *c.Other), fmt.Sprintf("%d", *c.Faulty)}
			total += *c.Total
			free += *c.Free
			allocated += *c.Allocated
			other += *c.Other
			faulty += *c.Faulty
			m.addShortData(row, pc)
		}
	}
	footerRow := ([]string{"Total", "", fmt.Sprintf("%d", total), fmt.Sprintf("%d", free), fmt.Sprintf("%d", allocated), fmt.Sprintf("%d", other), fmt.Sprintf("%d", faulty)})
	m.addShortData(footerRow, nil)
	m.shortHeader = []string{"Partition", "Size", "Total", "Free", "Allocated", "Other", "Faulty"}
	m.render()
}

// Print all MetalSwitch(es) in a table
func (m MetalSwitchTablePrinter) Print(data []*models.V1SwitchResponse) {
	sort.SliceStable(data, func(i, j int) bool { return *data[i].ID < *data[j].ID })
	for _, s := range data {
		id := strValue(s.ID)
		partition := ""
		if s.Partition != nil {
			partition = strValue(s.Partition.ID)
		}
		rack := strValue(s.RackID)

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
			if errorTime.After(syncTime) {
				syncError = fmt.Sprintf("%s ago: %s", humanizeDuration(time.Since(errorTime)), strValue(s.LastSyncError.Error))
			}
		}

		row := []string{id, partition, rack, shortStatus}
		wide := []string{id, partition, rack, syncAgeStr, syncDurStr, syncError}
		m.addShortData(row, s)
		m.addWideData(wide, s)
	}
	m.shortHeader = []string{"ID", "Partition", "Rack", "Status"}
	m.wideHeader = []string{"ID", "Partition", "Rack", "Last Sync", "Sync Duration", "Last Sync Error"}
	m.render()
}

type network struct {
	parent   *models.V1NetworkResponse
	children []*models.V1NetworkResponse
}

type networks []*network

func (nn *networks) appendChild(parentID string, child *models.V1NetworkResponse) bool {
	for _, n := range *nn {
		if *n.parent.ID == parentID {
			n.children = append(n.children, child)
			return true
		}
	}
	return false
}

// Print all MetalNetworks in a table
func (m MetalNetworkTablePrinter) Print(data []*models.V1NetworkResponse) {
	sort.SliceStable(data, func(i, j int) bool { return *data[i].ID < *data[j].ID })

	nn := &networks{}
	for _, n := range data {
		if n.Parentnetworkid == nil {
			*nn = append(*nn, &network{parent: n})
		}
	}
	for _, n := range data {
		if n.Parentnetworkid != nil {
			if !nn.appendChild(*n.Parentnetworkid, n) {
				*nn = append(*nn, &network{parent: n})
			}
		}
	}
	for _, n := range *nn {
		m.addNetwork("", n.parent)
		for i, c := range n.children {
			prefix := "\u251C"
			if i == len(n.children)-1 {
				prefix = "\u2514"
			}
			prefix += "\u2500\u2574"
			m.addNetwork(prefix, c)
		}
	}
	m.shortHeader = []string{"ID", "Name", "Project", "Partition", "Nat", "Prefixes", "", "IPs"}
	m.wideHeader = []string{"ID", "Description", "Name", "Project", "Partition", "Nat", "Prefixes", "Usage", "PrivateSuper"}
	m.render()
}

func (m *MetalNetworkTablePrinter) addNetwork(prefix string, n *models.V1NetworkResponse) {
	id := fmt.Sprintf("%s%s", prefix, strValue(n.ID))

	prefixes := strings.Join(n.Prefixes, ",")
	flag := false
	if n.Privatesuper != nil {
		flag = *n.Privatesuper
	}
	privateSuper := fmt.Sprintf("%t", flag)
	nat := fmt.Sprintf("%t", *n.Nat)

	usage := fmt.Sprintf("IPs:     %v/%v", *n.Usage.UsedIps, *n.Usage.AvailableIps)

	ipUse := float64(*n.Usage.UsedIps) / float64(*n.Usage.AvailableIps)
	shortIPUsage := nbr
	if ipUse >= 0.9 {
		shortIPUsage += color.RedString(dot)
	} else if ipUse >= 0.7 {
		shortIPUsage += color.YellowString(dot)
	} else {
		shortIPUsage += color.GreenString(dot)
	}

	shortPrefixUsage := ""
	if *n.Usage.AvailablePrefixes > 0 {
		prefixUse := float64(*n.Usage.UsedPrefixes) / float64(*n.Usage.AvailablePrefixes)
		if prefixUse >= 0.9 {
			shortPrefixUsage = exclamationMark
		}
		usage = fmt.Sprintf("%s\nPrefixes:%v/%v", usage, *n.Usage.UsedPrefixes, *n.Usage.AvailablePrefixes)
	}

	max := getMaxLineCount(n.Description, n.Name, n.Projectid, n.Partitionid, nat, prefixes, usage, privateSuper)
	for i := 0; i < max; i++ {
		id += "\n\u2502"
	}
	shortRow := []string{id, n.Name, n.Projectid, n.Partitionid, nat, prefixes, shortPrefixUsage, shortIPUsage}
	wideRow := []string{id, n.Description, n.Name, n.Projectid, n.Partitionid, nat, prefixes, usage, privateSuper}
	m.addShortData(shortRow, n)
	m.addWideData(wideRow, n)
}

func getMaxLineCount(ss ...string) int {
	max := 0
	for _, s := range ss {
		c := strings.Count(s, "\n")
		if c > max {
			max = c
		}
	}
	return max
}

// Print all MetalIPs in a table
func (m MetalIPTablePrinter) Print(data []*models.V1IPResponse) {
	data = sortIPs(data)
	for _, i := range data {
		ipaddress := strValue(i.Ipaddress)
		ipType := strValue(i.Type)
		network := strValue(i.Networkid)
		project := strValue(i.Projectid)
		var shortTags []string
		for _, t := range i.Tags {
			parts := strings.Split(t, "=")
			if strings.HasPrefix(t, tag.MachineID+"=") {
				shortTags = append(shortTags, "machine:"+parts[1])
			} else if strings.HasPrefix(t, tag.ClusterServiceFQN+"=") {
				shortTags = append(shortTags, "service:"+parts[1])
			} else {
				shortTags = append(shortTags, t)
			}
		}
		name := truncate(i.Name, "...", 30)
		description := truncate(i.Description, "...", 30)
		row := []string{ipaddress, description, name, network, project, ipType, strings.Join(shortTags, "\n")}
		wide := []string{ipaddress, i.Description, i.Name, network, project, ipType, strings.Join(i.Tags, "\n")}
		m.addShortData(row, i)
		m.addWideData(wide, i)
	}
	m.shortHeader = []string{"IP", "Description", "Name", "Network", "Project", "Type", "Tags"}
	m.wideHeader = []string{"IP", "Description", "Name", "Network", "Project", "Type", "Tags"}
	m.render()
}

// Print ipmi data from machines
func (m MachineWithIPMIPrinter) Print(data []*models.V1MachineIPMIResponse) {
	m.Order(data)
	for _, i := range data {
		id := strValue(i.ID)
		partition := ""
		if i.Partition != nil {
			partition = strValue(i.Partition.ID)
		}

		size := ""
		if i.Size != nil {
			size = strValue(i.Size.ID)
		}

		statusEmoji := ""
		if i.Liveliness != nil {
			switch *i.Liveliness {
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
		ipmi := i.IPMI
		if ipmi != nil {
			ipAddress = strValue(ipmi.Address)
			mac = strValue(ipmi.Mac)
			bmcVersion = strValue(ipmi.Bmcversion)
			fru := ipmi.Fru
			if fru != nil {
				bpn = fru.BoardPartNumber
				cs = fru.ChassisPartSerial
				ps = fru.ProductSerial
			}
		}
		biosVersion := ""
		bios := i.Bios
		if bios != nil {
			biosVersion = strValue(bios.Version)
		}

		row := []string{id, statusEmoji, ipAddress, mac, bpn, biosVersion, bmcVersion, size, partition}
		wide := []string{id, statusEmoji, ipAddress, mac, bpn, cs, ps, biosVersion, bmcVersion, size, partition}
		m.addShortData(row, m)
		m.addWideData(wide, i)
	}
	m.shortHeader = []string{"ID", "", "IP", "Mac", "Board Part Number", "Bios Version", "BMC Version", "Size", "Partition"}
	m.wideHeader = []string{"ID", "", "IP", "Mac", "Board Part Number", "Chassis Serial", "Product Serial", "Bios Version", "BMC Version", "Size", "Partition"}
	m.render()
}

// Print machine logs
func (m MetalMachineLogsPrinter) Print(data []*models.V1MachineProvisioningEvent) {
	for _, i := range data {
		row := []string{i.Time.String(), strValue(i.Event), i.Message}
		m.addShortData(row, m)
	}
	m.shortHeader = []string{"Time", "Event", "Message"}
	m.table.SetAutoWrapText(false)
	m.render()
}

// Print all MetalIPs in a table
func (m MetalProjectTablePrinter) Print(data []*models.V1ProjectResponse) {
	for _, i := range data {
		id := ""
		if i.Meta != nil {
			id = i.Meta.ID
		}
		name := i.Name
		description := i.Description
		tenant := i.TenantID
		row := []string{id, name, tenant}
		wide := []string{id, name, description, tenant}
		m.addShortData(row, i)
		m.addWideData(wide, i)
	}
	m.shortHeader = []string{"ID", "Name", "Tenant"}
	m.wideHeader = []string{"ID", "Name", "Description", "Tenant"}
	m.render()
}
