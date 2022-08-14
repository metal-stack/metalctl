package cmd

import (
	"strconv"
	"strings"
	"testing"

	"github.com/metal-stack/metal-go/api/client/project"
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
	project1 = &models.V1ProjectResponse{
		Meta: &models.V1Meta{
			Kind:       "Project",
			Apiversion: "v1",
			ID:         "1",
			Annotations: map[string]string{
				"a": "b",
			},
			Labels:  []string{"c"},
			Version: 1,
		},
		Description: "project 1",
		Name:        "project-1",
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
		TenantID: "metal-stack",
	}
	project2 = &models.V1ProjectResponse{
		Meta: &models.V1Meta{
			Kind:       "Project",
			Apiversion: "v1",
			ID:         "2",
			Annotations: map[string]string{
				"a": "b",
			},
			Labels:  []string{"c"},
			Version: 1,
		},
		Description: "project 2",
		Name:        "project-2",
		Quotas: &models.V1QuotaSet{
			Cluster: &models.V1Quota{},
			IP:      &models.V1Quota{},
			Machine: &models.V1Quota{},
		},
		TenantID: "metal-stack",
	}
	toProjectCreateRequestFromCLI = func(s *models.V1ProjectResponse) *models.V1ProjectCreateRequest {
		return &models.V1ProjectCreateRequest{
			Meta: &models.V1Meta{
				Apiversion:  s.Meta.Apiversion,
				Kind:        s.Meta.Kind,
				Annotations: s.Meta.Annotations,
				Labels:      s.Meta.Labels,
			},
			Description: s.Description,
			Name:        s.Name,
			Quotas: &models.V1QuotaSet{
				Cluster: &models.V1Quota{
					Quota: s.Quotas.Cluster.Quota,
				},
				IP: &models.V1Quota{
					Quota: s.Quotas.IP.Quota,
				},
				Machine: &models.V1Quota{
					Quota: s.Quotas.Machine.Quota,
				},
			},
			TenantID: s.TenantID,
		}
	}
	toProjectCreateRequest = func(s *models.V1ProjectResponse) *models.V1ProjectCreateRequest {
		return &models.V1ProjectCreateRequest{
			Meta: &models.V1Meta{
				Apiversion:  s.Meta.Apiversion,
				Kind:        s.Meta.Kind,
				ID:          s.Meta.ID,
				Annotations: s.Meta.Annotations,
				Labels:      s.Meta.Labels,
				Version:     s.Meta.Version,
			},
			Description: s.Description,
			Name:        s.Name,
			Quotas:      s.Quotas,
			TenantID:    s.TenantID,
		}
	}
	toProjectUpdateRequest = func(s *models.V1ProjectResponse) *models.V1ProjectUpdateRequest {
		return &models.V1ProjectUpdateRequest{
			Meta: &models.V1Meta{
				Apiversion:  s.Meta.Apiversion,
				Kind:        s.Meta.Kind,
				ID:          s.Meta.ID,
				Annotations: s.Meta.Annotations,
				Labels:      s.Meta.Labels,
				Version:     s.Meta.Version,
			},
			Description: s.Description,
			Name:        s.Name,
			Quotas:      s.Quotas,
			TenantID:    s.TenantID,
		}
	}
)

func Test_ProjectCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1ProjectResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1ProjectResponse) []string {
				return []string{"project", "list"}
			},
			mocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("FindProjects", testcommon.MatchIgnoreContext(t, project.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{})), nil).Return(&project.FindProjectsOK{
						Payload: []*models.V1ProjectResponse{
							project2,
							project1,
						},
					}, nil)
				},
			},
			want: []*models.V1ProjectResponse{
				project1,
				project2,
			},
			wantTable: pointer.Pointer(`
UID   TENANT        NAME        DESCRIPTION   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     c        a=b
2     metal-stack   project-2   project 2     c        a=b
`),
			wantWideTable: pointer.Pointer(`
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     1/3/2                          c        a=b
2     metal-stack   project-2   project 2     ∞/∞/∞                          c        a=b
`),
			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 project-1
2 project-2
`),
			wantMarkdown: pointer.Pointer(`
