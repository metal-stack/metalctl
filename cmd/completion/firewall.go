package completion

import "github.com/spf13/cobra"

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
