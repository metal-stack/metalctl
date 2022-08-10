package cmd

import (
	"errors"

	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/defaultscmds"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type imageCmd struct {
	*config
}

func newImageCmd(c *config) *cobra.Command {
	w := imageCmd{
		config: c,
	}

	cmds := defaultscmds.New(&defaultscmds.Config[*models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, *models.V1ImageResponse]{
		GenericCLI:        genericcli.NewGenericCLI[*models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, *models.V1ImageResponse](w),
		Singular:          "image",
		Plural:            "images",
		Description:       "os images available to be installed on machines.",
		AvailableSortKeys: sorters.ImageSortKeys(),
		ValidArgsFunc:     c.comp.ImageListCompletion,
		CreateRequestFromCLI: func() (*models.V1ImageCreateRequest, error) {
			return &models.V1ImageCreateRequest{
				ID:          pointer.Pointer(viper.GetString("id")),
				Name:        viper.GetString("name"),
				Description: viper.GetString("description"),
				URL:         pointer.Pointer(viper.GetString("url")),
				Features:    viper.GetStringSlice("features"),
			}, nil
		},
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "", "", "ID of the image. [required]")
			cmd.Flags().StringP("url", "", "", "url of the image. [required]")
			cmd.Flags().StringP("name", "n", "", "Name of the image. [optional]")
			cmd.Flags().StringP("description", "d", "", "Description of the image. [optional]")
			cmd.Flags().StringSlice("features", []string{}, "features of the image, can be one of machine|firewall")
			must(cmd.MarkFlagRequired("id"))
			must(cmd.MarkFlagRequired("url"))
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().Bool("show-usage", false, "show from how many allocated machines every image is used")
		},
	})

	return cmds.Build()
}

func (c imageCmd) Get(id string) (*models.V1ImageResponse, error) {
	resp, err := c.client.Image().FindImage(image.NewFindImageParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c imageCmd) List() ([]*models.V1ImageResponse, error) {
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

func (c imageCmd) Delete(id string) (*models.V1ImageResponse, error) {
	resp, err := c.client.Image().DeleteImage(image.NewDeleteImageParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c imageCmd) Create(rq *models.V1ImageCreateRequest) (*models.V1ImageResponse, error) {
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

func (c imageCmd) Update(rq *models.V1ImageUpdateRequest) (*models.V1ImageResponse, error) {
	resp, err := c.client.Image().UpdateImage(image.NewUpdateImageParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
