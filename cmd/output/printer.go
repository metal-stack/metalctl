package output

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"text/template"
	"time"

	"bytes"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/tag"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	bark            = "🚧"
	circle          = "↻"
	dot             = "●"
	exclamationMark = "❗"
	lock            = "🔒"
	nbr             = " "
	question        = "❓"
	skull           = "💀"
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
	// MetalMachineIssuesTablePrinter is a table printer for a MetalMachine issues
	MetalMachineIssuesTablePrinter struct {
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
	// MetalFirmwaresPrinter prints firmwares
	MetalFirmwaresPrinter struct {
		TablePrinter
	}

	// FilesystemLayoutPrinter is a table printer for Filesystemlayouts
	FilesystemLayoutPrinter struct {
		TablePrinter
	}
	// ContextPrinter is a table printer with context
	ContextPrinter struct {
		TablePrinter
	}
)

// New returns a suitable stdout printer for the given format
func New() Printer {
	printer, err := newPrinter(
		viper.GetString("output-format"),
		viper.GetString("order"),
		viper.GetString("template"),
		viper.GetBool("no-headers"),
	)
	if err != nil {
		log.Fatalf("unable to initialize printer:%v", err)
	}

	if viper.IsSet("force-color") {
		enabled := viper.GetBool("force-color")
		if enabled {
			color.NoColor = false
		} else {
			color.NoColor = true
		}
	}
	return printer
}

// render the table shortHeader and shortData are always expected.
func (t *TablePrinter) render() {
	if t.template != nil {
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
		return
	}

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
func newPrinter(format, order, tpl string, noHeaders bool) (Printer, error) {
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
			return nil, fmt.Errorf("template invalid:%w", err)
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
		return fmt.Errorf("unable to marshal to json:%w", err)
	}
	fmt.Printf("%s\n", string(json))
	return nil
}

// Print a model in yaml format
func (y YAMLPrinter) Print(data interface{}) error {
	yml, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("unable to marshal to yaml:%w", err)
	}
	fmt.Printf("%s", string(yml))
	return nil
}

