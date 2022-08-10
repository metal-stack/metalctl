package cmd

import (
	"errors"

	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type imageCmd struct {
	*config
	*genericcli.GenericCLI[*models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, *models.V1ImageResponse]
}

func newImageCmd(c *config) *cobra.Command {
	w := imageCmd{
		config:     c,
		GenericCLI: genericcli.NewGenericCLI[*models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, *models.V1ImageResponse](imageCRUD{config: c}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, *models.V1ImageResponse]{
		gcli:              w.GenericCLI,
		singular:          "image",
		plural:            "images",
		description:       "os images available to be installed on machines.",
		availableSortKeys: sorters.ImageSortKeys(),
		validArgsFunc:     c.comp.ImageListCompletion,
		createRequestFromCLI: func() (*models.V1ImageCreateRequest, error) {
			return &models.V1ImageCreateRequest{
				ID:          pointer.Pointer(viper.GetString("id")),
				Name:        viper.GetString("name"),
				Description: viper.GetString("description"),
				URL:         pointer.Pointer(viper.GetString("url")),
				Features:    viper.GetStringSlice("features"),
			}, nil
		},
	})

	cmds.createCmd.Flags().StringP("id", "", "", "ID of the image. [required]")
	cmds.createCmd.Flags().StringP("url", "", "", "url of the image. [required]")
	cmds.createCmd.Flags().StringP("name", "n", "", "Name of the image. [optional]")
	cmds.createCmd.Flags().StringP("description", "d", "", "Description of the image. [optional]")
	cmds.createCmd.Flags().StringSlice("features", []string{}, "features of the image, can be one of machine|firewall")
	must(cmds.createCmd.MarkFlagRequired("id"))
	must(cmds.createCmd.MarkFlagRequired("url"))

	cmds.listCmd.Flags().Bool("show-usage", false, "show from how many allocated machines every image is used")

	return cmds.buildRootCmd()
}

type imageCRUD struct {
	*config
}

func (c imageCRUD) Get(id string) (*models.V1ImageResponse, error) {
	resp, err := c.client.Image().FindImage(image.NewFindImageParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c imageCRUD) List() ([]*models.V1ImageResponse, error) {
	resp, err := c.client.Image().ListImages(image.NewListImagesParams().WithShowUsage(pointer.Pointer(viper.GetBool("show-usage"))), nil)
	if err != nil {
		return nil, err
	}

	err = sorters.ImageSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c imageCRUD) Delete(id string) (*models.V1ImageResponse, error) {
	resp, err := c.client.Image().DeleteImage(image.NewDeleteImageParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c imageCRUD) Create(rq *models.V1ImageCreateRequest) (*models.V1ImageResponse, error) {
	resp, err := c.client.Image().CreateImage(image.NewCreateImageParams().WithBody(rq), nil)
	if err != nil {
		var r *image.CreateImageConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c imageCRUD) Update(rq *models.V1ImageUpdateRequest) (*models.V1ImageResponse, error) {
	resp, err := c.client.Image().UpdateImage(image.NewUpdateImageParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
