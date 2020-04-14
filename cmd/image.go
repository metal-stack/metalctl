package cmd

import (
	"fmt"
	"net/http"

	metalgo "github.com/metal-stack/metal-go"
	imagemodel "github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	imageCmd = &cobra.Command{
		Use:   "image",
		Short: "manage images",
		Long:  "os images available to be installed on machines.",
	}

	imageListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all images",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imageList(driver)
		},
	}
	imageDescribeCmd = &cobra.Command{
		Use:   "describe <imageID>",
		Short: "describe a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imageDescribe(driver, args)
		},
	}
	imageCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imageCreate(driver)
		},
		PreRun: bindPFlags,
	}
	imageUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "update a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imageUpdate(driver)
		},
		PreRun: bindPFlags,
	}
	imageApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imageApply(driver)
		},
		PreRun: bindPFlags,
	}
	imageDeleteCmd = &cobra.Command{
		Use:   "delete <imageID>",
		Short: "delete a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imageDelete(driver, args)
		},
		PreRun: bindPFlags,
	}
	imageEditCmd = &cobra.Command{
		Use:   "edit <imageID>",
		Short: "edit a image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return imageEdit(driver, args)
		},
		PreRun: bindPFlags,
	}
)

func init() {
	imageCreateCmd.Flags().StringP("id", "", "", "ID of the image. [required]")
	imageCreateCmd.Flags().StringP("url", "", "", "url of the image. [required]")
	imageCreateCmd.Flags().StringP("name", "n", "", "Name of the image. [optional]")
	imageCreateCmd.Flags().StringP("description", "d", "", "Description of the image. [required]")
	imageCreateCmd.Flags().StringSlice("features", []string{}, "features of the image, can be one of machine|firewall")

	// TODO howto cope with these errors ?
	// err := imageUpdateCmd.MarkFlagRequired("file")
	// if err != nil {
	// 	panic(err)
	// }
	// err = imageApplyCmd.MarkFlagRequired("file")
	// if err != nil {
	// 	panic(err)
	// }

	imageCmd.AddCommand(imageListCmd)
	imageCmd.AddCommand(imageDescribeCmd)
	imageCmd.AddCommand(imageCreateCmd)
	imageCmd.AddCommand(imageUpdateCmd)
	imageCmd.AddCommand(imageDeleteCmd)
	imageCmd.AddCommand(imageApplyCmd)
	imageCmd.AddCommand(imageEditCmd)
}

func imageList(driver *metalgo.Driver) error {
	resp, err := driver.ImageList()
	if err != nil {
		return fmt.Errorf("image list error:%v", err)
	}
	return printer.Print(resp.Image)
}

func imageDescribe(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no image ID given")
	}
	imageID := args[0]
	resp, err := driver.ImageGet(imageID)
	if err != nil {
		return fmt.Errorf("image describe error:%v", err)
	}
	return detailer.Detail(resp.Image)
}

func imageCreate(driver *metalgo.Driver) error {
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
	resp, err := driver.ImageCreate(icr)
	if err != nil {
		return fmt.Errorf("image create error:%v", err)
	}
	return detailer.Detail(resp.Image)
}
func imageUpdate(driver *metalgo.Driver) error {
	iar, err := readImageCreateRequests(viper.GetString("file"))
	if err != nil {
		return err
	}
	resp, err := driver.ImageUpdate(iar)
	if err != nil {
		return fmt.Errorf("image update error:%v", err)
	}
	return detailer.Detail(resp.Image)
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
func imageApply(driver *metalgo.Driver) error {
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
		image, err := driver.ImageGet(iar.ID)
		if err != nil {
			if e, ok := err.(*imagemodel.FindImageDefault); ok {
				if e.Code() != http.StatusNotFound {
					return fmt.Errorf("image get error:%v", err)
				}
			}
		}
		if image.Image == nil {
			resp, err := driver.ImageCreate(iar)
			if err != nil {
				return fmt.Errorf("image update error:%v", err)
			}
			response = append(response, resp.Image)
			continue
		}
		if image.Image.ID != nil {
			resp, err := driver.ImageUpdate(iar)
			if err != nil {
				return fmt.Errorf("image create error:%v", err)
			}
			response = append(response, resp.Image)
			continue
		}
	}
	return detailer.Detail(response)
}

func imageDelete(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no image ID given")
	}
	imageID := args[0]
	resp, err := driver.ImageDelete(imageID)
	if err != nil {
		return fmt.Errorf("image delete error:%v", err)
	}
	return detailer.Detail(resp.Image)
}

func imageEdit(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no image ID given")
	}
	imageID := args[0]

	getFunc := func(id string) ([]byte, error) {
		resp, err := driver.ImageGet(imageID)
		if err != nil {
			return nil, fmt.Errorf("image describe error:%v", err)
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
		uresp, err := driver.ImageUpdate(iar)
		if err != nil {
			return fmt.Errorf("image update error:%v", err)
		}
		return detailer.Detail(uresp.Image)
	}

	return edit(imageID, getFunc, updateFunc)
}
