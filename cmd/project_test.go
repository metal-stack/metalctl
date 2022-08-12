package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/client/project"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_ProjectListCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		filterArgs   []string
		want         []*models.V1ProjectResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "list projects",
			metalMocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("FindProjects", testcommon.MatchIgnoreContext(t, project.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{})), nil).Return(&project.FindProjectsOK{
						Payload: []*models.V1ProjectResponse{
							{
								Meta: &models.V1Meta{
									ID: "2",
									Annotations: map[string]string{
										"a": "b",
									},
									Labels: []string{"c"},
								},
								Description: "project 2",
								Name:        "project-2",
								Quotas:      &models.V1QuotaSet{},
								TenantID:    "metal-stack",
							},
							{
								Meta: &models.V1Meta{
									ID: "1",
									Annotations: map[string]string{
										"a": "b",
									},
									Labels: []string{"c"},
								},
								Description: "project 1",
								Name:        "project-1",
								Quotas:      &models.V1QuotaSet{},
								TenantID:    "metal-stack",
							},
						},
					}, nil)
				},
			},
			want: []*models.V1ProjectResponse{
				{
					Meta: &models.V1Meta{
						ID: "1",
						Annotations: map[string]string{
							"a": "b",
						},
						Labels: []string{"c"},
					},
					Description: "project 1",
					Name:        "project-1",
					Quotas:      &models.V1QuotaSet{},
					TenantID:    "metal-stack",
				},
				{
					Meta: &models.V1Meta{
						ID: "2",
						Annotations: map[string]string{
							"a": "b",
						},
						Labels: []string{"c"},
					},
					Description: "project 2",
					Name:        "project-2",
					Quotas:      &models.V1QuotaSet{},
					TenantID:    "metal-stack",
				},
			},
			wantTable: `
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     ∞/∞/∞                          c        a=b
2     metal-stack   project-2   project 2     ∞/∞/∞                          c        a=b
`,
			template: "{{ .meta.id }} {{ .name }}",
			wantTemplate: `
1 project-1
2 project-2
`,
			wantMarkdown: `
| UID |   TENANT    |   NAME    | DESCRIPTION | QUOTAS CLUSTERS/MACHINES/IPS | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|------------------------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | ∞/∞/∞                        | c      | a=b         |
|   2 | metal-stack | project-2 | project 2   | ∞/∞/∞                        | c      | a=b         |
`,
		},
		{
			name:       "list projects with filters",
			filterArgs: []string{"--name", "project-1", "--tenant", "metal-stack", "--id", "1"},
			metalMocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("FindProjects", testcommon.MatchIgnoreContext(t, project.NewFindProjectsParams().WithBody(&models.V1ProjectFindRequest{
						Name:     "project-1",
						TenantID: "metal-stack",
						ID:       "1",
					})), nil).Return(&project.FindProjectsOK{
						Payload: []*models.V1ProjectResponse{
							{
								Meta: &models.V1Meta{
									ID: "1",
									Annotations: map[string]string{
										"a": "b",
									},
									Labels: []string{"c"},
								},
								Description: "project 1",
								Name:        "project-1",
								Quotas:      &models.V1QuotaSet{},
								TenantID:    "metal-stack",
							},
						},
					}, nil)
				},
			},
			want: []*models.V1ProjectResponse{
				{
					Meta: &models.V1Meta{
						ID: "1",
						Annotations: map[string]string{
							"a": "b",
						},
						Labels: []string{"c"},
					},
					Description: "project 1",
					Name:        "project-1",
					Quotas:      &models.V1QuotaSet{},
					TenantID:    "metal-stack",
				},
			},
			wantTable: `
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     ∞/∞/∞                          c        a=b
`,
			template: "{{ .meta.id }} {{ .name }}",
			wantTemplate: `
1 project-1
`,
			wantMarkdown: `
| UID |   TENANT    |   NAME    | DESCRIPTION | QUOTAS CLUSTERS/MACHINES/IPS | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|------------------------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | ∞/∞/∞                        | c      | a=b         |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[[]*models.V1ProjectResponse]{
			want:           tt.want,
			table:          pointer.Pointer(tt.wantTable),
			template:       pointer.Pointer(tt.template),
			templateOutput: pointer.Pointer(tt.wantTemplate),
			markdownTable:  pointer.Pointer(tt.wantMarkdown),
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, format := range outputFormats(formats) {
				format := format
				t.Run(fmt.Sprintf("%v", format.Args()), func(t *testing.T) {
					var out bytes.Buffer
					config, mock := newTestConfig(t, &out, tt.metalMocks, nil)

					cmd := newRootCmd(config)
					os.Args = append([]string{binaryName, "project", "list"}, format.Args()...)
					os.Args = append(os.Args, tt.filterArgs...)

					err := cmd.Execute()
					if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
						t.Errorf("error diff (+got -want):\n %s", diff)
					}

					format.Validate(t, out.Bytes())

					mock.AssertExpectations(t)
				})
			}
		})
	}
}

func Test_ProjectDescribeCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ProjectResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "list projects",
			metalMocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID("1")), nil).Return(&project.FindProjectOK{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								ID: "1",
								Annotations: map[string]string{
									"a": "b",
								},
								Labels: []string{"c"},
							},
							Description: "project 1",
							Name:        "project-1",
							Quotas:      &models.V1QuotaSet{},
							TenantID:    "metal-stack",
						},
					}, nil)
				},
			},
			want: &models.V1ProjectResponse{
				Meta: &models.V1Meta{
					ID: "1",
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: []string{"c"},
				},
				Description: "project 1",
				Name:        "project-1",
				Quotas:      &models.V1QuotaSet{},
				TenantID:    "metal-stack",
			},
			wantTable: `
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     ∞/∞/∞                          c        a=b
`,
			template: "{{ .meta.id }} {{ .name }}",
			wantTemplate: `
