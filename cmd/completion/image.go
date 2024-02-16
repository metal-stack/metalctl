package completion

import (
	"errors"
	"strings"

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
	return c.listValidImages()
}

func (c *Completion) listValidImages() ([]string, cobra.ShellCompDirective) {
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

func (c *Completion) ImageOSCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Image().ListImages(image.NewListImagesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Payload {
		if i.ID == nil {
			continue
		}
		os, _, err := osAndVersionFromImage(*i.ID)
		if err == nil {
			names = append(names, os)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) ImageVersionCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Image().ListImages(image.NewListImagesParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, i := range resp.Payload {
		if i.ID == nil {
			continue
		}
		_, version, err := osAndVersionFromImage(*i.ID)
		if err == nil {
			names = append(names, version)
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}

func osAndVersionFromImage(id string) (os string, version string, err error) {
	imageParts := strings.Split(id, "-")
	if len(imageParts) < 2 {
		return "", "", errors.New("image does not contain a version")
	}

	parts := len(imageParts) - 1
	os = strings.Join(imageParts[:parts], "-")
	version = strings.Join(imageParts[parts:], "")

	return
}
