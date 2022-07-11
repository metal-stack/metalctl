package completion

import "github.com/spf13/cobra"

func (c *Completion) PartitionListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.PartitionList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Partition {
		names = append(names, *p.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
