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
	return names, cobra.ShellCompDirectiveDefault
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
	return names, cobra.ShellCompDirectiveDefault
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
	return names, cobra.ShellCompDirectiveDefault
}
func filesystemLayoutListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.FilesystemLayoutList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp {
		names = append(names, *s.ID)
	}
	return names, cobra.ShellCompDirectiveDefault
}
func machineListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.MachineList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Machines {
		names = append(names, *m.ID)
	}
	return names, cobra.ShellCompDirectiveDefault
}
func networkListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.NetworkList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, n := range resp.Networks {
		names = append(names, *n.ID)
	}
	return names, cobra.ShellCompDirectiveDefault
}

func ipListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.IPList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.IPs {
		names = append(names, *i.Ipaddress)
	}
	return names, cobra.ShellCompDirectiveDefault
}
func projectListCompletion(driver *metalgo.Driver) ([]string, cobra.ShellCompDirective) {
	resp, err := driver.ProjectList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Project {
		names = append(names, p.Meta.ID)
	}
	return names, cobra.ShellCompDirectiveDefault
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
	return names, cobra.ShellCompDirectiveDefault
}
func outputFormatListCompletion() ([]string, cobra.ShellCompDirective) {
	return []string{"table", "wide", "markdown", "json", "yaml", "template"}, cobra.ShellCompDirectiveDefault
}
func outputOrderListCompletion() ([]string, cobra.ShellCompDirective) {
	return []string{"size", "id", "status", "event", "when", "partition", "project"}, cobra.ShellCompDirectiveDefault
}
