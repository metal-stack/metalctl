package cmd

import (
	"strconv"
	"strings"
	"testing"

	"github.com/metal-stack/metal-go/api/client/tenant"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	tenant1 = &models.V1TenantResponse{
		Meta: &models.V1Meta{
			Kind:       "Tenant",
			Apiversion: "v1",
			ID:         "1",
			Annotations: map[string]string{
				"a": "b",
			},
			Labels:  []string{"c"},
			Version: 1,
		},
		Description: "tenant 1",
		Name:        "tenant-1",
		Quotas: &models.V1QuotaSet{
			Cluster: &models.V1Quota{
				Quota: 1,
				Used:  1,
			},
			IP: &models.V1Quota{
				Quota: 2,
				Used:  2,
			},
			Machine: &models.V1Quota{
				Quota: 3,
				Used:  3,
			},
		},
	}
	tenant2 = &models.V1TenantResponse{
		Meta: &models.V1Meta{
			Kind:       "Tenant",
			Apiversion: "v1",
			ID:         "2",
			Annotations: map[string]string{
				"a": "b",
			},
			Labels:  []string{"c"},
			Version: 1,
		},
		Description: "tenant 2",
		Name:        "tenant-2",
		Quotas: &models.V1QuotaSet{
			Cluster: &models.V1Quota{},
			IP:      &models.V1Quota{},
			Machine: &models.V1Quota{},
		},
	}
)

