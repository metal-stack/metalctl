package cmd

import (
	"errors"
	"fmt"

	tenantmodel "github.com/metal-stack/metal-go/api/client/tenant"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metalctl/cmd/sorters"
	"github.com/metal-stack/metalctl/pkg/api"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type tenantCmd struct {
	*api.Config
}

func newTenantCmd(c *api.Config) *cobra.Command {
	w := &tenantCmd{
		Config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1TenantCreateRequest, *models.V1TenantUpdateRequest, *models.V1TenantResponse]{
		BinaryName:           binaryName,
		GenericCLI:           genericcli.NewGenericCLI[*models.V1TenantCreateRequest, *models.V1TenantUpdateRequest, *models.V1TenantResponse](w).WithFS(c.FS),
		Singular:             "tenant",
		Plural:               "tenants",
		Description:          "a tenant belongs to a tenant and groups together entities in metal-stack.",
		Sorter:               sorters.TenantSorter(),
		ValidArgsFn:          c.Comp.TenantListCompletion,
		DescribePrinter:      func() printers.Printer { return c.DescribePrinter },
		ListPrinter:          func() printers.Printer { return c.ListPrinter },
		CreateRequestFromCLI: w.createFromCLI,
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("id", "", "id of the tenant, max 10 characters.")
			cmd.Flags().String("name", "", "name of the tenant, max 10 characters.")
			cmd.Flags().String("description", "", "description of the tenant.")
			cmd.Flags().StringSlice("labels", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
			cmd.Flags().StringSlice("annotations", nil, "add initial annotations, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
			cmd.Flags().Int32("cluster-quota", 0, "cluster quota")
			cmd.Flags().Int32("machine-quota", 0, "machine quota")
			cmd.Flags().Int32("ip-quota", 0, "ip quota")

			cmd.MarkFlagsMutuallyExclusive("file", "name")
			cmd.MarkFlagsRequiredTogether("name", "description")
		},
		ListCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().StringP("name", "", "", "Name of the tenant.")
			cmd.Flags().StringP("id", "", "", "ID of the tenant.")
			cmd.Flags().StringSliceP("annotations", "", []string{}, "annotations")
		},
	}

	return genericcli.NewCmds(cmdsConfig)
}

func (c *tenantCmd) Get(id string) (*models.V1TenantResponse, error) {
	resp, err := c.Client.Tenant().GetTenant(tenantmodel.NewGetTenantParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *tenantCmd) List() ([]*models.V1TenantResponse, error) {
	var annotations map[string]string
	if viper.IsSet("annotations") {
		var err error
		annotations, err = genericcli.LabelsToMap(viper.GetStringSlice("annotations"))
		if err != nil {
			return nil, err
		}
	}

	resp, err := c.Client.Tenant().FindTenants(tenantmodel.NewFindTenantsParams().WithBody(&models.V1TenantFindRequest{
		ID:          viper.GetString("id"),
		Name:        viper.GetString("name"),
		Annotations: annotations,
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *tenantCmd) Delete(id string) (*models.V1TenantResponse, error) {
	resp, err := c.Client.Tenant().DeleteTenant(tenantmodel.NewDeleteTenantParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *tenantCmd) Create(rq *models.V1TenantCreateRequest) (*models.V1TenantResponse, error) {
	resp, err := c.Client.Tenant().CreateTenant(tenantmodel.NewCreateTenantParams().WithBody(rq), nil)
	if err != nil {
		var r *tenantmodel.CreateTenantConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c *tenantCmd) Update(rq *models.V1TenantUpdateRequest) (*models.V1TenantResponse, error) {
	if rq.Meta == nil {
		return nil, fmt.Errorf("tenant meta is nil")
	}

	getResp, err := c.Get(rq.Meta.ID)
	if err != nil {
		return nil, err
	}

	rq.Meta.Version = getResp.Meta.Version

	updateResp, err := c.Client.Tenant().UpdateTenant(tenantmodel.NewUpdateTenantParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

func (c *tenantCmd) Convert(r *models.V1TenantResponse) (string, *models.V1TenantCreateRequest, *models.V1TenantUpdateRequest, error) {
	if r.Meta == nil {
		return "", nil, nil, fmt.Errorf("meta is nil")
	}
	return r.Meta.ID, tenantResponseToCreate(r), tenantResponseToUpdate(r), nil
}

func tenantResponseToCreate(r *models.V1TenantResponse) *models.V1TenantCreateRequest {
	return &models.V1TenantCreateRequest{
		Meta: &models.V1Meta{
			Apiversion:  r.Meta.Apiversion,
			Kind:        r.Meta.Kind,
			ID:          r.Meta.ID,
			Annotations: r.Meta.Annotations,
			Labels:      r.Meta.Labels,
			Version:     r.Meta.Version,
		},
		Description: r.Description,
		Name:        r.Name,
		Quotas:      r.Quotas,
	}
}

func tenantResponseToUpdate(r *models.V1TenantResponse) *models.V1TenantUpdateRequest {
	return &models.V1TenantUpdateRequest{
		Name: r.Name,
		Meta: &models.V1Meta{
			Apiversion:  r.Meta.Apiversion,
			Kind:        r.Meta.Kind,
			ID:          r.Meta.ID,
			Annotations: r.Meta.Annotations,
			Labels:      r.Meta.Labels,
			Version:     r.Meta.Version,
		},
		Description:   r.Description,
		IamConfig:     r.IamConfig,
		DefaultQuotas: r.DefaultQuotas,
		Quotas:        r.Quotas,
	}
}

func (w *tenantCmd) createFromCLI() (*models.V1TenantCreateRequest, error) {
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

	annotations, err := genericcli.LabelsToMap(viper.GetStringSlice("annotations"))
	if err != nil {
		return nil, err
	}

	return &models.V1TenantCreateRequest{
			Name:        viper.GetString("name"),
			Description: viper.GetString("description"),
			Quotas: &models.V1QuotaSet{
				Cluster: clusterQuota,
				Machine: machineQuota,
				IP:      ipQuota,
			},
			Meta: &models.V1Meta{
				Kind:        "Tenant",
				Apiversion:  "v1",
				Annotations: annotations,
				Labels:      viper.GetStringSlice("labels"),
				ID:          viper.GetString("id"),
			},
		},
		nil
}
