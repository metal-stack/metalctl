package cmd

import (
	"errors"
	"fmt"

	sizemodel "github.com/metal-stack/metal-go/api/client/sizeimageconstraint"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeImageConstraintCmd struct {
	*config
	*genericcli.GenericCLI[*models.V1SizeImageConstraintCreateRequest, *models.V1SizeImageConstraintUpdateRequest, *models.V1SizeImageConstraintResponse]
}

func newSizeImageConstraintCmd(c *config) *cobra.Command {
	w := sizeImageConstraintCmd{
		config:     c,
		GenericCLI: genericcli.NewGenericCLI[*models.V1SizeImageConstraintCreateRequest, *models.V1SizeImageConstraintUpdateRequest, *models.V1SizeImageConstraintResponse](sizeImageConstraintCRUD{config: c}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1SizeImageConstraintCreateRequest, *models.V1SizeImageConstraintUpdateRequest, *models.V1SizeImageConstraintResponse]{
		gcli:              w.GenericCLI,
		singular:          "imageconstraint",
		plural:            "imageconstraints",
		description:       "If a size has specific requirements regarding the images which must fullfil certain constraints, this can be configured here.",
		aliases:           []string{"ic"},
		availableSortKeys: sorters.SizeImageConstraintSortKeys(),
		validArgsFunc:     c.comp.SizeImageConstraintListCompletion,
	})

	tryCmd := &cobra.Command{
		Use:   "try",
		Short: "try if size and image can be allocated",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.try()
		},
		PreRun: bindPFlags,
	}

	tryCmd.Flags().StringP("size", "", "", "size to check if allocaltion is possible")
	tryCmd.Flags().StringP("image", "", "", "image to check if allocaltion is possible")
	must(tryCmd.MarkFlagRequired("size"))
	must(tryCmd.MarkFlagRequired("image"))

	return cmds.buildRootCmd(tryCmd)
}

type sizeImageConstraintCRUD struct {
	*config
}

func (c sizeImageConstraintCRUD) Get(id string) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.client.Sizeimageconstraint().FindSizeImageConstraint(sizemodel.NewFindSizeImageConstraintParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCRUD) List() ([]*models.V1SizeImageConstraintResponse, error) {
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

func (c sizeImageConstraintCRUD) Delete(id string) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.client.Sizeimageconstraint().DeleteSizeImageConstraint(sizemodel.NewDeleteSizeImageConstraintParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCRUD) Create(rq *models.V1SizeImageConstraintCreateRequest) (*models.V1SizeImageConstraintResponse, error) {
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

func (c sizeImageConstraintCRUD) Update(rq *models.V1SizeImageConstraintUpdateRequest) (*models.V1SizeImageConstraintResponse, error) {
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

	fmt.Println("allocation is possible")

	return nil
}
