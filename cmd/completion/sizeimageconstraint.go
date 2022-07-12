package completion

import (
	sizemodel "github.com/metal-stack/metal-go/api/client/sizeimageconstraint"
	"github.com/spf13/cobra"
)

func (c *Completion) SizeImageConstraintListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Sizeimageconstraint().ListSizeImageConstraints(sizemodel.NewListSizeImageConstraintsParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, s := range resp.Payload {
		names = append(names, *s.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
