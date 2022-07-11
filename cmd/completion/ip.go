package completion

import "github.com/spf13/cobra"

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
