package cmd

import (
	"errors"
	"fmt"

	sizemodel "github.com/metal-stack/metal-go/api/client/sizeimageconstraint"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeImageConstraintCmd struct {
	*config
}

func newSizeImageConstraintCmd(c *config) *cobra.Command {
	w := sizeImageConstraintCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1SizeImageConstraintCreateRequest, *models.V1SizeImageConstraintUpdateRequest, *models.V1SizeImageConstraintResponse]{
		BinaryName:        binaryName,
		GenericCLI:        genericcli.NewGenericCLI[*models.V1SizeImageConstraintCreateRequest, *models.V1SizeImageConstraintUpdateRequest, *models.V1SizeImageConstraintResponse](w).WithFS(c.fs),
		Singular:          "imageconstraint",
		Plural:            "imageconstraints",
		Description:       "If a size has specific requirements regarding the images which must fullfil certain constraints, this can be configured here.",
		Aliases:           []string{"ic"},
		AvailableSortKeys: sorters.SizeImageConstraintSortKeys(),
		ValidArgsFn:       c.comp.SizeImageConstraintListCompletion,
		DescribePrinter:   func() printers.Printer { return c.describePrinter },
		ListPrinter:       func() printers.Printer { return c.listPrinter },
	}

	tryCmd := &cobra.Command{
		Use:   "try",
		Short: "try if size and image can be allocated",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.try()
		},
	}

	tryCmd.Flags().StringP("size", "", "", "size to check if allocaltion is possible")
	tryCmd.Flags().StringP("image", "", "", "image to check if allocaltion is possible")
	must(tryCmd.MarkFlagRequired("size"))
	must(tryCmd.MarkFlagRequired("image"))

	return genericcli.NewCmds(cmdsConfig, tryCmd)
}

func (c sizeImageConstraintCmd) Get(id string) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.client.Sizeimageconstraint().FindSizeImageConstraint(sizemodel.NewFindSizeImageConstraintParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCmd) List() ([]*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.client.Sizeimageconstraint().ListSizeImageConstraints(sizemodel.NewListSizeImageConstraintsParams(), nil)
	if err != nil {
		return nil, err
	}

	err = sorters.SizeImageConstraintSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCmd) Delete(id string) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.client.Sizeimageconstraint().DeleteSizeImageConstraint(sizemodel.NewDeleteSizeImageConstraintParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCmd) Create(rq *models.V1SizeImageConstraintCreateRequest) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.client.Sizeimageconstraint().CreateSizeImageConstraint(sizemodel.NewCreateSizeImageConstraintParams().WithBody(rq), nil)
	if err != nil {
		var r *sizemodel.CreateSizeImageConstraintConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCmd) Update(rq *models.V1SizeImageConstraintUpdateRequest) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.client.Sizeimageconstraint().UpdateSizeImageConstraint(sizemodel.NewUpdateSizeImageConstraintParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *sizeImageConstraintCmd) try() error {
	_, err := c.client.Sizeimageconstraint().TrySizeImageConstraint(sizemodel.NewTrySizeImageConstraintParams().WithBody(&models.V1SizeImageConstraintTryRequest{
		Size:  pointer.Pointer(viper.GetString("size")),
		Image: pointer.Pointer(viper.GetString("image")),
	}), nil)
	if err != nil {
		return err
	}

	fmt.Fprintln(c.out, "allocation is possible")

	return nil
}
