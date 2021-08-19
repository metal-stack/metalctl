package cmd

import (
	metalgo "github.com/metal-stack/metal-go"
	"github.com/spf13/cobra"
)

func imageListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.ImageList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Image {
		names = append(names, *i.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func partitionListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.PartitionList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Partition {
		names = append(names, *p.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func sizeListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.SizeList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp.Size {
		names = append(names, *s.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func filesystemLayoutListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.FilesystemLayoutList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp {
		names = append(names, *s.ID+"\t"+s.Description)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func machineListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.MachineList()
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
func firewallListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.FirewallList()
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
func networkListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.NetworkList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, n := range resp.Networks {
		names = append(names, *n.ID+"\t"+n.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func ipListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.IPList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.IPs {
		names = append(names, *i.Ipaddress+"\t"+i.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func projectListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.ProjectList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Project {
		names = append(names, p.Meta.ID+"\t"+p.TenantID+"/"+p.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func contextListCompletion() ([]string, cobra.ShellCompDirective) {
	ctxs, err := getContexts()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for name := range ctxs.Contexts {
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
func outputFormatListCompletion() ([]string, cobra.ShellCompDirective) {
	return []string{"table", "wide", "markdown", "json", "yaml", "template"}, cobra.ShellCompDirectiveNoFileComp
}
func outputOrderListCompletion() ([]string, cobra.ShellCompDirective) {
	return []string{"size", "id", "status", "event", "when", "partition", "project"}, cobra.ShellCompDirectiveNoFileComp
}

var machineListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return machineListCompletion(driver)
}
var firewallListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return firewallListCompletion(driver)
}
var filesystemLayoutListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return filesystemLayoutListCompletion(driver)
}
var imageListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return imageListCompletion(driver)
}
var networkListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return networkListCompletion(driver)
}
var ipListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return ipListCompletion(driver)
}
var partitionListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return partitionListCompletion(driver)
}
var projectListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return projectListCompletion(driver)
}
var sizeListCompletionFunc = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return sizeListCompletion(driver)
}