1 project-1
`,
			wantMarkdown: `
| UID |   TENANT    |   NAME    | DESCRIPTION | QUOTAS CLUSTERS/MACHINES/IPS | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|------------------------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | ∞/∞/∞                        | c      | a=b         |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ProjectResponse]{
			want:           tt.want,
			table:          pointer.Pointer(tt.wantTable),
			template:       pointer.Pointer(tt.template),
			templateOutput: pointer.Pointer(tt.wantTemplate),
			markdownTable:  pointer.Pointer(tt.wantMarkdown),
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, format := range outputFormats(formats) {
				format := format
				t.Run(fmt.Sprintf("%v", format.Args()), func(t *testing.T) {
					var out bytes.Buffer
					config, mock := newTestConfig(t, &out, tt.metalMocks, nil)

					cmd := newRootCmd(config)
					os.Args = append([]string{binaryName, "project", "describe", tt.want.Meta.ID}, format.Args()...)

					err := cmd.Execute()
					if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
						t.Errorf("error diff (+got -want):\n %s", diff)
					}

					format.Validate(t, out.Bytes())

					mock.AssertExpectations(t)
				})
			}
		})
	}
}

func Test_ProjectDeleteCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ProjectResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "delete projects",
			metalMocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("DeleteProject", testcommon.MatchIgnoreContext(t, project.NewDeleteProjectParams().WithID("1")), nil).Return(&project.DeleteProjectOK{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								ID: "1",
								Annotations: map[string]string{
									"a": "b",
								},
								Labels: []string{"c"},
							},
							Description: "project 1",
							Name:        "project-1",
							Quotas:      &models.V1QuotaSet{},
							TenantID:    "metal-stack",
						},
					}, nil)
				},
			},
			want: &models.V1ProjectResponse{
				Meta: &models.V1Meta{
					ID: "1",
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: []string{"c"},
				},
				Description: "project 1",
				Name:        "project-1",
				Quotas:      &models.V1QuotaSet{},
				TenantID:    "metal-stack",
			},
			wantTable: `
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     ∞/∞/∞                          c        a=b
`,
			template: "{{ .meta.id }} {{ .name }}",
			wantTemplate: `
1 project-1
`,
			wantMarkdown: `
| UID |   TENANT    |   NAME    | DESCRIPTION | QUOTAS CLUSTERS/MACHINES/IPS | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|------------------------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | ∞/∞/∞                        | c      | a=b         |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ProjectResponse]{
			want:           tt.want,
			table:          pointer.Pointer(tt.wantTable),
			template:       pointer.Pointer(tt.template),
			templateOutput: pointer.Pointer(tt.wantTemplate),
			markdownTable:  pointer.Pointer(tt.wantMarkdown),
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, format := range outputFormats(formats) {
				format := format
				t.Run(fmt.Sprintf("%v", format.Args()), func(t *testing.T) {
					var out bytes.Buffer
					config, mock := newTestConfig(t, &out, tt.metalMocks, nil)

					cmd := newRootCmd(config)
					os.Args = append([]string{binaryName, "project", "rm", tt.want.Meta.ID}, format.Args()...)

					err := cmd.Execute()
					if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
						t.Errorf("error diff (+got -want):\n %s", diff)
					}

					format.Validate(t, out.Bytes())

					mock.AssertExpectations(t)
				})
			}
		})
	}
}

func Test_ProjectCreateFromCLICmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ProjectResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "create projects",
			metalMocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(&models.V1ProjectCreateRequest{
						Description: "project 1",
						Meta: &models.V1Meta{
							Apiversion: "v1",
							Kind:       "Project",
							Annotations: map[string]string{
								"a": "b",
							},
							Labels: []string{"c"},
						},
						Name: "project-1",
						Quotas: &models.V1QuotaSet{
							Cluster: &models.V1Quota{Quota: 1},
							Machine: &models.V1Quota{Quota: 2},
							IP:      &models.V1Quota{Quota: 3},
						},
						TenantID: "metal-stack",
					})), nil).Return(&project.CreateProjectCreated{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								ID: "1",
								Annotations: map[string]string{
									"a": "b",
								},
								Labels: []string{"c"},
							},
							Description: "project 1",
							Name:        "project-1",
							Quotas: &models.V1QuotaSet{
								Cluster: &models.V1Quota{Quota: 1},
								Machine: &models.V1Quota{Quota: 2},
								IP:      &models.V1Quota{Quota: 3},
							},
							TenantID: "metal-stack",
						},
					}, nil)
				},
			},
			want: &models.V1ProjectResponse{
				Meta: &models.V1Meta{
					ID: "1",
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: []string{"c"},
				},
				Description: "project 1",
				Name:        "project-1",
				Quotas: &models.V1QuotaSet{
					Cluster: &models.V1Quota{Quota: 1},
					Machine: &models.V1Quota{Quota: 2},
					IP:      &models.V1Quota{Quota: 3},
				},
				TenantID: "metal-stack",
			},
			wantTable: `
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     1/2/3                          c        a=b
`,
			template: "{{ .meta.id }} {{ .name }}",
			wantTemplate: `
