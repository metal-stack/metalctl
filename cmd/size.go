package cmd

import (
	"errors"

	"github.com/dustin/go-humanize"
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type sizeCmdWrapper struct {
	c      metalgo.Client
	driver *metalgo.Driver
	gcli   *genericcli.GenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse]
}

func newSizeCmd(c *config) *cobra.Command {
	w := sizeCmdWrapper{
		c:      c.client,
		driver: c.driver,
		gcli:   genericcli.NewGenericCLI[*models.V1SizeCreateRequest, *models.V1SizeUpdateRequest, *models.V1SizeResponse](sizeGeneric{c: c.client}),
	}

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
			return w.list()
		},
		PreRun: bindPFlags,
	}
	sizeDescribeCmd := &cobra.Command{
		Use:   "describe <sizeID>",
		Short: "describe a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.describe(args)
		},
		ValidArgsFunction: c.comp.SizeListCompletion,
	}
	sizeTryCmd := &cobra.Command{
		Use:   "try",
		Short: "try a specific hardware spec and give the chosen size back",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.try()
		},
		PreRun: bindPFlags,
	}
	sizeCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.create()
		},
		PreRun: bindPFlags,
	}
	sizeUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.update()
		},
		PreRun: bindPFlags,
	}
	sizeApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.sizeApply()
		},
		PreRun: bindPFlags,
	}
	sizeDeleteCmd := &cobra.Command{
		Use:     "delete <sizeID>",
		Short:   "delete a size",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.delete(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.SizeListCompletion,
	}
	sizeEditCmd := &cobra.Command{
		Use:   "edit <sizeID>",
		Short: "edit a size",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.edit(args)
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

type sizeGeneric struct {
	c metalgo.Client
}

func (g sizeGeneric) Get(id string) (*models.V1SizeResponse, error) {
	resp, err := g.c.Size().FindSize(size.NewFindSizeParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (g sizeGeneric) Create(rq *models.V1SizeCreateRequest) (**models.V1SizeResponse, error) {
	resp, err := g.c.Size().CreateSize(size.NewCreateSizeParams().WithBody(rq), nil)
	if err != nil {
		var r *size.CreateSizeConflict
		if errors.As(err, &r) {
			return nil, nil
		}
		return nil, err
	}

	return &resp.Payload, nil
}

func (g sizeGeneric) Update(rq *models.V1SizeUpdateRequest) (*models.V1SizeResponse, error) {
	resp, err := g.c.Size().UpdateSize(size.NewUpdateSizeParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (w *sizeCmdWrapper) list() error {
	resp, err := w.c.Size().ListSizes(size.NewListSizesParams(), nil)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Payload)
}

func (w *sizeCmdWrapper) describe(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := w.gcli.Interface().Get(id)
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(resp)
}

func (w *sizeCmdWrapper) create() error {
	if viper.IsSet("file") {
		response, err := w.gcli.CreateFromFile(viper.GetString("file"))
		if err != nil {
			return err
		}

		return output.NewDetailer().Detail(response)
	}

	max := viper.GetInt64("min")
	min := viper.GetInt64("max")
	t := viper.GetString("type")

	icr := &models.V1SizeCreateRequest{
		Description: viper.GetString("description"),
		ID:          pointer.To(viper.GetString("id")),
		Name:        viper.GetString("name"),
		Constraints: []*models.V1SizeConstraint{
			{
				Max:  &max,
				Min:  &min,
				Type: &t,
			},
		},
	}

	resp, err := w.c.Size().CreateSize(size.NewCreateSizeParams().WithBody(icr), nil)
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(resp.Payload)
}

func (w *sizeCmdWrapper) update() error {
	response, err := w.gcli.UpdateFromFile(viper.GetString("file"))
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(response)
}

func (w *sizeCmdWrapper) delete(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	resp, err := w.c.Size().DeleteSize(size.NewDeleteSizeParams().WithID(id), nil)
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(resp.Payload)
}

func (w *sizeCmdWrapper) sizeApply() error {
	response, err := w.gcli.ApplyFromFile(viper.GetString("file"))
	if err != nil {
		return err
	}

	return output.New().Print(response)
}

func (w *sizeCmdWrapper) edit(args []string) error {
	id, err := genericcli.GetExactlyOneArg(args)
	if err != nil {
		return err
	}

	size, err := w.gcli.Edit(id)
	if err != nil {
		return err
	}

	return output.NewDetailer().Detail(size)
}

func (w *sizeCmdWrapper) try() error {
	cores := viper.GetInt32("cores")
	memory, err := humanize.ParseBytes(viper.GetString("memory"))
	if err != nil {
		return err
	}
	storagesize, err := humanize.ParseBytes(viper.GetString("storagesize"))
	if err != nil {
		return err
	}

	// TODO: replace driver with client
	resp, _ := w.driver.SizeTry(cores, memory, storagesize)

	return output.New().Print(resp.Logs)
}