// Print a model in yaml format
func (m ContextPrinter) Print(data *api.Contexts) error {
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
	case api.MachineIssues:
		MetalMachineIssuesTablePrinter{t}.Print(d)
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
	case *models.V1FirmwaresResponse:
		MetalFirmwaresPrinter{t}.Print(d)
	case *models.V1FilesystemLayoutResponse:
		FilesystemLayoutPrinter{t}.Print([]*models.V1FilesystemLayoutResponse{d})
	case []*models.V1FilesystemLayoutResponse:
		FilesystemLayoutPrinter{t}.Print(d)
	case *api.Contexts:
		return ContextPrinter{t}.Print(d)
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
					if A.Events == nil {
						return true
					}
					if B.Events == nil {
						return false
					}
					if time.Time(A.Events.LastEventTime).After(time.Time(B.Events.LastEventTime)) {
						return true
					}
					if A.Events.LastEventTime != B.Events.LastEventTime {
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

// Order Project
func (m MetalProjectTablePrinter) Order(data []*models.V1ProjectResponse) {
	cols := strings.Split(m.order, ",")
	if len(cols) > 0 {
		sort.SliceStable(data, func(i, j int) bool {
			A := data[i]
			B := data[j]
			for _, order := range cols {
				order = strings.ToLower(order)
				switch order {
				case "tenant":
					if A.TenantID == "" {
						return true
					}
					if B.TenantID == "" {
						return false
					}
					if A.TenantID < B.TenantID {
						return true
					}
					if A.TenantID != B.TenantID {
						return false
					}
				case "project":
					if A.Name == "" {
						return true
					}
					if B.Name == "" {
						return false
					}
					if A.Name < B.Name {
						return true
					}
					if A.Name != B.Name {
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
		if data[i].Ipmi.Address != nil && data[j].Ipmi.Address != nil {
			return *data[i].Ipmi.Address < *data[j].Ipmi.Address
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
					if A.Events == nil {
						return true
					}
					if B.Events == nil {
						return false
					}
					if time.Time(A.Events.LastEventTime).After(time.Time(B.Events.LastEventTime)) {
						return true
					}
					if A.Events.LastEventTime != B.Events.LastEventTime {
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
	m.wideHeader = []string{"ID", "", "Last Event", "When", "Age", "Description", "Name", "Hostname", "Project", "IPs", "Size", "Image", "Partition", "Started", "Tags", "Lock/Reserve"}
	for _, machine := range data {
		machineID := *machine.ID
		if machine.Ledstate != nil && *machine.Ledstate.Value == "LED-ON" {
			blue := color.New(color.FgBlue).SprintFunc()
			machineID = blue(machineID)
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
		truncatedHostname := truncate(hostname, 30)

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
		desc := alloc.Description
		if desc == "" {
			desc = machine.Description
		}

		row := []string{machineID, lockEmoji, statusEmoji, lastEventEmoji, when, age, truncatedHostname, project, sizeID, image, partitionID}
		wide := []string{machineID, status, lastEvent, when, age, desc, name, hostname, project, ips, sizeID, image, partitionID, started, tags, reserved}

		m.addShortData(row, machine)
		m.addWideData(wide, machine)
	}
	m.render()
}

// Print a MetalSize in a table
func (m MetalMachineIssuesTablePrinter) Print(data api.MachineIssues) {
	m.shortHeader = []string{"ID", "Power", "Lock", "Lock Reason", "Status", "Last Event", "When", "Issues"}
	m.wideHeader = []string{"ID", "Name", "Partition", "Project", "Power", "Status", "State", "Lock Reason", "Last Event", "When", "Issues"}

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
			if m.wide && issue.RefURL != "" {
				text += " (" + issue.RefURL + ")"
			}
			issues = append(issues, text)
		}

		row := []string{id, power, lockEmoji, lockDesc, statusEmoji, lastEventEmoji, when, strings.Join(issues, "\n")}
		widerow := []string{id, widename, partition, project, powerText, status, lockText, lockDescWide, lastEvent, when, strings.Join(issues, "\n")}

		m.addShortData(row, m)
		m.addWideData(widerow, m)
	}

	m.table.SetAutoWrapText(false)
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
		m.addWideData(row, firewall)
	}
	m.shortHeader = []string{"ID", "AGE", "Hostname", "Project", "Networks", "IPs", "Partition"}
	m.shortHeader = m.wideHeader
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
		m.addWideData(row, size)
	}
	m.shortHeader = []string{"ID", "Name", "Description", "CPU Range", "Memory Range", "Storage Range"}
	m.shortHeader = m.wideHeader
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
		m.addWideData(row, d)
	}
	m.shortHeader = []string{"Name", "Match", "CPU Constraint", "Memory Constraint", "Storage Constraint"}
	m.wideHeader = m.shortHeader
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
		usedBy := fmt.Sprintf("%d", len(image.Usedby))
		if m.wide {
			usedBy = strings.Join(image.Usedby, "\n")
		}

		row := []string{id, name, description, features, expiration, status, usedBy}
		m.addShortData(row, image)
		m.addWideData(row, image)
	}
	m.shortHeader = []string{"ID", "Name", "Description", "Features", "Expiration", "Status", "UsedBy"}
	m.wideHeader = []string{"ID", "Name", "Description", "Features", "Expiration", "Status", "UsedBy"}
	m.render()
}

// Print a MetalPartition in a table
func (m MetalPartitionTablePrinter) Print(data []*models.V1PartitionResponse) {
	for _, p := range data {
		id := strValue(p.ID)
		row := []string{id, p.Name, p.Description}
		m.addShortData(row, p)
		m.addWideData(row, p)
	}
	m.shortHeader = []string{"ID", "Name", "Description"}
	m.wideHeader = m.shortHeader
	m.render()
}

// Print a PartitionCapacity in a table
func (m MetalPartitionCapacityTablePrinter) Print(pcs []*models.V1PartitionCapacity) {
	sort.SliceStable(pcs, func(i, j int) bool { return *pcs[i].ID < *pcs[j].ID })
	totalCount := int32(0)
	freeCount := int32(0)
	allocatedCount := int32(0)
	faultyCount := int32(0)
	otherCount := int32(0)
	for _, pc := range pcs {
		pc := pc
		sort.SliceStable(pc.Servers, func(i, j int) bool { return *pc.Servers[i].Size < *pc.Servers[j].Size })
		for _, c := range pc.Servers {
			id := strValue(c.Size)
			allocated := fmt.Sprintf("%d", *c.Allocated)
			total := fmt.Sprintf("%d", *c.Total)
			free := fmt.Sprintf("%d", *c.Free)
			faulty := fmt.Sprintf("%d", *c.Faulty)
			other := fmt.Sprintf("%d", *c.Other)
			if m.wide {
				if len(c.Faultymachines) > 0 {
					faulty = strings.Join(c.Faultymachines, "\n")
				}
				if len(c.Othermachines) > 0 {
					other = strings.Join(c.Othermachines, "\n")
				}
			}
			row := []string{*pc.ID, id, total, free, allocated, other, faulty}
			totalCount += *c.Total
			freeCount += *c.Free
			allocatedCount += *c.Allocated
			otherCount += *c.Other
			faultyCount += *c.Faulty
			m.addShortData(row, pc)
			m.addWideData(row, pc)
		}
	}
	footerRow := ([]string{"Total", "", fmt.Sprintf("%d", totalCount), fmt.Sprintf("%d", freeCount), fmt.Sprintf("%d", allocatedCount), fmt.Sprintf("%d", otherCount), fmt.Sprintf("%d", faultyCount)})
	m.addShortData(footerRow, nil)
	m.addWideData(footerRow, nil)
	m.shortHeader = []string{"Partition", "Size", "Total", "Free", "Allocated", "Other", "Faulty"}
	m.wideHeader = m.shortHeader
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

		row := []string{id, partition, rack, shortStatus}
		wide := []string{id, partition, rack, mode, syncAgeStr, syncDurStr, syncError}
		m.addShortData(row, s)
		m.addWideData(wide, s)
	}
	m.shortHeader = []string{"ID", "Partition", "Rack", "Status"}
	m.wideHeader = []string{"ID", "Partition", "Rack", "Mode", "Last Sync", "Sync Duration", "Last Sync Error"}
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
		if n.Parentnetworkid == "" {
			*nn = append(*nn, &network{parent: n})
		}
	}
	for _, n := range data {
		if n.Parentnetworkid != "" {
			if !nn.appendChild(n.Parentnetworkid, n) {
				*nn = append(*nn, &network{parent: n})
			}
		}
	}
	for _, n := range *nn {
		m.addNetwork("", n.parent)
		for i, c := range n.children {
			prefix := "├"
			if i == len(n.children)-1 {
				prefix = "└"
			}
			prefix += "─╴"
			m.addNetwork(prefix, c)
		}
	}
	m.shortHeader = []string{"ID", "Name", "Project", "Partition", "Nat", "Shared", "Prefixes", "", "IPs"}
	m.wideHeader = []string{"ID", "Description", "Name", "Project", "Partition", "Nat", "Shared", "Prefixes", "Usage", "PrivateSuper", "Annotations"}
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
		usage = fmt.Sprintf("%s\nPrefixes:%d/%d", usage, *n.Usage.UsedPrefixes, *n.Usage.AvailablePrefixes)
	}

	max := getMaxLineCount(n.Description, n.Name, n.Projectid, n.Partitionid, nat, prefixes, usage, privateSuper)
	for i := 0; i < max-1; i++ {
		id += "\n│"
	}

	var as []string
	for k, v := range n.Labels {
		as = append(as, k+"="+v)
	}
	shared := "false"
	if n.Shared {
		shared = "true"
	}
	annotations := strings.Join(as, "\n")
	shortRow := []string{id, n.Name, n.Projectid, n.Partitionid, nat, shared, prefixes, shortPrefixUsage, shortIPUsage}
	wideRow := []string{id, n.Description, n.Name, n.Projectid, n.Partitionid, nat, shared, prefixes, usage, privateSuper, annotations}
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
		name := truncate(i.Name, 30)
		description := truncate(i.Description, 30)
		allocationUUID := ""
		if i.Allocationuuid != nil {
			allocationUUID = *i.Allocationuuid
		}
		row := []string{ipaddress, description, name, network, project, ipType, strings.Join(shortTags, "\n")}
		wide := []string{ipaddress, allocationUUID, i.Description, i.Name, network, project, ipType, strings.Join(i.Tags, "\n")}
		m.addShortData(row, i)
		m.addWideData(wide, i)
	}
	m.shortHeader = []string{"IP", "Description", "Name", "Network", "Project", "Type", "Tags"}
	m.wideHeader = []string{"IP", "Allocation UUID", "Description", "Name", "Network", "Project", "Type", "Tags"}
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
		power := color.WhiteString(dot)
		powerText := ""
		ipmi := i.Ipmi
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
			power, powerText = extractPowerState(ipmi)
		}
		biosVersion := ""
		bios := i.Bios
		if bios != nil {
			biosVersion = strValue(bios.Version)
		}

		row := []string{id, statusEmoji, power, ipAddress, mac, bpn, biosVersion, bmcVersion, size, partition}
		wide := []string{id, statusEmoji, powerText, ipAddress, mac, bpn, cs, ps, biosVersion, bmcVersion, size, partition}
		m.addShortData(row, m)
		m.addWideData(wide, i)
	}
	m.shortHeader = []string{"ID", "Status", "Power", "IP", "Mac", "Board Part Number", "Bios Version", "BMC Version", "Size", "Partition"}
	m.wideHeader = []string{"ID", "Status", "Power", "IP", "Mac", "Board Part Number", "Chassis Serial", "Product Serial", "Bios Version", "BMC Version", "Size", "Partition"}
	m.render()
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

