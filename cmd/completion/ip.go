package completion

import (
	"github.com/metal-stack/metal-go/api/client/ip"
	"github.com/spf13/cobra"
)

func (c *Completion) IpListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.IP().ListIPs(ip.NewListIPsParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Payload {
		names = append(names, *i.Ipaddress+"\t"+i.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
