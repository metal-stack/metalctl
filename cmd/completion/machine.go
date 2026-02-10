package completion

import (
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/cobra"
)

func (c *Completion) MachineListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Machine().ListMachines(machine.NewListMachinesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Payload {
		name := *m.ID
		if m.Allocation != nil && *m.Allocation.Hostname != "" {
			name = name + "\t" + *m.Allocation.Hostname
		}
		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) MachineManufacturerCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Machine().FindIPMIMachines(machine.NewFindIPMIMachinesParams().WithBody(&models.V1MachineFindRequest{}), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Payload {
		if m == nil || m.Ipmi == nil || m.Ipmi.Fru == nil {
			continue
		}

		names = append(names, m.Ipmi.Fru.ProductManufacturer)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) MachineProductPartNumberCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Machine().FindIPMIMachines(machine.NewFindIPMIMachinesParams().WithBody(&models.V1MachineFindRequest{}), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Payload {
		if m == nil || m.Ipmi == nil || m.Ipmi.Fru == nil {
			continue
		}

		names = append(names, m.Ipmi.Fru.ProductPartNumber)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) MachineProductSerialCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Machine().FindIPMIMachines(machine.NewFindIPMIMachinesParams().WithBody(&models.V1MachineFindRequest{}), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Payload {
		if m == nil || m.Ipmi == nil || m.Ipmi.Fru == nil {
			continue
		}

		names = append(names, m.Ipmi.Fru.ProductSerial)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) MachineBoardPartNumberCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Machine().FindIPMIMachines(machine.NewFindIPMIMachinesParams().WithBody(&models.V1MachineFindRequest{}), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Payload {
		if m == nil || m.Ipmi == nil || m.Ipmi.Fru == nil {
			continue
		}

		names = append(names, m.Ipmi.Fru.BoardPartNumber)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) IssueTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Machine().ListIssues(machine.NewListIssuesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, issue := range resp.Payload {
		if issue.ID == nil {
			continue
		}

		name := *issue.ID
		description := pointer.SafeDeref(issue.Description)
		if description != "" {
			name = name + "\t" + description
		}

		names = append(names, name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) IssueSeverityCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Machine().ListIssues(machine.NewListIssuesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	severities := map[string]bool{}
	for _, issue := range resp.Payload {
		if issue.Severity == nil {
			continue
		}

		severities[*issue.Severity] = true
	}

	var names []string
	for s := range severities {
		names = append(names, s)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) MachineRackListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	mfr := &models.V1MachineFindRequest{}

	if cmd.Flag("partition") != nil {
		if partition := cmd.Flag("partition").Value.String(); partition != "" {
			mfr.PartitionID = partition
		}
	}

	resp, err := c.client.Machine().FindMachines(machine.NewFindMachinesParams().WithBody(mfr), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, m := range resp.Payload {
		names = append(names, m.Rackid)
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}
