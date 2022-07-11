package completion

import "github.com/spf13/cobra"

func (c *Completion) FilesystemLayoutListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.FilesystemLayoutList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp {
		names = append(names, *s.ID+"\t"+s.Description)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
