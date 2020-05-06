package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/olekukonko/tablewriter"
)

type (
	// Detailer prints more details than a Printer
	Detailer interface {
		Detail(data interface{}) error
	}
	// TableDetailer produces a human readable model representation
	TableDetailer struct {
		table *tablewriter.Table
	}
	// MetalSwitchTableDetailer is a table printer for a MetalSwitch
	MetalSwitchTableDetailer struct {
		TableDetailer
	}
	// MetalAnyTableDetailer is a Detailer for a any payload
	MetalAnyTableDetailer struct {
		TableDetailer
	}
)

// NewDetailer create a new Detailer which will print more details about metal objects.
func NewDetailer(format string) (Detailer, error) {
	var detailer Detailer
	switch format {
	case "yaml":
		detailer = &YAMLPrinter{}
	case "json":
		detailer = &JSONPrinter{}
	case "table", "wide", "markdown", "template":
		detailer = newTableDetailer("custom")
	default:
		return nil, fmt.Errorf("unknown format:%s", format)
	}
	return detailer, nil
}

func newTableDetailer(format string) TableDetailer {
	table := tablewriter.NewWriter(os.Stdout)
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
	return TableDetailer{
		table: table,
	}
}

// Detail is identical to Print for json
func (j JSONPrinter) Detail(data interface{}) error {
	return j.Print(data)
}

// Detail is identical to Print for json
func (y YAMLPrinter) Detail(data interface{}) error {
	return y.Print(data)
}

// Detail a model in a human readable table
func (t TableDetailer) Detail(data interface{}) error {
	switch d := data.(type) {
	case []*models.V1SwitchResponse:
		MetalSwitchTableDetailer{t}.Detail(d)
	default:
		MetalAnyTableDetailer{t}.Detail(d)
	}
	return nil
}

// Detail MetalSwitch
func (m MetalSwitchTableDetailer) Detail(data []*models.V1SwitchResponse) {
	sort.SliceStable(data, func(i, j int) bool { return *data[i].ID < *data[j].ID })
	d := [][]string{}
	for _, sw := range data {
		filterBySwp := map[string]models.V1BGPFilter{}
		for _, n := range sw.Nics {
			swp := *(n.Name)
			if n.Filter != nil {
				filterBySwp[swp] = *(n.Filter)
			}
		}
		sort.SliceStable(sw.Connections, func(i, j int) bool { return *((*sw.Connections[i]).Nic.Name) < *((*sw.Connections[j]).Nic.Name) })
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
			d = append(d, row)
			for i := 1; i < max; i++ {
				row = append([]string{"", "", "", "", ""}, filterColumns(f, i)...)
				d = append(d, row)
			}
		}
	}
	m.table.SetHeader([]string{"Partition", "Rack", "Switch", "Port", "Machine", "VNI-Filter", "CIDR-Filter"})
	m.table.AppendBulk(d)
	m.table.Render()
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

// Detail MetalIP
func (m MetalAnyTableDetailer) Detail(data interface{}) {
	y := &YAMLPrinter{}
	y.Print(data)
}
