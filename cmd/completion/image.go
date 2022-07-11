package completion

import "github.com/spf13/cobra"

func (c *Completion) ImageListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.driver.ImageList()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Image {
		names = append(names, *i.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