1 project-1
`,
			wantMarkdown: `
| UID |   TENANT    |   NAME    | DESCRIPTION | QUOTAS CLUSTERS/MACHINES/IPS | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|------------------------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | 1/2/3                        | c      | a=b         |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ProjectResponse]{
			want:           tt.want,
			table:          pointer.Pointer(tt.wantTable),
			template:       pointer.Pointer(tt.template),
			templateOutput: pointer.Pointer(tt.wantTemplate),
			markdownTable:  pointer.Pointer(tt.wantMarkdown),
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, format := range outputFormats(formats) {
				format := format
				t.Run(fmt.Sprintf("%v", format.Args()), func(t *testing.T) {
					var out bytes.Buffer
					config, mock := newTestConfig(t, &out, tt.metalMocks, nil)

					cmd := newRootCmd(config)
					os.Args = append([]string{binaryName, "project", "create",
						"--name", tt.want.Name,
						"--description", tt.want.Description,
						"--tenant", tt.want.TenantID,
						"--label", "c",
						"--annotation", "a=b",
						"--cluster-quota", "1",
						"--machine-quota", "2",
						"--ip-quota", "3",
					}, format.Args()...)

					err := cmd.Execute()
					if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
						t.Errorf("error diff (+got -want):\n %s", diff)
					}

					format.Validate(t, out.Bytes())

					mock.AssertExpectations(t)
				})
			}
		})
	}
}

