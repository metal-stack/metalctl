package cmd

import (
	"errors"

	"github.com/dustin/go-humanize"
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeGeneric struct {
	c metalgo.Client
}

func (a sizeGeneric) Get(id string) (*models.V1SizeResponse, error) {
	resp, err := a.c.Size().FindSize(size.NewFindSizeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (a sizeGeneric) Create(rq *models.V1SizeCreateRequest) (**models.V1SizeResponse, error) {
	resp, err := a.c.Size().CreateSize(size.NewCreateSizeParams().WithBody(rq), nil)
	if err != nil {
		var r *size.CreateSizeConflict
		if errors.As(err, &r) {
			return nil, nil
		}
		return nil, err
	}

	return &resp.Payload, nil
}

func (a sizeGeneric) Update(rq *models.V1SizeUpdateRequest) (*models.V1SizeResponse, error) {
	resp, err := a.c.Size().UpdateSize(size.NewUpdateSizeParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func newSizeCmd(c *config) *cobra.Command {
	g := sizeGeneric{c: c.client}
	genericCLI := genericcli.NewGenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse](g)

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
			return c.sizeDescribe(args, g)
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
			return c.sizeCreate(genericCLI)
		},
		PreRun: bindPFlags,
	}
	sizeUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeUpdate(genericCLI)
		},
		PreRun: bindPFlags,
	}
	sizeApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.sizeApply(genericCLI)
		},
		PreRun: bindPFlags,
	}
	sizeDeleteCmd := &cobra.Command{
		Use:     "delete <sizeID>",
		Short:   "delete a size",
		Aliases: []string{"destroy", "rm", "remove"},
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
			return c.sizeEdit(args, genericCLI)
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

func (c *config) sizeDescribe(args []string, g genericcli.Generic[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := g.Get(id)
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(resp)
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

func (c *config) sizeCreate(g *genericcli.GenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]) error {
	if viper.GetString("file") != "" {
		response, err := g.CreateFromFile(viper.GetString("file"))
		if err != nil {
			return err
		}

		return output.NewDetailer().Detail(response)
	}

	max := viper.GetInt64("min")
	min := viper.GetInt64("max")
	t := viper.GetString("type")

	icr := metalgo.SizeCreateRequest{
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

	resp, err := c.driver.SizeCreate(icr)
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(resp.Size)
}

func (c *config) sizeUpdate(g *genericcli.GenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]) error {
	response, err := g.UpdateFromFile(viper.GetString("file"))
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(response)
}

func (c *config) sizeApply(g *genericcli.GenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]) error {
	response, err := g.ApplyFromFile(viper.GetString("file"))
	if err != nil {
		return err
	}

	return output.New().Print(response)
}

func (c *config) sizeDelete(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := c.driver.SizeDelete(id)
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(resp.Size)
}

func (c *config) sizeEdit(args []string, g *genericcli.GenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	size, err := g.Edit(id)
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(size)
}
