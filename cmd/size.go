package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/tag"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeCmd struct {
	*config
}

func newSizeCmd(c *config) *cobra.Command {
	w := &sizeCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]{
		BinaryName:      binaryName,
		GenericCLI:      genericcli.NewGenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse](w).WithFS(c.fs),
		Singular:        "size",
		Plural:          "sizes",
		Description:     "a size matches a machine in terms of cpu cores, ram and storage.",
		Sorter:          sorters.SizeSorter(),
		ValidArgsFn:     c.comp.SizeListCompletion,
		DescribePrinter: func() printers.Printer { return c.describePrinter },
		ListPrinter:     func() printers.Printer { return c.listPrinter },
		CreateRequestFromCLI: func() (*models.V1SizeCreateRequest, error) {
			return &models.V1SizeCreateRequest{
				ID:          pointer.Pointer(viper.GetString("id")),
				Name:        viper.GetString("name"),
				Description: viper.GetString("description"),
				Constraints: []*models.V1SizeConstraint{
					{
						Max:  viper.GetInt64("max"),
						Min:  viper.GetInt64("min"),
						Type: pointer.Pointer(viper.GetString("type")),
					},
				},
			}, nil
		},
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("id", "", "", "ID of the size. [required]")
			cmd.Flags().StringP("name", "n", "", "Name of the size. [optional]")
			cmd.Flags().StringP("description", "d", "", "Description of the size. [required]")
			// FIXME constraints must be given in a slice
			cmd.Flags().Int64P("min", "", 0, "min value of given size constraint type. [required]")
			cmd.Flags().Int64P("max", "", 0, "min value of given size constraint type. [required]")
			cmd.Flags().StringP("type", "", "", "type of constraints. [required]")
		},
	}

	reservationsCmd := newSizeReservationsCmd(c)

	suggestCmd := &cobra.Command{
		Use:   "suggest <id>",
		Short: "suggest size from a given machine id",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.suggest(args)
		},
	}

	suggestCmd.Flags().String("machine-id", "", "Machine id used to create the size suggestion. [required]")
	suggestCmd.Flags().String("name", "suggested-size", "The name of the suggested size")
	suggestCmd.Flags().String("description", "a suggested size", "The description of the suggested size")
	suggestCmd.Flags().StringSlice("labels", []string{}, "labels to add to the size")

	genericcli.Must(suggestCmd.RegisterFlagCompletionFunc("machine-id", c.comp.MachineListCompletion))

	return genericcli.NewCmds(cmdsConfig, newSizeImageConstraintCmd(c), reservationsCmd, suggestCmd)
}

func (c *sizeCmd) Get(id string) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().FindSize(size.NewFindSizeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeCmd) List() ([]*models.V1SizeResponse, error) {
	resp, err := c.client.Size().ListSizes(size.NewListSizesParams(), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeCmd) Delete(id string) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().DeleteSize(size.NewDeleteSizeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeCmd) Create(rq *models.V1SizeCreateRequest) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().CreateSize(size.NewCreateSizeParams().WithBody(rq), nil)
	if err != nil {
		var r *size.CreateSizeConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeCmd) Update(rq *models.V1SizeUpdateRequest) (*models.V1SizeResponse, error) {
	resp, err := c.client.Size().UpdateSize(size.NewUpdateSizeParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *sizeCmd) Convert(r *models.V1SizeResponse) (string, *models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, error) {
	if r.ID == nil {
		return "", nil, nil, fmt.Errorf("id is nil")
	}
	return *r.ID, sizeResponseToCreate(r), sizeResponseToUpdate(r), nil
}

func sizeResponseToCreate(r *models.V1SizeResponse) *models.V1SizeCreateRequest {
	var constraints []*models.V1SizeConstraint
	for i := range r.Constraints {
		constraints = append(constraints, &models.V1SizeConstraint{
			Max:        r.Constraints[i].Max,
			Min:        r.Constraints[i].Min,
			Type:       r.Constraints[i].Type,
			Identifier: r.Constraints[i].Identifier,
		})
	}
	return &models.V1SizeCreateRequest{
		Constraints: constraints,
		Description: r.Description,
		ID:          r.ID,
		Name:        r.Name,
	}
}

func sizeResponseToUpdate(r *models.V1SizeResponse) *models.V1SizeUpdateRequest {
	var constraints []*models.V1SizeConstraint
	for i := range r.Constraints {
		constraints = append(constraints, &models.V1SizeConstraint{
			Max:        r.Constraints[i].Max,
			Min:        r.Constraints[i].Min,
			Type:       r.Constraints[i].Type,
			Identifier: r.Constraints[i].Identifier,
		})
	}
	return &models.V1SizeUpdateRequest{
		Constraints: constraints,
		Description: r.Description,
		ID:          r.ID,
		Name:        r.Name,
		Labels:      r.Labels,
	}
}

// non-generic command handling

func (c *sizeCmd) suggest(args []string) error {
	sizeid, _ := genericcli.GetExactlyOneArg(args)

	var (
		machineid   = viper.GetString("machine-id")
		name        = viper.GetString("name")
		description = viper.GetString("description")
		labels      = tag.NewTagMap(viper.GetStringSlice("labels"))
		now         = time.Now()
	)

	if sizeid == "" {
		sizeid = uuid.NewString()
	}

	if machineid == "" {
		return fmt.Errorf("machine-id flag is required")
	}

	resp, err := c.client.Size().Suggest(size.NewSuggestParams().WithBody(&models.V1SizeSuggestRequest{
		MachineID: &machineid,
	}), nil)
	if err != nil {
		return err
	}

	return c.describePrinter.Print(&models.V1SizeResponse{
		ID:          &sizeid,
		Name:        name,
		Description: description,
		Constraints: resp.Payload,
		Labels:      labels,
		Changed:     strfmt.DateTime(now),
		Created:     strfmt.DateTime(now),
	})
}