// Print machine logs
func (m MetalMachineLogsPrinter) Print(data []*models.V1MachineProvisioningEvent) {
	for _, i := range data {
		row := []string{i.Time.String(), strValue(i.Event), i.Message}
		m.addShortData(row, m)
		m.addWideData(row, m)
	}
	m.shortHeader = []string{"Time", "Event", "Message"}
	m.wideHeader = m.shortHeader
	m.table.SetAutoWrapText(false)
	m.render()
}

// Print all MetalIPs in a table
func (m MetalProjectTablePrinter) Print(data []*models.V1ProjectResponse) {
	m.wideHeader = []string{"UID", "Tenant", "Name", "Description", "Quotas Clusters/Machines/IPs", "Labels", "Annotations"}
	m.shortHeader = m.wideHeader
	m.Order(data)
	for _, pr := range data {
		quotas := "∞/∞/∞"
		if pr.Quotas != nil {
			clusterQuota := "∞"
			machineQuota := "∞"
			ipQuota := "∞"
			qs := pr.Quotas
			if qs.Cluster != nil {
				if qs.Cluster.Quota != 0 {
					clusterQuota = strconv.FormatInt(int64(qs.Cluster.Quota), 10)
				}
			}
			if qs.Machine != nil {
				if qs.Machine.Quota != 0 {
					machineQuota = strconv.FormatInt(int64(qs.Machine.Quota), 10)
				}
			}
			if qs.IP != nil {
				if qs.IP.Quota != 0 {
					ipQuota = strconv.FormatInt(int64(qs.IP.Quota), 10)
				}
			}
			quotas = fmt.Sprintf("%s/%s/%s", clusterQuota, machineQuota, ipQuota)
		}
		labels := strings.Join(pr.Meta.Labels, "\n")
		as := []string{}
		for k, v := range pr.Meta.Annotations {
			as = append(as, k+"="+v)
		}
		annotations := strings.Join(as, "\n")

		wide := []string{pr.Meta.ID, pr.TenantID, pr.Name, pr.Description, quotas, labels, annotations}
		m.addShortData(wide, pr)
		m.addWideData(wide, pr)
	}
	m.render()
}