| UID |   TENANT    |   NAME    | DESCRIPTION | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | c      | a=b         |
|   2 | metal-stack | project-2 | project 2   | c      | a=b         |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1ProjectResponse) []string {
				return []string{"project", "list", "--name", "project-1", "--tenant", "metal-stack", "--id", want[0].Meta.ID}
			},
			mocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("FindProjects", testcommon.MatchIgnoreContext(t, project.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{
						Name:     "project-1",
						TenantID: "metal-stack",
						ID:       "1",
					})), nil).Return(&project.FindProjectsOK{
						Payload: []*models.V1ProjectResponse{
							project1,
						},
					}, nil)
				},
			},
			want: []*models.V1ProjectResponse{
				project1,
			},
			wantTable: pointer.Pointer(`
UID   TENANT        NAME        DESCRIPTION   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     c        a=b
`),
			wantWideTable: pointer.Pointer(`
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     1/3/2                          c        a=b
`),
			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 project-1
`),
			wantMarkdown: pointer.Pointer(`
| UID |   TENANT    |   NAME    | DESCRIPTION | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | c      | a=b         |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1ProjectResponse) []string {
				return []string{"project", "apply", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1ProjectResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(toProjectCreateRequest(project1))), nil).Return(nil, &project.CreateProjectConflict{}).Once()
					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.FindProjectOK{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								Version: 0,
							},
						},
					}, nil)
					mock.On("UpdateProject", testcommon.MatchIgnoreContext(t, project.NewUpdateProjectParams().WithBody(toProjectUpdateRequest(project1))), nil).Return(&project.UpdateProjectOK{
						Payload: project1,
					}, nil)
					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(toProjectCreateRequest(project2))), nil).Return(&project.CreateProjectCreated{
						Payload: project2,
					}, nil)
				},
			},
			want: []*models.V1ProjectResponse{
				project1,
				project2,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_ProjectCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1ProjectResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1ProjectResponse) []string {
				return []string{"project", "describe", want.Meta.ID}
			},
			mocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.FindProjectOK{
						Payload: project1,
					}, nil)
				},
			},
			want: project1,
			wantTable: pointer.Pointer(`
UID   TENANT        NAME        DESCRIPTION   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     c        a=b
`),
			wantWideTable: pointer.Pointer(`
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     1/3/2                          c        a=b
`),
			template: pointer.Pointer("{{ .meta.id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 project-1
`),
			wantMarkdown: pointer.Pointer(`
| UID |   TENANT    |   NAME    | DESCRIPTION | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | c      | a=b         |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1ProjectResponse) []string {
				return []string{"project", "rm", want.Meta.ID}
			},
			mocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("DeleteProject", testcommon.MatchIgnoreContext(t, project.NewDeleteProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.DeleteProjectOK{
						Payload: project1,
					}, nil)
				},
			},
			want: project1,
		},
		{
			name: "create",
			cmd: func(want *models.V1ProjectResponse) []string {
				return []string{"project", "create",
					"--name", want.Name,
					"--description", want.Description,
					"--tenant", want.TenantID,
					"--label", strings.Join(want.Meta.Labels, ","),
					"--annotation", strings.Join(genericcli.MapToLabels(want.Meta.Annotations), ","),
					"--cluster-quota", strconv.FormatInt(int64(want.Quotas.Cluster.Quota), 10),
					"--machine-quota", strconv.FormatInt(int64(want.Quotas.Machine.Quota), 10),
					"--ip-quota", strconv.FormatInt(int64(want.Quotas.IP.Quota), 10),
				}
			},
			mocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(toProjectCreateRequestFromCLI(project1))), nil).Return(&project.CreateProjectCreated{
						Payload: project1,
					}, nil)
				},
			},
			want: project1,
		},
		{
			name: "create from file",
			cmd: func(want *models.V1ProjectResponse) []string {
				return []string{"project", "create", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1ProjectResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(toProjectCreateRequest(project1))), nil).Return(&project.CreateProjectCreated{
						Payload: project1,
					}, nil)
				},
			},
			want: project1,
		},
		{
			name: "update from file",
			cmd: func(want *models.V1ProjectResponse) []string {
				return []string{"project", "update", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1ProjectResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID(project1.Meta.ID)), nil).Return(&project.FindProjectOK{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								Version: 0,
							},
						},
					}, nil)
					mock.On("UpdateProject", testcommon.MatchIgnoreContext(t, project.NewUpdateProjectParams().WithBody(toProjectUpdateRequest(project1))), nil).Return(&project.UpdateProjectOK{
						Payload: project1,
					}, nil)
				},
			},
			want: project1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
