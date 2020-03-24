package cmd

import (
	"fmt"
	"net/http"

	"github.com/dustin/go-humanize"
	metalgo "github.com/metal-stack/metal-go"
	sizemodel "github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	sizeCmd = &cobra.Command{
		Use:   "size",
		Short: "manage sizes",
		Long:  "a size is a distinct hardware equipment in terms of cpu cores, ram and storage of a machine.",
	}

	sizeListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all sizes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sizeList(driver)
		},
		PreRun: bindPFlags,
	}
	sizeDescribeCmd = &cobra.Command{
		Use:   "describe <sizeID>",
		Short: "describe a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sizeDescribe(driver, args)
		},
	}
	sizeTryCmd = &cobra.Command{
		Use:   "try",
		Short: "try a specific hardware spec and give the chosen size back",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sizeTry(driver)
		},
		PreRun: bindPFlags,
	}
	sizeCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "create a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sizeCreate(driver)
		},
		PreRun: bindPFlags,
	}
	sizeUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "update a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sizeUpdate(driver)
		},
		PreRun: bindPFlags,
	}
	sizeApplyCmd = &cobra.Command{
		Use:   "apply",
		Short: "create/update a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sizeApply(driver)
		},
		PreRun: bindPFlags,
	}
	sizeDeleteCmd = &cobra.Command{
		Use:   "delete <sizeID>",
		Short: "delete a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sizeDelete(driver, args)
		},
		PreRun: bindPFlags,
	}
	sizeEditCmd = &cobra.Command{
		Use:   "edit <sizeID>",
		Short: "edit a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sizeEdit(driver, args)
		},
		PreRun: bindPFlags,
	}
)

func init() {

	sizeCreateCmd.Flags().StringP("id", "", "", "ID of the size. [required]")
	sizeCreateCmd.Flags().StringP("name", "n", "", "Name of the size. [optional]")
	sizeCreateCmd.Flags().StringP("description", "d", "", "Description of the size. [required]")
	// FIXME constraints must be given in a slice
	sizeCreateCmd.Flags().Int64P("min", "", 0, "min value of given size constraint type. [required]")
	sizeCreateCmd.Flags().Int64P("max", "", 0, "min value of given size constraint type. [required]")
	sizeCreateCmd.Flags().StringP("type", "", "", "type of constraints. [required]")

	sizeUpdateCmd.MarkFlagRequired("file")
	sizeApplyCmd.MarkFlagRequired("file")

	sizeTryCmd.Flags().Int32P("cores", "C", 1, "Cores of the hardware to try")
	sizeTryCmd.Flags().StringP("memory", "M", "", "Memory of the hardware to try, can be given in bytes or any human readable size spec")
	sizeTryCmd.Flags().StringP("storagesize", "S", "", "Total storagesize of the hardware to try, can be given in bytes or any human readable size spec")

	sizeCmd.AddCommand(sizeListCmd)
	sizeCmd.AddCommand(sizeDescribeCmd)
	sizeCmd.AddCommand(sizeTryCmd)
	sizeCmd.AddCommand(sizeCreateCmd)
	sizeCmd.AddCommand(sizeUpdateCmd)
	sizeCmd.AddCommand(sizeDeleteCmd)
	sizeCmd.AddCommand(sizeApplyCmd)
	sizeCmd.AddCommand(sizeEditCmd)
}

func sizeList(driver *metalgo.Driver) error {
	resp, err := driver.SizeList()
	if err != nil {
		return fmt.Errorf("size list error:%v", err)
	}
	return printer.Print(resp.Size)
}

func sizeDescribe(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no size ID given")
	}
	sizeID := args[0]
	resp, err := driver.SizeGet(sizeID)
	if err != nil {
		return fmt.Errorf("size describe error:%v", err)
	}
	return detailer.Detail(resp.Size)
}

func sizeTry(driver *metalgo.Driver) error {

	cores := viper.GetInt32("cores")
	memory, err := humanize.ParseBytes(viper.GetString("memory"))
	if err != nil {
		return err
	}
	storagesize, err := humanize.ParseBytes(viper.GetString("storagesize"))
	if err != nil {
		return err
	}

	resp, _ := driver.SizeTry(cores, memory, storagesize)
	return printer.Print(resp.Logs)
}

