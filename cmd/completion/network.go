package completion

import "github.com/spf13/cobra"

func (c *Completion) NetworkListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.NetworkList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, n := range resp.Networks {
		names = append(names, *n.ID+"\t"+n.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
