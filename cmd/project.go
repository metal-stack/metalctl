package cmd

import (
	"errors"
	"fmt"

	v1 "github.com/metal-stack/masterdata-api/api/rest/v1"
	metalgo "github.com/metal-stack/metal-go"
	projectmodel "github.com/metal-stack/metal-go/api/client/project"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/output"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newProjectCmd(c *config) *cobra.Command {
	genericCLI := genericcli.NewGenericCLI[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse](projectGeneric{c: c.client})

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
			return c.projectList()
		},
	}
	projectDescribeCmd := &cobra.Command{
		Use:   "describe <projectID>",
		Short: "describe a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectDescribe(args)
		},
		ValidArgsFunction: c.comp.ProjectListCompletion,
	}
	projectCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "create a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectCreate()
		},
		PreRun: bindPFlags,
	}
	projectDeleteCmd := &cobra.Command{
		Use:     "delete <projectID>",
		Short:   "delete a project",
		Aliases: []string{"destroy", "rm", "remove"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectDelete(args)
		},
		PreRun:            bindPFlags,
		ValidArgsFunction: c.comp.ProjectListCompletion,
	}
	projectApplyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create/update a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectApply(genericCLI)
		},
		PreRun: bindPFlags,
	}
	projectEditCmd := &cobra.Command{
		Use:   "edit <projectID>",
		Short: "edit a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.projectEdit(args, genericCLI)
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

	projectListCmd.Flags().StringP("name", "", "", "Name of the project.")
	projectListCmd.Flags().StringP("id", "", "", "ID of the project.")
	projectListCmd.Flags().StringP("tenant", "", "", "tenant of this project.")

	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDescribeCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectApplyCmd)
	projectCmd.AddCommand(projectEditCmd)

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

func (a projectGeneric) Create(rq *models.V1ProjectCreateRequest) (**models.V1ProjectResponse, error) {
	resp, err := a.c.Project().CreateProject(projectmodel.NewCreateProjectParams().WithBody(rq), nil)
	if err != nil {
		var r *projectmodel.CreateProjectConflict
		if errors.As(err, &r) {
			return nil, nil
		}
		return nil, err
	}

	return &resp.Payload, nil
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

func (c *config) projectList() error {
	if atLeastOneViperStringFlagGiven("id", "name", "tenant") {
		pfr := v1.ProjectFindRequest{}
		id := viper.GetString("id")
		name := viper.GetString("name")
		tenantID := viper.GetString("tenant")

		if id != "" {
			pfr.Id = &id
		}
		if name != "" {
			pfr.Name = &name
		}
		if tenantID != "" {
			pfr.TenantId = &tenantID
		}
		resp, err := c.driver.ProjectFind(pfr)
		if err != nil {
			return err
		}
		return output.New().Print(resp.Project)
	}
	resp, err := c.driver.ProjectList()
	if err != nil {
		return err
	}
	return output.New().Print(resp.Project)
}

func (c *config) projectDescribe(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no project ID given")
	}
	projectID := args[0]
	resp, err := c.driver.ProjectGet(projectID)
	if err != nil {
		return err
	}
	return output.NewDetailer().Detail(resp.Project)
}

func (c *config) projectCreate() error {
	tenant := viper.GetString("tenant")
	name := viper.GetString("name")
	desc := viper.GetString("description")
	labels := viper.GetStringSlice("label")
	as := viper.GetStringSlice("annotation")
	var (
		clusterQuota, machineQuota, ipQuota *v1.Quota
	)
	if viper.IsSet("cluster-quota") {
		q := viper.GetInt32("cluster-quota")
		clusterQuota = &v1.Quota{Quota: &q}
	}
	if viper.IsSet("machine-quota") {
		q := viper.GetInt32("machine-quota")
		machineQuota = &v1.Quota{Quota: &q}
	}
	if viper.IsSet("ip-quota") {
		q := viper.GetInt32("ip-quota")
		ipQuota = &v1.Quota{Quota: &q}
	}

	annotations, err := annotationsAsMap(as)
	if err != nil {
		return err
	}

	p := v1.Project{
		Name:        name,
		Description: desc,
		TenantId:    tenant,
		Quotas: &v1.QuotaSet{
			Cluster: clusterQuota,
			Machine: machineQuota,
			Ip:      ipQuota,
		},
		Meta: &v1.Meta{
			Kind:        "Project",
			Apiversion:  "v1",
			Annotations: annotations,
			Labels:      labels,
		},
	}
	pcr := v1.ProjectCreateRequest{
		Project: p,
	}

	response, err := c.driver.ProjectCreate(pcr)
	if err != nil {
		return err
	}

	return output.New().Print(response.Project)
}

func (c *config) projectApply(g *genericcli.GenericCLI[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse]) error {
	response, err := g.ApplyFromFile(viper.GetString("file"))
	if err != nil {
		return err
	}

	return output.New().Print(response)
}

func (c *config) projectEdit(args []string, g *genericcli.GenericCLI[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse]) error {
	id, err := projectID("edit", args)
	if err != nil {
		return err
	}

	response, err := g.Edit(id)
	if err != nil {
		return err
	}

	return output.New().Print(response)
}

func (c *config) projectDelete(args []string) error {
	id, err := projectID("delete", args)
	if err != nil {
		return err
	}

	response, err := c.driver.ProjectDelete(id)
	if err != nil {
		return err
	}

	return output.New().Print(response.Project)
}

func projectID(verb string, args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("project %s requires projectID as argument", verb)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return "", fmt.Errorf("project %s requires exactly one projectID as argument", verb)
}