func Test_ProjectCreateFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ProjectResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "create projects",
			metalMocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(&models.V1ProjectCreateRequest{
						Meta: &models.V1Meta{
							Apiversion: "v1",
							Kind:       "Project",
							ID:         "1",
							Annotations: map[string]string{
								"a": "b",
							},
							Labels: []string{"c"},
						},
						Name:        "project-1",
						Description: "project 1",
						Quotas: &models.V1QuotaSet{
							Cluster: &models.V1Quota{Quota: 1},
							Machine: &models.V1Quota{Quota: 2},
							IP:      &models.V1Quota{Quota: 3},
						},
						TenantID: "metal-stack",
					})), nil).Return(&project.CreateProjectCreated{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								Apiversion: "v1",
								Kind:       "Project",
								ID:         "1",
								Annotations: map[string]string{
									"a": "b",
								},
								Labels: []string{"c"},
							},
							Description: "project 1",
							Name:        "project-1",
							Quotas: &models.V1QuotaSet{
								Cluster: &models.V1Quota{Quota: 1},
								Machine: &models.V1Quota{Quota: 2},
								IP:      &models.V1Quota{Quota: 3},
							},
							TenantID: "metal-stack",
						},
					}, nil)
				},
			},
			want: &models.V1ProjectResponse{
				Meta: &models.V1Meta{
					Apiversion: "v1",
					Kind:       "Project",
					ID:         "1",
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: []string{"c"},
				},
				Description: "project 1",
				Name:        "project-1",
				Quotas: &models.V1QuotaSet{
					Cluster: &models.V1Quota{Quota: 1},
					Machine: &models.V1Quota{Quota: 2},
					IP:      &models.V1Quota{Quota: 3},
				},
				TenantID: "metal-stack",
			},
			wantTable: `
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     1/2/3                          c        a=b
`,
			template: "{{ .meta.id }} {{ .name }}",
			wantTemplate: `
1 project-1
`,
			wantMarkdown: `
| UID |   TENANT    |   NAME    | DESCRIPTION | QUOTAS CLUSTERS/MACHINES/IPS | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|------------------------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | 1/2/3                        | c      | a=b         |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ProjectResponse]{
			want:           tt.want,
			table:          pointer.Pointer(tt.wantTable),
			template:       pointer.Pointer(tt.template),
			templateOutput: pointer.Pointer(tt.wantTemplate),
			markdownTable:  pointer.Pointer(tt.wantMarkdown),
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, format := range outputFormats(formats) {
				format := format
				t.Run(fmt.Sprintf("%v", format.Args()), func(t *testing.T) {
					var out bytes.Buffer
					config, mock := newTestConfig(t, &out, tt.metalMocks, func(fs afero.Fs) {
						require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, tt.want), 0755))
					})

					cmd := newRootCmd(config)
					os.Args = append([]string{binaryName, "project", "create", "-f", "/file.yaml"}, format.Args()...)

					err := cmd.Execute()
					if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
						t.Errorf("error diff (+got -want):\n %s", diff)
					}

					format.Validate(t, out.Bytes())

					mock.AssertExpectations(t)
				})
			}
		})
	}
}

func Test_ProjectUpdateFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ProjectResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "update projects",
			metalMocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID("1")), nil).Return(&project.FindProjectOK{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								Version: 0,
							},
						},
					}, nil)
					mock.On("UpdateProject", testcommon.MatchIgnoreContext(t, project.NewUpdateProjectParams().WithBody(&models.V1ProjectUpdateRequest{
						Meta: &models.V1Meta{
							Apiversion: "v1",
							Kind:       "Project",
							ID:         "1",
							Annotations: map[string]string{
								"a": "b",
							},
							Labels:  []string{"c"},
							Version: 1,
						},
						Name:        "project-1",
						Description: "project 1",
						Quotas: &models.V1QuotaSet{
							Cluster: &models.V1Quota{Quota: 1},
							Machine: &models.V1Quota{Quota: 2},
							IP:      &models.V1Quota{Quota: 3},
						},
						TenantID: "metal-stack",
					})), nil).Return(&project.UpdateProjectOK{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								Apiversion: "v1",
								Kind:       "Project",
								ID:         "1",
								Annotations: map[string]string{
									"a": "b",
								},
								Labels: []string{"c"},
							},
							Description: "project 1",
							Name:        "project-1",
							Quotas: &models.V1QuotaSet{
								Cluster: &models.V1Quota{Quota: 1},
								Machine: &models.V1Quota{Quota: 2},
								IP:      &models.V1Quota{Quota: 3},
							},
							TenantID: "metal-stack",
						},
					}, nil)
				},
			},
			want: &models.V1ProjectResponse{
				Meta: &models.V1Meta{
					Apiversion: "v1",
					Kind:       "Project",
					ID:         "1",
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: []string{"c"},
				},
				Description: "project 1",
				Name:        "project-1",
				Quotas: &models.V1QuotaSet{
					Cluster: &models.V1Quota{Quota: 1},
					Machine: &models.V1Quota{Quota: 2},
					IP:      &models.V1Quota{Quota: 3},
				},
				TenantID: "metal-stack",
			},
			wantTable: `
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     1/2/3                          c        a=b
`,
			template: "{{ .meta.id }} {{ .name }}",
			wantTemplate: `
