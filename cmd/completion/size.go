package completion

import "github.com/spf13/cobra"

func (c *Completion) SizeListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.SizeList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp.Size {
		names = append(names, *s.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
