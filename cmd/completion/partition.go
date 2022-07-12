package completion

import (
	"github.com/metal-stack/metal-go/api/client/partition"
	"github.com/spf13/cobra"
)

func (c *Completion) PartitionListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Partition().ListPartitions(partition.NewListPartitionsParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Payload {
		names = append(names, *p.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