// Print ipmi data from machines
func (m MetalFirmwaresPrinter) Print(data *models.V1FirmwaresResponse) {
	for k, vv := range data.Revisions {
		for v, bb := range vv.VendorRevisions {
			for b, rr := range bb.BoardRevisions {
				sort.Strings(rr)
				for _, rev := range rr {
					row := []string{k, v, b, rev}
					wide := row
					m.addShortData(row, m)
					m.addWideData(wide, m)
				}
			}
		}
	}
	m.shortHeader = []string{"Firmware", "Vendor", "Board", "Revision"}
	m.wideHeader = m.shortHeader
	m.render()
}
func (m FilesystemLayoutPrinter) Print(data []*models.V1FilesystemLayoutResponse) {
	for _, fsl := range data {
		imageConstraints := []string{}
		for os, v := range fsl.Constraints.Images {
			imageConstraints = append(imageConstraints, os+" "+v)
		}

		fsls := fsl.Filesystems
		sort.Slice(fsls, func(i, j int) bool { return depth(fsls[i].Path) < depth(fsls[j].Path) })
		fss := bytes.NewBufferString("")

		w := tabwriter.NewWriter(fss, 0, 0, 0, ' ', 0)
		for _, fs := range fsls {
			fmt.Fprintf(w, "%s\t  \t%s\n", fs.Path, *fs.Device)
		}
		err := w.Flush()
		if err != nil {
			panic(err)
		}

		row := []string{strValue(fsl.ID), fsl.Description, fss.String(), strings.Join(fsl.Constraints.Sizes, "\n"), strings.Join(imageConstraints, "\n")}
		m.addShortData(row, m)
		m.addWideData(row, m)
	}
	m.shortHeader = []string{"ID", "Description", "Filesystems", "Sizes", "Images"}
	m.table.SetAutoWrapText(false)
	m.render()
}
func depth(path string) uint {
	var count uint = 0
	for p := filepath.Clean(path); p != "/"; count++ {
		p = filepath.Dir(p)
	}
	return count
}

