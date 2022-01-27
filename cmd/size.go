package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dustin/go-humanize"
	metalgo "github.com/metal-stack/metal-go"
	sizemodel "github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func newSizeCmd(c *config) *cobra.Command {
	sizeCmd := &cobra.Command{
		Use:   "size",
		Short: "manage sizes",
		Long:  "a size is a distinct hardware equipment in terms of cpu cores, ram and storage of a machine.",
	}

	sizeListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all sizes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeList()
		},
		PreRun: bindPFlags,
	}
	sizeDescribeCmd := &cobra.Command{
		Use:   "describe <sizeID>",
		Short: "describe a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeDescribe(args)
		},
		ValidArgsFunction: c.comp.SizeListCompletion,
	}
	sizeTryCmd := &cobra.Command{
		Use:   "try",
		Short: "try a specific hardware spec and give the chosen size back",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeTry()
		},
		PreRun: bindPFlags,
	}
	sizeCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeCreate()
		},
		PreRun: bindPFlags,
	}
	sizeUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeUpdate()
		},
		PreRun: bindPFlags,
	}
	sizeApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeApply()
		},
		PreRun: bindPFlags,
	}
	sizeDeleteCmd := &cobra.Command{
		Use:   "delete <sizeID>",
		Short: "delete a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeDelete(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.SizeListCompletion,
	}
	sizeEditCmd := &cobra.Command{
		Use:   "edit <sizeID>",
		Short: "edit a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeEdit(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.SizeListCompletion,
	}

	sizeCreateCmd.Flags().StringP("id", "", "", "ID of the size. [required]")
	sizeCreateCmd.Flags().StringP("name", "n", "", "Name of the size. [optional]")
	sizeCreateCmd.Flags().StringP("description", "d", "", "Description of the size. [required]")
	// FIXME constraints must be given in a slice
	sizeCreateCmd.Flags().Int64P("min", "", 0, "min value of given size constraint type. [required]")
	sizeCreateCmd.Flags().Int64P("max", "", 0, "min value of given size constraint type. [required]")
	sizeCreateCmd.Flags().StringP("type", "", "", "type of constraints. [required]")

	sizeApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl size describe c1-xlarge-x86 > c1-xlarge-x86.yaml
# vi c1-xlarge-x86.yaml
## either via stdin
# cat c1-xlarge-x86.yaml | metalctl size apply -f -
## or via file
# metalctl size apply -f c1-xlarge-x86.yaml`)
	must(sizeApplyCmd.MarkFlagRequired("file"))

	sizeUpdateCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl size describe c1-xlarge-x86 > c1-xlarge-x86.yaml
# vi c1-xlarge-x86.yaml
## either via stdin
# cat c1-xlarge-x86.yaml | metalctl size update -f -
## or via file
# metalctl size update -f c1-xlarge-x86.yaml`)
	must(sizeUpdateCmd.MarkFlagRequired("file"))

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
	sizeCmd.AddCommand(newSizeImageConstraintCmd(c))

	return sizeCmd
}

func (c *config) sizeList() error {
	resp, err := c.driver.SizeList()
	if err != nil {
		return err
	}
	return output.New().Print(resp.Size)
}

func (c *config) sizeDescribe(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no size ID given")
	}
	sizeID := args[0]
	resp, err := c.driver.SizeGet(sizeID)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Size)
}

func (c *config) sizeTry() error {

	cores := viper.GetInt32("cores")
	memory, err := humanize.ParseBytes(viper.GetString("memory"))
	if err != nil {
		return err
	}
	storagesize, err := humanize.ParseBytes(viper.GetString("storagesize"))
	if err != nil {
		return err
	}

	resp, _ := c.driver.SizeTry(cores, memory, storagesize)
	return output.New().Print(resp.Logs)
}

func (c *config) sizeCreate() error {
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

	resp, err := c.driver.SizeCreate(icr)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Size)
}

func (c *config) sizeUpdate() error {
	icrs, err := readSizeCreateRequests(viper.GetString("file"))
	if err != nil {
		return err
	}
	if len(icrs) != 1 {
		return fmt.Errorf("size update error more or less than one size given:%d", len(icrs))
	}
	resp, err := c.driver.SizeUpdate(icrs[0])
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Size)
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
func (c *config) sizeApply() error {
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
		p, err := c.driver.SizeGet(iar.ID)
		if err != nil {
			var r *sizemodel.FindSizeDefault
			if !errors.As(err, &r) {
				return err
			}
			if r.Code() != http.StatusNotFound {
				return err
			}
		}
		if p.Size == nil {
			resp, err := c.driver.SizeCreate(iar)
			if err != nil {
				return err
			}
			response = append(response, resp.Size)
			continue
		}

		resp, err := c.driver.SizeUpdate(iar)
		if err != nil {
			return err
		}
		response = append(response, resp.Size)
	}
	return output.NewDetailer().Detail(response)
}

func (c *config) sizeDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no size ID given")
	}
	sizeID := args[0]
	resp, err := c.driver.SizeDelete(sizeID)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Size)
}

func (c *config) sizeEdit(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no size ID given")
	}
	sizeID := args[0]

	getFunc := func(id string) ([]byte, error) {
		resp, err := c.driver.SizeGet(sizeID)
		if err != nil {
			return nil, err
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
		uresp, err := c.driver.SizeUpdate(iars[0])
		if err != nil {
			return err
		}
		return output.NewDetailer().Detail(uresp.Size)
	}

	return edit(sizeID, getFunc, updateFunc)
}
