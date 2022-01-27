package cmd

import (
	"errors"
	"fmt"
	"net/http"

	sizemodel "github.com/metal-stack/metal-go/api/client/sizeimageconstraint"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newSizeImageConstraintCmd(c *config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "imageconstraint",
		Aliases: []string{"ic"},
		Short:   "manage size to image constraints",
		Long:    "If a size has specific requirements regarding the images which must fullfill certain constraints, this can be configured here.",
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all size image constraints",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeImageConstraintList()
		},
		PreRun: bindPFlags,
	}
	tryCmd := &cobra.Command{
		Use:   "try",
		Short: "try if size and image can be allocated",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeImageConstraintTry()
		},
		PreRun: bindPFlags,
	}
	describeCmd := &cobra.Command{
		Use:   "describe <sizeID>",
		Short: "describe a size image constraints",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeImageConstraintDescribe(args)
		},
		ValidArgsFunction: c.comp.SizeImageConstraintListCompletion,
	}
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a size image constraints",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeImageConstraintApply()
		},
		PreRun: bindPFlags,
	}
	deleteCmd := &cobra.Command{
		Use:   "delete <sizeID>",
		Short: "delete a size image constraints",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeImageConstraintDelete(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.SizeImageConstraintListCompletion,
	}

	applyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example:

# metalctl sizeimageconstraint describe c1-xlarge-x86 > c1-xlarge-x86.yaml
# vi c1-xlarge-x86.yaml
## either via stdin
# cat c1-xlarge-x86.yaml | metalctl size apply -f -
## or via file
# metalctl sizeimageconstraint apply -f c1-xlarge-x86.yaml`)
	must(applyCmd.MarkFlagRequired("file"))

	tryCmd.Flags().StringP("size", "", "", "size to check if allocaltion is possible")
	tryCmd.Flags().StringP("image", "", "", "image to check if allocaltion is possible")
	must(tryCmd.MarkFlagRequired("size"))
	must(tryCmd.MarkFlagRequired("image"))

	cmd.AddCommand(listCmd)
	cmd.AddCommand(describeCmd)
	cmd.AddCommand(deleteCmd)
	cmd.AddCommand(applyCmd)
	cmd.AddCommand(tryCmd)

	return cmd
}

func (c *config) sizeImageConstraintList() error {
	param := sizemodel.NewListSizeImageConstraintsParams()
	resp, err := c.driver.SizeImageConstraint.ListSizeImageConstraints(param, nil)
	if err != nil {
		return err
	}
	return output.New().Print(resp.Payload)
}

func (c *config) sizeImageConstraintTry() error {
	size := viper.GetString("size")
	image := viper.GetString("image")

	err := c.driver.TrySizeImageConstraint(size, image)
	if err != nil {
		return err
	}
	fmt.Println("allocation is possible")
	return nil
}
func (c *config) sizeImageConstraintDescribe(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no size ID given")
	}
	id := args[0]
	param := sizemodel.NewFindSizeImageConstraintParams()
	param.SetID(id)
	resp, err := c.driver.SizeImageConstraint.FindSizeImageConstraint(param, nil)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Payload)
}

// TODO: General apply method would be useful as these are quite a lot of lines and it's getting erroneous
func (c *config) sizeImageConstraintApply() error {
	var sics []models.V1SizeImageConstraintCreateRequest
	var sic models.V1SizeImageConstraintCreateRequest
	err := readFrom(viper.GetString("file"), &sic, func(data interface{}) {
		doc := data.(*models.V1SizeImageConstraintCreateRequest)
		sics = append(sics, *doc)
		// the request needs to be renewed as otherwise the pointers in the request struct will
		// always point to same last value in the multi-document loop
		sic = models.V1SizeImageConstraintCreateRequest{}
	})
	if err != nil {
		return err
	}
	var response []*models.V1SizeImageConstraintResponse
	for _, sic := range sics {
		sic := sic
		param := sizemodel.NewFindSizeImageConstraintParams()
		param.SetID(*sic.ID)
		p, err := c.driver.SizeImageConstraint.FindSizeImageConstraint(param, nil)
		if err != nil {
			fmt.Printf("Error:%#v", err)
			var r *sizemodel.FindSizeImageConstraintDefault
			if !errors.As(err, &r) {
				return err
			}
			if r.Payload.StatusCode != http.StatusNotFound {
				return err
			}
		}
		if p == nil {
			param := sizemodel.NewCreateSizeImageConstraintParams()
			param.SetBody(&sic)
			resp, err := c.driver.SizeImageConstraint.CreateSizeImageConstraint(param, nil)
			if err != nil {
				return err
			}
			response = append(response, resp.Payload)
			continue
		}

		sicur := &models.V1SizeImageConstraintUpdateRequest{
			ID:          sic.ID,
			Description: sic.Description,
			Name:        sic.Name,
			Constraints: sic.Constraints,
		}
		uparam := sizemodel.NewUpdateSizeImageConstraintParams()
		uparam.SetBody(sicur)
		resp, err := c.driver.SizeImageConstraint.UpdateSizeImageConstraint(uparam, nil)
		if err != nil {
			return err
		}
		response = append(response, resp.Payload)
	}
	return output.NewDetailer().Detail(response)
}

func (c *config) sizeImageConstraintDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no size ID given")
	}
	id := args[0]
	param := sizemodel.NewDeleteSizeImageConstraintParams()
	param.ID = id
	resp, err := c.driver.SizeImageConstraint.DeleteSizeImageConstraint(param, nil)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Payload)
}
