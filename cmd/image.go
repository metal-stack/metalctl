package cmd

import (
	"errors"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type imageCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	gcli   *genericcli.GenericCLI[*models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, *models.V1ImageResponse]
}

func newImageCmd(c *config) *cobra.Command {
	w := imageCmd{
		c:      c.client,
		driver: c.driver,
		gcli:   genericcli.NewGenericCLI[*models.V1ImageCreateRequest, *models.V1ImageUpdateRequest, *models.V1ImageResponse](imageGeneric{c: c.client}),
	}

	imageCmd := &cobra.Command{
		Use:   "image",
		Short: "manage images",
		Long:  "os images available to be installed on machines.",
	}

	imageListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all images",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.imageList()
		},
		PreRun: bindPFlags,
	}
	imageDescribeCmd := &cobra.Command{
		Use:   "describe <imageID>",
		Short: "describe a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.DescribeAndPrint(args, genericcli.NewYAMLPrinter())
		},
		ValidArgsFunction: c.comp.ImageListCompletion,
	}
	imageCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.IsSet("file") {
				return w.gcli.CreateFromFileAndPrint(viper.GetString("file"), genericcli.NewYAMLPrinter())
			}

			return w.gcli.CreateAndPrint(&models.V1ImageCreateRequest{
				ID:          pointer.Pointer(viper.GetString("id")),
				Name:        viper.GetString("name"),
				Description: viper.GetString("description"),
				URL:         pointer.Pointer(viper.GetString("url")),
				Features:    viper.GetStringSlice("features"),
			}, genericcli.NewYAMLPrinter())
		},
		PreRun: bindPFlags,
	}
	imageUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.UpdateFromFileAndPrint(viper.GetString("file"), genericcli.NewYAMLPrinter())
		},
		PreRun: bindPFlags,
	}
	imageApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.ApplyFromFileAndPrint(viper.GetString("file"), output.New())
		},
		PreRun: bindPFlags,
	}
	imageDeleteCmd := &cobra.Command{
		Use:     "delete <imageID>",
		Short:   "delete a image",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.DeleteAndPrint(args, genericcli.NewYAMLPrinter())
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.ImageListCompletion,
	}
	imageEditCmd := &cobra.Command{
		Use:   "edit <imageID>",
		Short: "edit a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.EditAndPrint(args, genericcli.NewYAMLPrinter())
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.ImageListCompletion,
	}

	imageCreateCmd.Flags().StringP("id", "", "", "ID of the image. [required]")
	imageCreateCmd.Flags().StringP("url", "", "", "url of the image. [required]")
	imageCreateCmd.Flags().StringP("name", "n", "", "Name of the image. [optional]")
	imageCreateCmd.Flags().StringP("description", "d", "", "Description of the image. [required]")
	imageCreateCmd.Flags().StringSlice("features", []string{}, "features of the image, can be one of machine|firewall")

	imageUpdateCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl image describe ubuntu-19.04 > ubuntu.yaml
# vi ubuntu.yaml
## either via stdin
# cat ubuntu.yaml | metalctl image update -f -
## or via file
# metalctl image update -f ubuntu.yaml`)
	must(imageUpdateCmd.MarkFlagRequired("file"))

	imageApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl image describe ubuntu-19.04 > ubuntu.yaml
# vi ubuntu.yaml
## either via stdin
# cat ubuntu.yaml | metalctl image apply -f -
## or via file
# metalctl image apply -f ubuntu.yaml`)
	must(imageApplyCmd.MarkFlagRequired("file"))

	imageListCmd.Flags().Bool("show-usage", false, "show from how many allocated machines every image is used")

	imageCmd.AddCommand(imageListCmd)
	imageCmd.AddCommand(imageDescribeCmd)
	imageCmd.AddCommand(imageCreateCmd)
	imageCmd.AddCommand(imageUpdateCmd)
	imageCmd.AddCommand(imageDeleteCmd)
	imageCmd.AddCommand(imageApplyCmd)
	imageCmd.AddCommand(imageEditCmd)

	return imageCmd
}

type imageGeneric struct {
	c metalgo.Client
}

func (g imageGeneric) Get(id string) (*models.V1ImageResponse, error) {
	resp, err := g.c.Image().FindImage(image.NewFindImageParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (g imageGeneric) Delete(id string) (*models.V1ImageResponse, error) {
	resp, err := g.c.Image().DeleteImage(image.NewDeleteImageParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (g imageGeneric) Create(rq *models.V1ImageCreateRequest) (*models.V1ImageResponse, error) {
	resp, err := g.c.Image().CreateImage(image.NewCreateImageParams().WithBody(rq), nil)
	if err != nil {
		var r *image.CreateImageConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (g imageGeneric) Update(rq *models.V1ImageUpdateRequest) (*models.V1ImageResponse, error) {
	resp, err := g.c.Image().UpdateImage(image.NewUpdateImageParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *imageCmd) imageList() error {
	var (
		resp *metalgo.ImageListResponse
		err  error
	)
	if viper.GetBool("show-usage") {
		resp, err = c.driver.ImageListWithUsage()
	} else {
		resp, err = c.driver.ImageList()
	}
	if err != nil {
		return err
	}
	return output.New().Print(resp.Image)
}
