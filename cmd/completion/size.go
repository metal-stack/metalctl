package completion

import (
	"github.com/metal-stack/metal-go/api/client/size"
	"github.com/spf13/cobra"
)

func (c *Completion) SizeListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Size().ListSizes(size.NewListSizesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp.Payload {
		names = append(names, *s.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) SizeReservationsListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Size().ListSizeReservations(size.NewListSizeReservationsParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp.Payload {
		names = append(names, *s.ID+"\t"+s.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
