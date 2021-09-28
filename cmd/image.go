package cmd

import (
	"errors"
	"fmt"
	"net/http"

	metalgo "github.com/metal-stack/metal-go"
	imagemodel "github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func newImageCmd(c *config) *cobra.Command {
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
			return c.imageList(args)
		},
	}
	imageDescribeCmd := &cobra.Command{
		Use:   "describe <imageID>",
		Short: "describe a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.imageDescribe(args)
		},
		ValidArgsFunction: c.comp.ImageListCompletion,
	}
	imageCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.imageCreate(args)
		},
		PreRun: bindPFlags,
	}
	imageUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.imageUpdate(args)
		},
		PreRun: bindPFlags,
	}
	imageApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.imageApply(args)
		},
		PreRun: bindPFlags,
	}
	imageDeleteCmd := &cobra.Command{
		Use:     "delete <imageID>",
		Aliases: []string{"rm"},
		Short:   "delete a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.imageDelete(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.ImageListCompletion,
	}
	imageEditCmd := &cobra.Command{
		Use:   "edit <imageID>",
		Short: "edit a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.imageEdit(args)
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

	imageCmd.AddCommand(imageListCmd)
	imageCmd.AddCommand(imageDescribeCmd)
	imageCmd.AddCommand(imageCreateCmd)
	imageCmd.AddCommand(imageUpdateCmd)
	imageCmd.AddCommand(imageDeleteCmd)
	imageCmd.AddCommand(imageApplyCmd)
	imageCmd.AddCommand(imageEditCmd)

	return imageCmd
}

func (c *config) imageList(args []string) error {
	resp, err := c.driver.ImageList()
	if err != nil {
		return err
	}
	return output.New().Print(resp.Image)
}

func (c *config) imageDescribe(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no image ID given")
	}
	imageID := args[0]
	resp, err := c.driver.ImageGet(imageID)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Image)
}

func (c *config) imageCreate(args []string) error {
	var icr metalgo.ImageCreateRequest
	if viper.GetString("file") != "" {
		var iars []metalgo.ImageCreateRequest
		err := readFrom(viper.GetString("file"), &icr, func(data interface{}) {
			doc := data.(*metalgo.ImageCreateRequest)
			iars = append(iars, *doc)
		})
		if err != nil {
			return err
		}
		if len(iars) != 1 {
			return fmt.Errorf("image create error more or less than one image given:%d", len(iars))
		}
		icr = iars[0]
	} else {
		icr = metalgo.ImageCreateRequest{
			Description: viper.GetString("description"),
			ID:          viper.GetString("id"),
			Name:        viper.GetString("name"),
			URL:         viper.GetString("url"),
			Features:    viper.GetStringSlice("features"),
		}
	}
	resp, err := c.driver.ImageCreate(icr)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Image)
}
func (c *config) imageUpdate(args []string) error {
	iar, err := readImageCreateRequests(viper.GetString("file"))
	if err != nil {
		return err
	}
	resp, err := c.driver.ImageUpdate(iar)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Image)
}

func readImageCreateRequests(filename string) (metalgo.ImageCreateRequest, error) {
	var iar metalgo.ImageCreateRequest
	err := readFrom(filename, &iar, func(data interface{}) {
		doc := data.(*metalgo.ImageCreateRequest)
		iar = *doc
	})
	if err != nil {
		return iar, err
	}
	return iar, nil
}

// TODO: General apply method would be useful as these are quite a lot of lines and it's getting erroneous
func (c *config) imageApply(args []string) error {
	var iars []metalgo.ImageCreateRequest
	var iar metalgo.ImageCreateRequest
	err := readFrom(viper.GetString("file"), &iar, func(data interface{}) {
		doc := data.(*metalgo.ImageCreateRequest)
		iars = append(iars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		iar = metalgo.ImageCreateRequest{}
	})
	if err != nil {
		return err
	}
	var response []*models.V1ImageResponse
	for _, iar := range iars {
		image, err := c.driver.ImageGet(iar.ID)
		if err != nil {
			var r *imagemodel.FindImageDefault
			if !errors.As(err, &r) {
				return err
			}
			if r.Code() != http.StatusNotFound {
				return err
			}
		}
		if image.Image == nil {
			resp, err := c.driver.ImageCreate(iar)
			if err != nil {
				return err
			}
			response = append(response, resp.Image)
			continue
		}
		if image.Image.ID != nil {
			resp, err := c.driver.ImageUpdate(iar)
			if err != nil {
				return err
			}
			response = append(response, resp.Image)
			continue
		}
	}
	return output.NewDetailer().Detail(response)
}

func (c *config) imageDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no image ID given")
	}
	imageID := args[0]
	resp, err := c.driver.ImageDelete(imageID)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Image)
}

func (c *config) imageEdit(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no image ID given")
	}
	imageID := args[0]

	getFunc := func(id string) ([]byte, error) {
		resp, err := c.driver.ImageGet(imageID)
		if err != nil {
			return nil, err
		}
		content, err := yaml.Marshal(resp.Image)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		iar, err := readImageCreateRequests(filename)
		if err != nil {
			return err
		}
		fmt.Printf("new image classification:%s\n", *iar.Classification)
		uresp, err := c.driver.ImageUpdate(iar)
		if err != nil {
			return err
		}
		return output.NewDetailer().Detail(uresp.Image)
	}

	return edit(imageID, getFunc, updateFunc)
}
