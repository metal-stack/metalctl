package completion

import (
	"github.com/metal-stack/metal-go/api/client/network"
	"github.com/spf13/cobra"
)

func (c *Completion) NetworkListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Network().ListNetworks(network.NewListNetworksParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, n := range resp.Payload {
		names = append(names, *n.ID+"\t"+n.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) NetworkDestinationPrefixesCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Network().ListNetworks(network.NewListNetworksParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var prefixes []string
	for _, n := range resp.Payload {
		for _, prefix := range n.Destinationprefixes {
			prefixes = append(prefixes, prefix)
		}
	}
	return prefixes, cobra.ShellCompDirectiveNoFileComp
}