func Test_TenantCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1TenantResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1TenantResponse) []string {
				return []string{"tenant", "list"}
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("FindTenants", testcommon.MatchIgnoreContext(t, tenant.NewFindTenantsParams().WithBody(&models.V1TenantFindRequest{})), nil).Return(&tenant.FindTenantsOK{
						Payload: []*models.V1TenantResponse{
							tenant2,
							tenant1,
						},
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
				tenant2,
			},
			wantTable: pointer.Pointer(`
ID   NAME       DESCRIPTION
1    tenant-1   tenant 1
2    tenant-2   tenant 2
`),
			wantWideTable: pointer.Pointer(`
ID   NAME       DESCRIPTION   LABELS   ANNOTATIONS   QUOTAS
1    tenant-1   tenant 1      c        a=b           1 Cluster(s)
                                                     3 Machine(s)
                                                     2 IP(s)
2    tenant-2   tenant 2      c        a=b           ∞ Cluster(s)
                                                     ∞ Machine(s)
                                                     ∞ IP(s)
`),
			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 tenant-1
2 tenant-2
`),
			wantMarkdown: pointer.Pointer(`
| ID |   NAME   | DESCRIPTION |
|----|----------|-------------|
|  1 | tenant-1 | tenant 1    |
|  2 | tenant-2 | tenant 2    |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1TenantResponse) []string {
				args := []string{"tenant", "list", "--name", "tenant-1", "--annotations", "a=b", "--id", want[0].Meta.ID}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("FindTenants", testcommon.MatchIgnoreContext(t, tenant.NewFindTenantsParams().WithBody(&models.V1TenantFindRequest{
						Name: "tenant-1",
						ID:   "1",
						Annotations: map[string]string{
							"a": "b",
						},
					})), nil).Return(&tenant.FindTenantsOK{
						Payload: []*models.V1TenantResponse{
							tenant1,
						},
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
			},
			wantTable: pointer.Pointer(`
ID   NAME       DESCRIPTION
1    tenant-1   tenant 1
        `),
			wantWideTable: pointer.Pointer(`
ID   NAME       DESCRIPTION   LABELS   ANNOTATIONS   QUOTAS
1    tenant-1   tenant 1      c        a=b           1 Cluster(s)
                                                     3 Machine(s)
                                                     2 IP(s)
        `),
			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
        1 tenant-1
        `),
			wantMarkdown: pointer.Pointer(`
| ID |   NAME   | DESCRIPTION |
|----|----------|-------------|
|  1 | tenant-1 | tenant 1    |
        `),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1TenantResponse) []string {
				return appendFromFileCommonArgs("tenant", "apply")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1TenantResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("CreateTenant", testcommon.MatchIgnoreContext(t, tenant.NewCreateTenantParams().WithBody(tenantResponseToCreate(tenant1))), nil).Return(nil, &tenant.CreateTenantConflict{}).Once()
					mock.On("GetTenant", testcommon.MatchIgnoreContext(t, tenant.NewGetTenantParams().WithID(tenant1.Meta.ID)), nil).Return(&tenant.GetTenantOK{
						Payload: tenant1,
					}, nil)
					mock.On("UpdateTenant", testcommon.MatchIgnoreContext(t, tenant.NewUpdateTenantParams().WithBody(tenantResponseToUpdate(tenant1))), nil).Return(&tenant.UpdateTenantOK{
						Payload: tenant1,
					}, nil)
					mock.On("CreateTenant", testcommon.MatchIgnoreContext(t, tenant.NewCreateTenantParams().WithBody(tenantResponseToCreate(tenant2))), nil).Return(&tenant.CreateTenantCreated{
						Payload: tenant2,
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
				tenant2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1TenantResponse) []string {
				return appendFromFileCommonArgs("tenant", "create")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1TenantResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("CreateTenant", testcommon.MatchIgnoreContext(t, tenant.NewCreateTenantParams().WithBody(tenantResponseToCreate(tenant1))), nil).Return(&tenant.CreateTenantCreated{
						Payload: tenant1,
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1TenantResponse) []string {
				return appendFromFileCommonArgs("tenant", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1TenantResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("GetTenant", testcommon.MatchIgnoreContext(t, tenant.NewGetTenantParams().WithID(tenant1.Meta.ID)), nil).Return(&tenant.GetTenantOK{
						Payload: tenant1,
					}, nil)
					mock.On("UpdateTenant", testcommon.MatchIgnoreContext(t, tenant.NewUpdateTenantParams().WithBody(tenantResponseToUpdate(tenant1))), nil).Return(&tenant.UpdateTenantOK{
						Payload: tenant1,
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1TenantResponse) []string {
				return appendFromFileCommonArgs("tenant", "delete")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1TenantResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("DeleteTenant", testcommon.MatchIgnoreContext(t, tenant.NewDeleteTenantParams().WithID(tenant1.Meta.ID)), nil).Return(&tenant.DeleteTenantOK{
						Payload: tenant1,
					}, nil)
				},
			},
			want: []*models.V1TenantResponse{
				tenant1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_TenantCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1TenantResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1TenantResponse) []string {
				return []string{"tenant", "describe", want.Meta.ID}
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("GetTenant", testcommon.MatchIgnoreContext(t, tenant.NewGetTenantParams().WithID(tenant1.Meta.ID)), nil).Return(&tenant.GetTenantOK{
						Payload: tenant1,
					}, nil)
				},
			},
			want: tenant1,
			wantTable: pointer.Pointer(`
ID   NAME       DESCRIPTION
1    tenant-1   tenant 1
`),
			wantWideTable: pointer.Pointer(`
ID   NAME       DESCRIPTION   LABELS   ANNOTATIONS   QUOTAS
1    tenant-1   tenant 1      c        a=b           1 Cluster(s)
                                                     3 Machine(s)
                                                     2 IP(s)
`),
			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 tenant-1
`),
			wantMarkdown: pointer.Pointer(`
| ID |   NAME   | DESCRIPTION |
|----|----------|-------------|
|  1 | tenant-1 | tenant 1    |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1TenantResponse) []string {
				return []string{"tenant", "rm", want.Meta.ID}
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					mock.On("DeleteTenant", testcommon.MatchIgnoreContext(t, tenant.NewDeleteTenantParams().WithID(tenant1.Meta.ID)), nil).Return(&tenant.DeleteTenantOK{
						Payload: tenant1,
					}, nil)
				},
			},
			want: tenant1,
		},
		{
			name: "create",
			cmd: func(want *models.V1TenantResponse) []string {
				args := []string{"tenant", "create",
					"--id", want.Meta.ID,
					"--name", want.Name,
					"--description", want.Description,
					"--labels", strings.Join(want.Meta.Labels, ","),
					"--annotations", strings.Join(genericcli.MapToLabels(want.Meta.Annotations), ","),
					"--cluster-quota", strconv.FormatInt(int64(want.Quotas.Cluster.Quota), 10),
					"--machine-quota", strconv.FormatInt(int64(want.Quotas.Machine.Quota), 10),
					"--ip-quota", strconv.FormatInt(int64(want.Quotas.IP.Quota), 10),
				}
				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
				return args
			},
			mocks: &client.MetalMockFns{
				Tenant: func(mock *mock.Mock) {
					p := tenant1
					p.Meta.Version = 0
					p.Quotas.Cluster.Used = 0
					p.Quotas.IP.Used = 0
					p.Quotas.Machine.Used = 0
					mock.On("CreateTenant", testcommon.MatchIgnoreContext(t, tenant.NewCreateTenantParams().WithBody(tenantResponseToCreate(p))), nil).Return(&tenant.CreateTenantCreated{
						Payload: tenant1,
					}, nil)
				},
			},
			want: tenant1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
