package completion

import (
	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
)

func (c *Completion) ImageClassificationCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{
		models.DatastoreImageSearchQueryClassificationDeprecated,
		models.DatastoreImageSearchQueryClassificationPreview,
		models.DatastoreImageSearchQueryClassificationSupported,
	}, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) ImageFeatureCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"machine", "firewall"}, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) ImageListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Image().ListImages(image.NewListImagesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Payload {
		if i.ID == nil {
			continue
		}
		names = append(names, *i.ID)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) ImageNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Image().ListImages(image.NewListImagesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Payload {
		names = append(names, i.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
