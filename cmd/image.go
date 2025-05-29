package cmd

import (
	"errors"
	"fmt"

	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
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

	cmdsConfig := &genericcli.CmdsConfig[*models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, *models.V1ImageResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI(w).WithFS(c.fs),
		Singular:        "image",
		Plural:          "images",
		Description:     "os images available to be installed on machines.",
		Sorter:          sorters.ImageSorter(),
		ValidArgsFn:     c.comp.ImageListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
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
			cmd.Flags().StringP("id", "", "", "ID of the image.")
			cmd.Flags().StringP("url", "", "", "url of the image.")
			cmd.Flags().StringP("name", "n", "", "Name of the image.")
			cmd.Flags().StringP("description", "d", "", "Description of the image.")
			cmd.Flags().StringSlice("features", []string{}, "features of the image, can be one of machine|firewall")

			cmd.MarkFlagsMutuallyExclusive("file", "id")
			cmd.MarkFlagsRequiredTogether("id", "url")
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().Bool("show-usage", false, "show from how many allocated machines every image is used")

			cmd.Flags().String("classification", "", "Classification of this image.")
			cmd.Flags().String("features", "", "Features of this image.")
			cmd.Flags().String("id", "", "ID of the image.")
			cmd.Flags().String("name", "", "Name of the image.")
			cmd.Flags().String("os", "", "OS derivate of this image.")
			cmd.Flags().String("version", "", "Version of this image.")

			genericcli.Must(cmd.RegisterFlagCompletionFunc("id", c.comp.ImageListCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("name", c.comp.ImageNameCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("classification", c.comp.ImageClassificationCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("features", c.comp.ImageFeatureCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("os", c.comp.ImageOSCompletion))
			genericcli.Must(cmd.RegisterFlagCompletionFunc("version", c.comp.ImageVersionCompletion))
		},
	}

	return genericcli.NewCmds(cmdsConfig)
}

func (c imageCmd) Get(id string) (*models.V1ImageResponse, error) {
	resp, err := c.client.Image().FindImage(image.NewFindImageParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c imageCmd) List() ([]*models.V1ImageResponse, error) {
	resp, err := c.client.Image().FindImages(image.NewFindImagesParams().WithBody(&models.V1ImageFindRequest{
		Classification: viper.GetString("classification"),
		Features:       viper.GetStringSlice("features"),
		ID:             viper.GetString("id"),
		Name:           viper.GetString("name"),
		Os:             viper.GetString("os"),
		Version:        viper.GetString("version"),
	}).WithShowUsage(pointer.Pointer(viper.GetBool("show-usage"))), nil)
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

func (c imageCmd) Convert(r *models.V1ImageResponse) (string, *models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.ID, imageResponseToCreate(r), imageResponseToUpdate(r), nil
}

func imageResponseToCreate(r *models.V1ImageResponse) *models.V1ImageCreateRequest {
	return &models.V1ImageCreateRequest{
		Classification: r.Classification,
		Description:    r.Description,
		ExpirationDate: pointer.SafeDeref(r.ExpirationDate),
		Features:       r.Features,
		ID:             r.ID,
		Name:           r.Name,
		URL:            &r.URL,
	}
}

func imageResponseToUpdate(r *models.V1ImageResponse) *models.V1ImageUpdateRequest {
	return &models.V1ImageUpdateRequest{
		Classification: r.Classification,
		Description:    r.Description,
		ExpirationDate: r.ExpirationDate,
		Features:       r.Features,
		ID:             r.ID,
		Name:           r.Name,
		URL:            r.URL,
		Usedby:         r.Usedby, // TODO this field should not be in here
	}
}
