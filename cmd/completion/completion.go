package completion

import (
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
)

type Completion struct {
	driver *metalgo.Driver
}

func NewCompletion(driver *metalgo.Driver) *Completion {
	return &Completion{
		driver: driver,
	}
}

func (c *Completion) ImageListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.ImageList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Image {
		names = append(names, *i.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) PartitionListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.PartitionList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Partition {
		names = append(names, *p.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) SizeListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.SizeList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp.Size {
		names = append(names, *s.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) FilesystemLayoutListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.FilesystemLayoutList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp {
		names = append(names, *s.ID+"\t"+s.Description)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) MachineListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.MachineList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Machines {
		name := *m.ID
		if m.Allocation != nil && *m.Allocation.Hostname != "" {
			name = name + "\t" + *m.Allocation.Hostname
		}
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) FirewallListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.FirewallList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Firewalls {
		name := *m.ID
		if m.Allocation != nil && *m.Allocation.Hostname != "" {
			name = name + "\t" + *m.Allocation.Hostname
		}
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) NetworkListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.NetworkList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, n := range resp.Networks {
		names = append(names, *n.ID+"\t"+n.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) IpListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.IPList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.IPs {
		names = append(names, *i.Ipaddress+"\t"+i.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) ProjectListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.ProjectList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Project {
		names = append(names, p.Meta.ID+"\t"+p.TenantID+"/"+p.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) ContextListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctxs, err := api.GetContexts()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for name := range ctxs.Contexts {
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func OutputFormatListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"table", "wide", "markdown", "json", "yaml", "template"}, cobra.ShellCompDirectiveNoFileComp
}
func OutputOrderListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"size", "id", "status", "event", "when", "partition", "project"}, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) FirmwareKindCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{string(metalgo.Bmc), string(metalgo.Bios)}, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) FirmwareVendorCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.ListFirmwares("", "", "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var vendors []string
	for _, vv := range resp.Firmwares.Revisions {
		for v := range vv.VendorRevisions {
			vendors = append(vendors, v)
		}
	}
	return vendors, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) FirmwareBoardCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.ListFirmwares("", "", "")
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var boards []string
	for _, vv := range resp.Firmwares.Revisions {
		for _, bb := range vv.VendorRevisions {
			for b := range bb.BoardRevisions {
				boards = append(boards, b)
			}
		}
	}
	return boards, cobra.ShellCompDirectiveNoFileComp
}
func (c *Completion) FirmwareRevisionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return c.firmwareRevisions("", "")
}
func (c *Completion) FirmwareBiosRevisionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return c.firmwareRevisions(args[0], metalgo.Bios)
}
func (c *Completion) FirmwareBmcRevisionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return c.firmwareRevisions(args[0], metalgo.Bmc)
}

func (c *Completion) firmwareRevisions(machineID string, kind metalgo.FirmwareKind) ([]string, cobra.ShellCompDirective) {
	vendor := ""
	board := ""
	if machineID != "" {
		m, err := c.driver.MachineIPMIGet(machineID)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		board = m.Machine.Ipmi.Fru.BoardPartNumber
		vendor = m.Machine.Ipmi.Fru.BoardMfg
	}
	resp, err := c.driver.ListFirmwares(kind, vendor, board)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var revisions []string
	for _, vv := range resp.Firmwares.Revisions {
		for _, bb := range vv.VendorRevisions {
			for _, rr := range bb.BoardRevisions {
				revisions = append(revisions, rr...)
			}
		}
	}
	return revisions, cobra.ShellCompDirectiveNoFileComp
}
