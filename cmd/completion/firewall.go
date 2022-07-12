package completion

import (
	"github.com/metal-stack/metal-go/api/client/firewall"
	"github.com/spf13/cobra"
)

func (c *Completion) FirewallListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Firewall().ListFirewalls(firewall.NewListFirewallsParams(), nil)
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
