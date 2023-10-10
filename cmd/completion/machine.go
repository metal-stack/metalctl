package completion

import (
	"github.com/metal-stack/metal-go/api/client/machine"
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

func (c *Completion) IssueTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Machine().ListIssues(machine.NewListIssuesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, issue := range resp.Payload {
		issue := issue

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
		issue := issue

		if issue.ID == nil {
			continue
		}
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
