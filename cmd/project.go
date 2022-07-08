package cmd

import (
	"errors"

	metalgo "github.com/metal-stack/metal-go"
	projectmodel "github.com/metal-stack/metal-go/api/client/project"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type projectCmd struct {
	c      metalgo.Client
	driver *metalgo.Driver
	gcli   *genericcli.GenericCLI[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse]
}

func newProjectCmd(c *config) *cobra.Command {
	w := projectCmd{
		c:      c.client,
		driver: c.driver,
		gcli:   genericcli.NewGenericCLI[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse](projectGeneric{c: c.client}),
	}

	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "manage projects",
		Long:  "a project groups multiple networks for a tenant",
	}

	projectListCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.list()
		},
	}
	projectDescribeCmd := &cobra.Command{
		Use:   "describe <projectID>",
		Short: "describe a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.DescribeAndPrint(args, genericcli.NewYAMLPrinter())
		},
		ValidArgsFunction: c.comp.ProjectListCompletion,
	}
	projectCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.IsSet("file") {
				return w.gcli.CreateFromFileAndPrint(viper.GetString("file"), genericcli.NewYAMLPrinter())
			}
			return w.create()
		},
		PreRun: bindPFlags,
	}
	projectUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.UpdateFromFileAndPrint(viper.GetString("file"), genericcli.NewYAMLPrinter())
		},
		PreRun: bindPFlags,
	}
	projectDeleteCmd := &cobra.Command{
		Use:     "delete <projectID>",
		Short:   "delete a project",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.DeleteAndPrint(args, genericcli.NewYAMLPrinter())
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.ProjectListCompletion,
	}
	projectApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.ApplyFromFileAndPrint(viper.GetString("file"), output.New())
		},
		PreRun: bindPFlags,
	}
	projectEditCmd := &cobra.Command{
		Use:   "edit <projectID>",
		Short: "edit a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.gcli.EditAndPrint(args, genericcli.NewYAMLPrinter())
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.ProjectListCompletion,
	}

	projectCreateCmd.Flags().String("name", "", "name of the project, max 10 characters. [required]")
	projectCreateCmd.Flags().String("description", "", "description of the project. [required]")
	projectCreateCmd.Flags().String("tenant", "", "create project for given tenant")
	projectCreateCmd.Flags().StringSlice("label", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
	projectCreateCmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
	projectCreateCmd.Flags().Int32("cluster-quota", 0, "cluster quota")
	projectCreateCmd.Flags().Int32("machine-quota", 0, "machine quota")
	projectCreateCmd.Flags().Int32("ip-quota", 0, "ip quota")
	must(projectCreateCmd.MarkFlagRequired("name"))

	projectApplyCmd.Flags().StringP("file", "f", "", `filename of the create or update request in yaml format, or - for stdin.
Example project update:

# cloudctl project describe project1 -o yaml > project1.yaml
# vi project1.yaml
## either via stdin
# cat project1.yaml | cloudctl project apply -f -
## or via file
# cloudctl project apply -f project1.yaml
`)
	must(projectApplyCmd.MarkFlagRequired("file"))

	projectUpdateCmd.Flags().StringP("file", "f", "", "filename of the update request in yaml format, or - for stdin.")
	must(projectUpdateCmd.MarkFlagRequired("file"))

	projectListCmd.Flags().StringP("name", "", "", "Name of the project.")
	projectListCmd.Flags().StringP("id", "", "", "ID of the project.")
	projectListCmd.Flags().StringP("tenant", "", "", "tenant of this project.")

	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDescribeCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectApplyCmd)
	projectCmd.AddCommand(projectEditCmd)
	projectCmd.AddCommand(projectUpdateCmd)

	must(viper.BindPFlags(projectListCmd.Flags()))

	return projectCmd
}

type projectGeneric struct {
	c metalgo.Client
}

func (a projectGeneric) Get(id string) (*models.V1ProjectResponse, error) {
	resp, err := a.c.Project().FindProject(projectmodel.NewFindProjectParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (a projectGeneric) Delete(id string) (*models.V1ProjectResponse, error) {
	resp, err := a.c.Project().DeleteProject(projectmodel.NewDeleteProjectParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (a projectGeneric) Create(rq *models.V1ProjectCreateRequest) (*models.V1ProjectResponse, error) {
	resp, err := a.c.Project().CreateProject(projectmodel.NewCreateProjectParams().WithBody(rq), nil)
	if err != nil {
		var r *projectmodel.CreateProjectConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (a projectGeneric) Update(rq *models.V1ProjectUpdateRequest) (*models.V1ProjectResponse, error) {
	resp, err := a.c.Project().FindProject(projectmodel.NewFindProjectParams().WithID(rq.Meta.ID), nil)
	if err != nil {
		return nil, err
	}

	// FIXME: should not be done by the client, see https://github.com/fi-ts/cloudctl/pull/26
	rq.Meta.Version = resp.Payload.Meta.Version + 1

	updateResp, err := a.c.Project().UpdateProject(projectmodel.NewUpdateProjectParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

// non-generic command handling

func (w *projectCmd) list() error {
	if atLeastOneViperStringFlagGiven("id", "name", "tenant") {
		rq := &models.V1ProjectFindRequest{
			ID:       viper.GetString("id"),
			Name:     viper.GetString("name"),
			TenantID: viper.GetString("tenant"),
		}

		resp, err := w.c.Project().FindProjects(projectmodel.NewFindProjectsParams().WithBody(rq), nil)
		if err != nil {
			return err
		}

		return output.New().Print(resp.Payload)
	}

	resp, err := w.c.Project().ListProjects(projectmodel.NewListProjectsParams(), nil)
	if err != nil {
		return err
	}

	return output.New().Print(resp.Payload)
}

func (w *projectCmd) create() error {
	var (
		clusterQuota, machineQuota, ipQuota *models.V1Quota
	)
	if viper.IsSet("cluster-quota") {
		clusterQuota = &models.V1Quota{Quota: viper.GetInt32("cluster-quota")}
	}
	if viper.IsSet("machine-quota") {
		machineQuota = &models.V1Quota{Quota: viper.GetInt32("machine-quota")}
	}
	if viper.IsSet("ip-quota") {
		ipQuota = &models.V1Quota{Quota: viper.GetInt32("ip-quota")}
	}

	annotations, err := genericcli.LabelsToMap(viper.GetStringSlice("annotation"))
	if err != nil {
		return err
	}

	return w.gcli.CreateAndPrint(&models.V1ProjectCreateRequest{
		Name:        viper.GetString("name"),
		Description: viper.GetString("description"),
		TenantID:    viper.GetString("tenant"),
		Quotas: &models.V1QuotaSet{
			Cluster: clusterQuota,
			Machine: machineQuota,
			IP:      ipQuota,
		},
		Meta: &models.V1Meta{
			Kind:        "Project",
			Apiversion:  "v1",
			Annotations: annotations,
			Labels:      viper.GetStringSlice("label"),
		},
	}, genericcli.NewYAMLPrinter())
}
