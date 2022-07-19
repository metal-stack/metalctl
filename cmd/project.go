package cmd

import (
	"errors"

	metalgo "github.com/metal-stack/metal-go"
	projectmodel "github.com/metal-stack/metal-go/api/client/project"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/sorters"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type projectCmd struct {
	c metalgo.Client
	*genericcli.GenericCLI[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse]
}

func newProjectCmd(c *config) *cobra.Command {
	w := projectCmd{
		c:          c.client,
		GenericCLI: genericcli.NewGenericCLI[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse](projectCRUD{Client: c.client}),
	}

	cmds := newDefaultCmds(&defaultCmdsConfig[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse]{
		gcli:                 w.GenericCLI,
		singular:             "project",
		plural:               "projects",
		description:          "a project groups multiple networks for a tenant.",
		availableSortKeys:    sorters.ProjectSortKeys(),
		validArgsFunc:        c.comp.ProjectListCompletion,
		createRequestFromCLI: w.createFromCLI,
	})

	cmds.createCmd.Flags().String("name", "", "name of the project, max 10 characters. [required]")
	cmds.createCmd.Flags().String("description", "", "description of the project. [required]")
	cmds.createCmd.Flags().String("tenant", "", "create project for given tenant")
	cmds.createCmd.Flags().StringSlice("label", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
	cmds.createCmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
	cmds.createCmd.Flags().Int32("cluster-quota", 0, "cluster quota")
	cmds.createCmd.Flags().Int32("machine-quota", 0, "machine quota")
	cmds.createCmd.Flags().Int32("ip-quota", 0, "ip quota")
	must(cmds.createCmd.MarkFlagRequired("name"))

	cmds.listCmd.Flags().StringP("name", "", "", "Name of the project.")
	cmds.listCmd.Flags().StringP("id", "", "", "ID of the project.")
	cmds.listCmd.Flags().StringP("tenant", "", "", "tenant of this project.")
	must(viper.BindPFlags(cmds.listCmd.Flags()))

	return cmds.buildRootCmd()
}

type projectCRUD struct {
	metalgo.Client
}

func (c projectCRUD) Get(id string) (*models.V1ProjectResponse, error) {
	resp, err := c.Project().FindProject(projectmodel.NewFindProjectParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCRUD) List() ([]*models.V1ProjectResponse, error) {
	resp, err := c.Project().FindProjects(projectmodel.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{
		ID:       viper.GetString("id"),
		Name:     viper.GetString("name"),
		TenantID: viper.GetString("tenant"),
	}), nil)
	if err != nil {
		return nil, err
	}

	err = sorters.ProjectSort(resp.Payload)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCRUD) Delete(id string) (*models.V1ProjectResponse, error) {
	resp, err := c.Project().DeleteProject(projectmodel.NewDeleteProjectParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCRUD) Create(rq *models.V1ProjectCreateRequest) (*models.V1ProjectResponse, error) {
	resp, err := c.Project().CreateProject(projectmodel.NewCreateProjectParams().WithBody(rq), nil)
	if err != nil {
		var r *projectmodel.CreateProjectConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCRUD) Update(rq *models.V1ProjectUpdateRequest) (*models.V1ProjectResponse, error) {
	resp, err := c.Project().FindProject(projectmodel.NewFindProjectParams().WithID(rq.Meta.ID), nil)
	if err != nil {
		return nil, err
	}

	// FIXME: should not be done by the client, see https://github.com/fi-ts/cloudctl/pull/26
	rq.Meta.Version = resp.Payload.Meta.Version + 1

	updateResp, err := c.Project().UpdateProject(projectmodel.NewUpdateProjectParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

func (w *projectCmd) createFromCLI() (*models.V1ProjectCreateRequest, error) {
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
		return nil, err
	}

	return &models.V1ProjectCreateRequest{
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
	}, nil
}