// genericObject transforms the input to a struct which has fields with the same name as in the json struct.
// this is handy for template rendering as the output of -o json|yaml can be used as the input for the template
func genericObject(input interface{}) map[string]interface{} {
	b, err := json.Marshal(input)
	if err != nil {
		fmt.Printf("unable to marshall input:%v", err)
		os.Exit(1)
	}
	var result interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		fmt.Printf("unable to unmarshal input:%v", err)
		os.Exit(1)
	}
	return result.(map[string]interface{})
}

// strValue returns the value of a string pointer of not nil, otherwise empty string
func strValue(strPtr *string) string {
	if strPtr != nil {
		return *strPtr
	}
	return ""
}

//nolint:unparam
func truncate(input string, maxlength int) string {
	elipsis := "..."
	il := len(input)
	el := len(elipsis)
	if il <= maxlength {
		return input
	}
	if maxlength <= el {
		return input[:maxlength]
	}
	startlength := ((maxlength - el) / 2) - el/2

	output := input[:startlength] + elipsis
	missing := maxlength - len(output)
	output = output + input[il-missing:]
	return output
}

func truncateEnd(input string, maxlength int) string {
	elipsis := "..."
	length := len(input) + len(elipsis)
	if length <= maxlength {
		return input
	}
	return input[:maxlength] + elipsis
}

func humanizeDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"d", days},
		{"h", hours},
		{"m", minutes},
		{"s", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		default:
			parts = append(parts, fmt.Sprintf("%d%s", chunk.amount, chunk.singularName))
		}
	}

	if len(parts) == 0 {
		return "0s"
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " ")
}
func sortIPs(v1ips []*models.V1IPResponse) []*models.V1IPResponse {

	v1ipmap := make(map[string]*models.V1IPResponse)
	var ips []string
	for _, v1ip := range v1ips {
		v1ipmap[*v1ip.Ipaddress] = v1ip
		ips = append(ips, *v1ip.Ipaddress)
	}

	realIPs := make([]net.IP, 0, len(ips))

	for _, ip := range ips {
		realIPs = append(realIPs, net.ParseIP(ip))
	}

	sort.Slice(realIPs, func(i, j int) bool {
		return bytes.Compare(realIPs[i], realIPs[j]) < 0
	})

	var result []*models.V1IPResponse
	for _, ip := range realIPs {
		result = append(result, v1ipmap[ip.String()])
	}
	return result
}
