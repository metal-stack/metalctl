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

func (c *Completion) FirewallEgressCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	egressrules := []string{
		"tcp@0.0.0.0/0@443@\"allow outgoing https\"\tdefault outgoing https",
		"tcp@0.0.0.0/0@53@\"allow outgoing dns via tcp\"\tdefault outgoing dns via tcp",
		"udp@0.0.0.0/0@53#123@\"allow outgoing dns and ntp via udp\"\tdefault outgoing dns and ntp via udp",
	}
	return egressrules, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) FirewallIngressCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ingressrules := []string{
		"tcp@0.0.0.0/0@22@\"allow incoming ssh\"\tallow incoming ssh",
	}
	return ingressrules, cobra.ShellCompDirectiveNoFileComp
}
