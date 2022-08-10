package cmd

import (
	"errors"

	projectmodel "github.com/metal-stack/metal-go/api/client/project"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/defaultscmds"
	"github.com/metal-stack/metalctl/cmd/sorters"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type projectCmd struct {
	*config
}

func newProjectCmd(c *config) *cobra.Command {
	w := projectCmd{
		config: c,
	}

	cmds := defaultscmds.New(&defaultscmds.Config[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse]{
		GenericCLI:           genericcli.NewGenericCLI[*models.V1ProjectCreateRequest, *models.V1ProjectUpdateRequest, *models.V1ProjectResponse](w),
		Singular:             "project",
		Plural:               "projects",
		Description:          "a project groups multiple networks for a tenant.",
		AvailableSortKeys:    sorters.ProjectSortKeys(),
		ValidArgsFunc:        c.comp.ProjectListCompletion,
		CreateRequestFromCLI: w.createFromCLI,
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("name", "", "name of the project, max 10 characters. [required]")
			cmd.Flags().String("description", "", "description of the project. [required]")
			cmd.Flags().String("tenant", "", "create project for given tenant")
			cmd.Flags().StringSlice("label", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
			cmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
			cmd.Flags().Int32("cluster-quota", 0, "cluster quota")
			cmd.Flags().Int32("machine-quota", 0, "machine quota")
			cmd.Flags().Int32("ip-quota", 0, "ip quota")
			must(cmd.MarkFlagRequired("name"))
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("name", "", "", "Name of the project.")
			cmd.Flags().StringP("id", "", "", "ID of the project.")
			cmd.Flags().StringP("tenant", "", "", "tenant of this project.")
			must(viper.BindPFlags(cmd.Flags()))
		},
	})

	return cmds.Build()
}

func (c projectCmd) Get(id string) (*models.V1ProjectResponse, error) {
	resp, err := c.client.Project().FindProject(projectmodel.NewFindProjectParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCmd) List() ([]*models.V1ProjectResponse, error) {
	resp, err := c.client.Project().FindProjects(projectmodel.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{
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

func (c projectCmd) Delete(id string) (*models.V1ProjectResponse, error) {
	resp, err := c.client.Project().DeleteProject(projectmodel.NewDeleteProjectParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCmd) Create(rq *models.V1ProjectCreateRequest) (*models.V1ProjectResponse, error) {
	resp, err := c.client.Project().CreateProject(projectmodel.NewCreateProjectParams().WithBody(rq), nil)
	if err != nil {
		var r *projectmodel.CreateProjectConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c projectCmd) Update(rq *models.V1ProjectUpdateRequest) (*models.V1ProjectResponse, error) {
	resp, err := c.client.Project().FindProject(projectmodel.NewFindProjectParams().WithID(rq.Meta.ID), nil)
	if err != nil {
		return nil, err
	}

	// FIXME: should not be done by the client, see https://github.com/fi-ts/cloudctl/pull/26
	rq.Meta.Version = resp.Payload.Meta.Version + 1

	updateResp, err := c.client.Project().UpdateProject(projectmodel.NewUpdateProjectParams().WithBody(rq), nil)
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