1 project-1
`,
			wantMarkdown: `
| UID |   TENANT    |   NAME    | DESCRIPTION | QUOTAS CLUSTERS/MACHINES/IPS | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|------------------------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | 1/2/3                        | c      | a=b         |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ProjectResponse]{
			want:           tt.want,
			table:          pointer.Pointer(tt.wantTable),
			template:       pointer.Pointer(tt.template),
			templateOutput: pointer.Pointer(tt.wantTemplate),
			markdownTable:  pointer.Pointer(tt.wantMarkdown),
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, format := range outputFormats(formats) {
				format := format
				t.Run(fmt.Sprintf("%v", format.Args()), func(t *testing.T) {
					var out bytes.Buffer
					config, mock := newTestConfig(t, &out, tt.metalMocks, func(fs afero.Fs) {
						require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, tt.want), 0755))
					})

					cmd := newRootCmd(config)
					os.Args = append([]string{binaryName, "project", "update", "-f", "/file.yaml"}, format.Args()...)

					err := cmd.Execute()
					if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
						t.Errorf("error diff (+got -want):\n %s", diff)
					}

					format.Validate(t, out.Bytes())

					mock.AssertExpectations(t)
				})
			}
		})
	}
}

func Test_ProjectApplyFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         []*models.V1ProjectResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "apply projects from file",
			metalMocks: &client.MetalMockFns{
				Project: func(mock *mock.Mock) {
					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(&models.V1ProjectCreateRequest{
						Meta: &models.V1Meta{
							Apiversion: "v1",
							Kind:       "Project",
							ID:         "2",
							Annotations: map[string]string{
								"a": "b",
							},
							Labels: []string{"c"},
						},
						Name:        "project-2",
						Description: "project 2",
						Quotas: &models.V1QuotaSet{
							Cluster: &models.V1Quota{Quota: 1},
							Machine: &models.V1Quota{Quota: 2},
							IP:      &models.V1Quota{Quota: 3},
						},
						TenantID: "metal-stack",
					})), nil).Return(nil, &project.CreateProjectConflict{}).Once()
					mock.On("FindProject", testcommon.MatchIgnoreContext(t, project.NewFindProjectParams().WithID("2")), nil).Return(&project.FindProjectOK{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								Version: 0,
							},
						},
					}, nil)
					mock.On("UpdateProject", testcommon.MatchIgnoreContext(t, project.NewUpdateProjectParams().WithBody(&models.V1ProjectUpdateRequest{
						Meta: &models.V1Meta{
							Apiversion: "v1",
							Kind:       "Project",
							ID:         "2",
							Annotations: map[string]string{
								"a": "b",
							},
							Labels:  []string{"c"},
							Version: 1,
						},
						Name:        "project-2",
						Description: "project 2",
						Quotas: &models.V1QuotaSet{
							Cluster: &models.V1Quota{Quota: 1},
							Machine: &models.V1Quota{Quota: 2},
							IP:      &models.V1Quota{Quota: 3},
						},
						TenantID: "metal-stack",
					})), nil).Return(&project.UpdateProjectOK{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								Apiversion: "v1",
								Kind:       "Project",
								ID:         "2",
								Annotations: map[string]string{
									"a": "b",
								},
								Labels:  []string{"c"},
								Version: 0, // otherwise does not work for apply file generation from want
							},
							Name:        "project-2",
							Description: "project 2",
							Quotas: &models.V1QuotaSet{
								Cluster: &models.V1Quota{Quota: 1},
								Machine: &models.V1Quota{Quota: 2},
								IP:      &models.V1Quota{Quota: 3},
							},
							TenantID: "metal-stack",
						},
					}, nil)
					mock.On("CreateProject", testcommon.MatchIgnoreContext(t, project.NewCreateProjectParams().WithBody(&models.V1ProjectCreateRequest{
						Meta: &models.V1Meta{
							Apiversion: "v1",
							Kind:       "Project",
							ID:         "1",
							Annotations: map[string]string{
								"a": "b",
							},
							Labels: []string{"c"},
						},
						Name:        "project-1",
						Description: "project 1",
						Quotas: &models.V1QuotaSet{
							Cluster: &models.V1Quota{Quota: 1},
							Machine: &models.V1Quota{Quota: 2},
							IP:      &models.V1Quota{Quota: 3},
						},
						TenantID: "metal-stack",
					})), nil).Return(&project.CreateProjectCreated{
						Payload: &models.V1ProjectResponse{
							Meta: &models.V1Meta{
								Apiversion: "v1",
								Kind:       "Project",
								ID:         "1",
								Annotations: map[string]string{
									"a": "b",
								},
								Labels: []string{"c"},
							},
							Name:        "project-1",
							Description: "project 1",
							Quotas: &models.V1QuotaSet{
								Cluster: &models.V1Quota{Quota: 1},
								Machine: &models.V1Quota{Quota: 2},
								IP:      &models.V1Quota{Quota: 3},
							},
							TenantID: "metal-stack",
						},
					}, nil)
				},
			},
			want: []*models.V1ProjectResponse{
				{
					Meta: &models.V1Meta{
						Apiversion: "v1",
						Kind:       "Project",
						ID:         "1",
						Annotations: map[string]string{
							"a": "b",
						},
						Labels: []string{"c"},
					},
					Name:        "project-1",
					Description: "project 1",
					Quotas: &models.V1QuotaSet{
						Cluster: &models.V1Quota{Quota: 1},
						Machine: &models.V1Quota{Quota: 2},
						IP:      &models.V1Quota{Quota: 3},
					},
					TenantID: "metal-stack",
				},
				{
					Meta: &models.V1Meta{
						Apiversion: "v1",
						Kind:       "Project",
						ID:         "2",
						Annotations: map[string]string{
							"a": "b",
						},
						Labels: []string{"c"},
					},
					Name:        "project-2",
					Description: "project 2",
					Quotas: &models.V1QuotaSet{
						Cluster: &models.V1Quota{Quota: 1},
						Machine: &models.V1Quota{Quota: 2},
						IP:      &models.V1Quota{Quota: 3},
					},
					TenantID: "metal-stack",
				},
			},
			wantTable: `
UID   TENANT        NAME        DESCRIPTION   QUOTAS CLUSTERS/MACHINES/IPS   LABELS   ANNOTATIONS
1     metal-stack   project-1   project 1     1/2/3                          c        a=b
2     metal-stack   project-2   project 2     1/2/3                          c        a=b
`,
			template: "{{ .meta.id }} {{ .name }}",
			wantTemplate: `
1 project-1
2 project-2
`,
			wantMarkdown: `
| UID |   TENANT    |   NAME    | DESCRIPTION | QUOTAS CLUSTERS/MACHINES/IPS | LABELS | ANNOTATIONS |
|-----|-------------|-----------|-------------|------------------------------|--------|-------------|
|   1 | metal-stack | project-1 | project 1   | 1/2/3                        | c      | a=b         |
|   2 | metal-stack | project-2 | project 2   | 1/2/3                        | c      | a=b         |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[[]*models.V1ProjectResponse]{
			want:           tt.want,
			table:          pointer.Pointer(tt.wantTable),
			template:       pointer.Pointer(tt.template),
			templateOutput: pointer.Pointer(tt.wantTemplate),
			markdownTable:  pointer.Pointer(tt.wantMarkdown),
		}
		t.Run(tt.name, func(t *testing.T) {
			for _, format := range outputFormats(formats) {
				format := format
				t.Run(fmt.Sprintf("%v", format.Args()), func(t *testing.T) {
					var out bytes.Buffer
					config, mock := newTestConfig(t, &out, tt.metalMocks, func(fs afero.Fs) {
						var parts []string
						for _, elem := range tt.want {
							parts = append(parts, string(mustMarshal(t, elem)))
						}
						content := strings.Join(parts, "\n---\n")
						require.NoError(t, afero.WriteFile(fs, "/file.yaml", []byte(content), 0755))
					})

					cmd := newRootCmd(config)
					os.Args = append([]string{binaryName, "project", "apply", "-f", "/file.yaml"}, format.Args()...)

					err := cmd.Execute()
					if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
						t.Errorf("error diff (+got -want):\n %s", diff)
					}

					format.Validate(t, out.Bytes())

					mock.AssertExpectations(t)
				})
			}
		})
	}
}
