package cmd

import (
	"errors"
	"fmt"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/size"
	sizemodel "github.com/metal-stack/metal-go/api/client/sizeimageconstraint"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeImageConstraintCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	*genericcli.GenericCLI[*models.V1SizeImageConstraintCreateRequest, *models.V1SizeImageConstraintUpdateRequest, *models.V1SizeImageConstraintResponse]
}

func newSizeImageConstraintCmd(c *config) *cobra.Command {
	w := sizeImageConstraintCmd{
		c:          c.client,
		driver:     c.driver,
		GenericCLI: genericcli.NewGenericCLI[*models.V1SizeImageConstraintCreateRequest, *models.V1SizeImageConstraintUpdateRequest, *models.V1SizeImageConstraintResponse](sizeImageConstraintCRUD{Client: c.client}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1SizeImageConstraintCreateRequest, *models.V1SizeImageConstraintUpdateRequest, *models.V1SizeImageConstraintResponse]{
		gcli:          w.GenericCLI,
		singular:      "imageconstraint",
		plural:        "imageconstraints",
		description:   "If a size has specific requirements regarding the images which must fullfil certain constraints, this can be configured here.",
		aliases:       []string{"ic"},
		validArgsFunc: c.comp.SizeImageConstraintListCompletion,
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

	root := cmds.RootCmd()

	root.AddCommand(tryCmd)

	return root
}

type sizeImageConstraintCRUD struct {
	metalgo.Client
}

func (c sizeImageConstraintCRUD) Get(id string) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.Sizeimageconstraint().FindSizeImageConstraint(sizemodel.NewFindSizeImageConstraintParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCRUD) List() ([]*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.Sizeimageconstraint().ListSizeImageConstraints(sizemodel.NewListSizeImageConstraintsParams(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCRUD) Delete(id string) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.Sizeimageconstraint().DeleteSizeImageConstraint(sizemodel.NewDeleteSizeImageConstraintParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCRUD) Create(rq *models.V1SizeImageConstraintCreateRequest) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.Sizeimageconstraint().CreateSizeImageConstraint(sizemodel.NewCreateSizeImageConstraintParams().WithBody(rq), nil)
	if err != nil {
		var r *size.CreateSizeConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c sizeImageConstraintCRUD) Update(rq *models.V1SizeImageConstraintUpdateRequest) (*models.V1SizeImageConstraintResponse, error) {
	resp, err := c.Sizeimageconstraint().UpdateSizeImageConstraint(sizemodel.NewUpdateSizeImageConstraintParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// non-generic command handling

func (c *sizeImageConstraintCmd) try() error {
	size := viper.GetString("size")
	image := viper.GetString("image")

	err := c.driver.TrySizeImageConstraint(size, image)
	if err != nil {
		return err
	}
	fmt.Println("allocation is possible")
	return nil
}
