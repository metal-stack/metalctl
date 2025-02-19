package completion

import (
	"github.com/metal-stack/metal-go/api/client/ip"
	"github.com/metal-stack/metal-go/api/models"
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

func (c *Completion) IPAddressFamilyCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{models.V1IPAllocateRequestAddressfamilyIPV4, models.V1IPAllocateRequestAddressfamilyIPV6}, cobra.ShellCompDirectiveNoFileComp
}
