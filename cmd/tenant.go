package cmd

import (
	"errors"

	tenantmodel "github.com/metal-stack/metal-go/api/client/tenant"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metalctl/cmd/sorters"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type tenantCmd struct {
	*config
}

func newTenantCmd(c *config) *cobra.Command {
	w := tenantCmd{
		config: c,
	}

	cmdsConfig := &genericcli.CmdsConfig[*models.V1TenantCreateRequest, *models.V1TenantUpdateRequest, *models.V1TenantResponse]{
		BinaryName:           binaryName,
		GenericCLI:           genericcli.NewGenericCLI[*models.V1TenantCreateRequest, *models.V1TenantUpdateRequest, *models.V1TenantResponse](w).WithFS(c.fs),
		Singular:             "tenant",
		Plural:               "tenants",
		Description:          "a tenant belongs to a tenant and groups together entities in metal-stack.",
		Sorter:               sorters.TenantSorter(),
		ValidArgsFn:          c.comp.TenantListCompletion,
		DescribePrinter:      func() printers.Printer { return c.describePrinter },
		ListPrinter:          func() printers.Printer { return c.listPrinter },
		CreateRequestFromCLI: w.createFromCLI,
		CreateCmdMutateFn: func(cmd *cobra.Command) {
			cmd.Flags().String("name", "", "name of the tenant, max 10 characters.")
			cmd.Flags().String("description", "", "description of the tenant.")
			cmd.Flags().String("tenant", "", "create tenant for given tenant")
			cmd.Flags().StringSlice("label", nil, "add initial label, can be given multiple times to add multiple labels, e.g. --label=foo --label=bar")
			cmd.Flags().StringSlice("annotation", nil, "add initial annotation, must be in the form of key=value, can be given multiple times to add multiple annotations, e.g. --annotation key=value --annotation foo=bar")
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

func (c tenantCmd) Get(id string) (*models.V1TenantResponse, error) {
	resp, err := c.client.Tenant().GetTenant(tenantmodel.NewGetTenantParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c tenantCmd) List() ([]*models.V1TenantResponse, error) {
	annotations, err := genericcli.LabelsToMap(viper.GetStringSlice("annotation"))
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Tenant().FindTenants(tenantmodel.NewFindTenantsParams().WithBody(&models.V1TenantFindRequest{
		ID:          viper.GetString("id"),
		Name:        viper.GetString("name"),
		Annotations: annotations,
	}), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c tenantCmd) Delete(id string) (*models.V1TenantResponse, error) {
	resp, err := c.client.Tenant().DeleteTenant(tenantmodel.NewDeleteTenantParams().WithID(id), nil)
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c tenantCmd) Create(rq *models.V1TenantCreateRequest) (*models.V1TenantResponse, error) {
	resp, err := c.client.Tenant().CreateTenant(tenantmodel.NewCreateTenantParams().WithBody(rq), nil)
	if err != nil {
		var r *tenantmodel.CreateTenantConflict
		if errors.As(err, &r) {
			return nil, genericcli.AlreadyExistsError()
		}
		return nil, err
	}

	return resp.Payload, nil
}

func (c tenantCmd) Update(rq *models.V1TenantUpdateRequest) (*models.V1TenantResponse, error) {
	resp, err := c.client.Tenant().GetTenant(tenantmodel.NewGetTenantParams().WithID(rq.Meta.ID), nil)
	if err != nil {
		return nil, err
	}

	// FIXME: should not be done by the client, see https://github.com/fi-ts/cloudctl/pull/26
	rq.Meta.Version = resp.Payload.Meta.Version + 1

	updateResp, err := c.client.Tenant().UpdateTenant(tenantmodel.NewUpdateTenantParams().WithBody(rq), nil)
	if err != nil {
		return nil, err
	}

	return updateResp.Payload, nil
}

func (c tenantCmd) ToCreate(r *models.V1TenantResponse) (*models.V1TenantCreateRequest, error) {
	return tenantResponseToCreate(r), nil
}

func (c tenantCmd) ToUpdate(r *models.V1TenantResponse) (*models.V1TenantUpdateRequest, error) {
	return tenantResponseToUpdate(r), nil
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
		IamConfig:     &models.V1IAMConfig{},
		DefaultQuotas: &models.V1QuotaSet{},
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

	annotations, err := genericcli.LabelsToMap(viper.GetStringSlice("annotation"))
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
				Labels:      viper.GetStringSlice("label"),
			},
		},
		nil
}