func sizeCreate(driver *metalgo.Driver) error {
	var icrs []metalgo.SizeCreateRequest
	var icr metalgo.SizeCreateRequest
	if viper.GetString("file") != "" {
		err := readFrom(viper.GetString("file"), &icr, func(data interface{}) {
			doc := data.(*metalgo.SizeCreateRequest)
			icrs = append(icrs, *doc)
		})
		if err != nil {
			return err
		}
		if len(icrs) != 1 {
			return fmt.Errorf("size create error more or less than one size given:%d", len(icrs))
		}
		icr = icrs[0]
	} else {
		max := viper.GetInt64("min")
		min := viper.GetInt64("max")
		t := viper.GetString("type")
		icr = metalgo.SizeCreateRequest{
			Description: viper.GetString("description"),
			ID:          viper.GetString("id"),
			Name:        viper.GetString("name"),
			Constraints: []*models.V1SizeConstraint{
				{
					Max:  &max,
					Min:  &min,
					Type: &t,
				},
			},
		}
	}

	resp, err := driver.SizeCreate(icr)
	if err != nil {
		return fmt.Errorf("size create error:%v", err)
	}
	return detailer.Detail(resp.Size)
}

func sizeUpdate(driver *metalgo.Driver) error {
	icrs, err := readSizeCreateRequests(viper.GetString("file"))
	if err != nil {
		return err
	}
	if len(icrs) != 1 {
		return fmt.Errorf("size update error more or less than one size given:%d", len(icrs))
	}
	resp, err := driver.SizeUpdate(icrs[0])
	if err != nil {
		return fmt.Errorf("size update error:%v", err)
	}
	return detailer.Detail(resp.Size)
}

func readSizeCreateRequests(filename string) ([]metalgo.SizeCreateRequest, error) {
	var icrs []metalgo.SizeCreateRequest
	var uir metalgo.SizeCreateRequest
	err := readFrom(filename, &uir, func(data interface{}) {
		doc := data.(*metalgo.SizeCreateRequest)
		icrs = append(icrs, *doc)
	})
	if err != nil {
		return nil, err
	}
	if len(icrs) != 1 {
		return nil, fmt.Errorf("size update error more or less than one size given:%d", len(icrs))
	}
	return icrs, nil
}

// TODO: General apply method would be useful as these are quite a lot of lines and it's getting erroneous
func sizeApply(driver *metalgo.Driver) error {
	var iars []metalgo.SizeCreateRequest
	var iar metalgo.SizeCreateRequest
	err := readFrom(viper.GetString("file"), &iar, func(data interface{}) {
		doc := data.(*metalgo.SizeCreateRequest)
		iars = append(iars, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		iar = metalgo.SizeCreateRequest{}
	})
	if err != nil {
		return err
	}
	var response []*models.V1SizeResponse
	for _, iar := range iars {
		p, err := driver.SizeGet(iar.ID)
		if err != nil {
			if e, ok := err.(*sizemodel.FindSizeDefault); ok {
				if e.Code() != http.StatusNotFound {
					return fmt.Errorf("size get error:%v", err)
				}
			}
		}
		if p.Size == nil {
			resp, err := driver.SizeCreate(iar)
			if err != nil {
				return fmt.Errorf("size update error:%v", err)
			}
			response = append(response, resp.Size)
			continue
		}
		if p.Size.ID != nil {
			resp, err := driver.SizeUpdate(iar)
			if err != nil {
				return fmt.Errorf("size create error:%v", err)
			}
			response = append(response, resp.Size)
			continue
		}
	}
	return detailer.Detail(response)
}

func sizeDelete(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no size ID given")
	}
	sizeID := args[0]
	resp, err := driver.SizeDelete(sizeID)
	if err != nil {
		return fmt.Errorf("size delete error:%v", err)
	}
	return detailer.Detail(resp.Size)
}

func sizeEdit(driver *metalgo.Driver, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no size ID given")
	}
	sizeID := args[0]

	getFunc := func(id string) ([]byte, error) {
		resp, err := driver.SizeGet(sizeID)
		if err != nil {
			return nil, fmt.Errorf("size describe error:%v", err)
		}
		content, err := yaml.Marshal(resp.Size)
		if err != nil {
			return nil, err
		}
		return content, nil
	}
	updateFunc := func(filename string) error {
		iars, err := readSizeCreateRequests(filename)
		if err != nil {
			return err
		}
		if len(iars) != 1 {
			return fmt.Errorf("size update error more or less than one size given:%d", len(iars))
		}
		uresp, err := driver.SizeUpdate(iars[0])
		if err != nil {
			return fmt.Errorf("size update error:%v", err)
		}
		return detailer.Detail(uresp.Size)
	}

	return edit(sizeID, getFunc, updateFunc)
}
