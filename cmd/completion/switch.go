package completion

import (
	"github.com/metal-stack/metal-go/api/client/switch_operations"
	"github.com/spf13/cobra"
)

func (c *Completion) SwitchListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.SwitchOperations().ListSwitches(switch_operations.NewListSwitchesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp.Payload {
		names = append(names, *s.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
