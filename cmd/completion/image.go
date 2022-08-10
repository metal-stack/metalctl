package completion

import (
	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/spf13/cobra"
)

func (c *Completion) ImageListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Image().ListImages(image.NewListImagesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Payload {
		names = append(names, *i.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
