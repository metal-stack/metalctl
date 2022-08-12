package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/client/partition"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_PartitionListCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         []*models.V1PartitionResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "list partitions",
			metalMocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("ListPartitions", testcommon.MatchIgnoreContext(t, partition.NewListPartitionsParams()), nil).Return(&partition.ListPartitionsOK{
						Payload: []*models.V1PartitionResponse{
							{
								Bootconfig: &models.V1PartitionBootConfiguration{
									Commandline: "commandline",
									Imageurl:    "imageurl",
									Kernelurl:   "kernelurl",
								},
								Description:                "partition 1",
								ID:                         pointer.Pointer("1"),
								Mgmtserviceaddress:         "mgmt",
								Name:                       "partition-1",
								Privatenetworkprefixlength: 24,
							},
							{
								Bootconfig: &models.V1PartitionBootConfiguration{
									Commandline: "commandline",
									Imageurl:    "imageurl",
									Kernelurl:   "kernelurl",
								},
								Description:                "partition 2",
								ID:                         pointer.Pointer("2"),
								Mgmtserviceaddress:         "mgmt",
								Name:                       "partition-2",
								Privatenetworkprefixlength: 24,
							},
						},
					}, nil)
				},
			},
			want: []*models.V1PartitionResponse{
				{
					Bootconfig: &models.V1PartitionBootConfiguration{
						Commandline: "commandline",
						Imageurl:    "imageurl",
						Kernelurl:   "kernelurl",
					},
					Description:                "partition 1",
					ID:                         pointer.Pointer("1"),
					Mgmtserviceaddress:         "mgmt",
					Name:                       "partition-1",
					Privatenetworkprefixlength: 24,
				},
				{
					Bootconfig: &models.V1PartitionBootConfiguration{
						Commandline: "commandline",
						Imageurl:    "imageurl",
						Kernelurl:   "kernelurl",
					},
					Description:                "partition 2",
					ID:                         pointer.Pointer("2"),
					Mgmtserviceaddress:         "mgmt",
					Name:                       "partition-2",
					Privatenetworkprefixlength: 24,
				},
			},
			wantTable: `
ID   NAME          DESCRIPTION
1    partition-1   partition 1
2    partition-2   partition 2
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
1 partition-1
2 partition-2
`,
			wantMarkdown: `
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
|  2 | partition-2 | partition 2 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[[]*models.V1PartitionResponse]{
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
					os.Args = append([]string{binaryName, "partition", "list"}, format.Args()...)

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

func Test_PartitionDescribeCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1PartitionResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "describe partition",
			metalMocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("FindPartition", testcommon.MatchIgnoreContext(t, partition.NewFindPartitionParams().WithID("1")), nil).Return(&partition.FindPartitionOK{
						Payload: &models.V1PartitionResponse{
							Bootconfig: &models.V1PartitionBootConfiguration{
								Commandline: "commandline",
								Imageurl:    "imageurl",
								Kernelurl:   "kernelurl",
							},
							Description:                "partition 1",
							ID:                         pointer.Pointer("1"),
							Mgmtserviceaddress:         "mgmt",
							Name:                       "partition-1",
							Privatenetworkprefixlength: 24,
						},
					}, nil)
				},
			},
			want: &models.V1PartitionResponse{
				Bootconfig: &models.V1PartitionBootConfiguration{
					Commandline: "commandline",
					Imageurl:    "imageurl",
					Kernelurl:   "kernelurl",
				},
				Description:                "partition 1",
				ID:                         pointer.Pointer("1"),
				Mgmtserviceaddress:         "mgmt",
				Name:                       "partition-1",
				Privatenetworkprefixlength: 24,
			},
			wantTable: `
ID   NAME          DESCRIPTION
1    partition-1   partition 1
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
1 partition-1
`,
			wantMarkdown: `
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1PartitionResponse]{
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
					os.Args = append([]string{binaryName, "partition", "describe", *tt.want.ID}, format.Args()...)

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

func Test_PartitionDeleteCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1PartitionResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "remove partition",
			metalMocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("DeletePartition", testcommon.MatchIgnoreContext(t, partition.NewDeletePartitionParams().WithID("1")), nil).Return(&partition.DeletePartitionOK{
						Payload: &models.V1PartitionResponse{
							Bootconfig: &models.V1PartitionBootConfiguration{
								Commandline: "commandline",
								Imageurl:    "imageurl",
								Kernelurl:   "kernelurl",
							},
							Description:                "partition 1",
							ID:                         pointer.Pointer("1"),
							Mgmtserviceaddress:         "mgmt",
							Name:                       "partition-1",
							Privatenetworkprefixlength: 24,
						},
					}, nil)
				},
			},
			want: &models.V1PartitionResponse{
				Bootconfig: &models.V1PartitionBootConfiguration{
					Commandline: "commandline",
					Imageurl:    "imageurl",
					Kernelurl:   "kernelurl",
				},
				Description:                "partition 1",
				ID:                         pointer.Pointer("1"),
				Mgmtserviceaddress:         "mgmt",
				Name:                       "partition-1",
				Privatenetworkprefixlength: 24,
			},
			wantTable: `
ID   NAME          DESCRIPTION
1    partition-1   partition 1
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
1 partition-1
`,
			wantMarkdown: `
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1PartitionResponse]{
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
					os.Args = append([]string{binaryName, "partition", "rm", *tt.want.ID}, format.Args()...)

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

func Test_PartitionCreateFromCLICmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1PartitionResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "create partition",
			metalMocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("CreatePartition", testcommon.MatchIgnoreContext(t, partition.NewCreatePartitionParams().WithBody(&models.V1PartitionCreateRequest{
						Bootconfig: &models.V1PartitionBootConfiguration{
							Commandline: "commandline",
							Imageurl:    "imageurl",
							Kernelurl:   "kernelurl",
						},
						Description:        "partition 1",
						ID:                 pointer.Pointer("1"),
						Mgmtserviceaddress: "mgmt",
						Name:               "partition-1",
					})), nil).Return(&partition.CreatePartitionCreated{
						Payload: &models.V1PartitionResponse{
							Bootconfig: &models.V1PartitionBootConfiguration{
								Commandline: "commandline",
								Imageurl:    "imageurl",
								Kernelurl:   "kernelurl",
							},
							Description:        "partition 1",
							ID:                 pointer.Pointer("1"),
							Mgmtserviceaddress: "mgmt",
							Name:               "partition-1",
						},
					}, nil)
				},
			},
			want: &models.V1PartitionResponse{
				Bootconfig: &models.V1PartitionBootConfiguration{
					Commandline: "commandline",
					Imageurl:    "imageurl",
					Kernelurl:   "kernelurl",
				},
				Description:        "partition 1",
				ID:                 pointer.Pointer("1"),
				Mgmtserviceaddress: "mgmt",
				Name:               "partition-1",
			},
			wantTable: `
ID   NAME          DESCRIPTION
1    partition-1   partition 1
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
1 partition-1
`,
			wantMarkdown: `
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1PartitionResponse]{
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
					os.Args = append([]string{binaryName, "partition", "create",
						"--id", *tt.want.ID,
						"--name", tt.want.Name,
						"--description", tt.want.Description,
						"--cmdline", tt.want.Bootconfig.Commandline,
						"--kernelurl", tt.want.Bootconfig.Kernelurl,
						"--imageurl", tt.want.Bootconfig.Imageurl,
						"--mgmtserver", tt.want.Mgmtserviceaddress,
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

func Test_PartitionCreateFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1PartitionResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "create partition",
			metalMocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("CreatePartition", testcommon.MatchIgnoreContext(t, partition.NewCreatePartitionParams().WithBody(&models.V1PartitionCreateRequest{
						Bootconfig: &models.V1PartitionBootConfiguration{
							Commandline: "commandline",
							Imageurl:    "imageurl",
							Kernelurl:   "kernelurl",
						},
						Description:                "partition 1",
						ID:                         pointer.Pointer("1"),
						Mgmtserviceaddress:         "mgmt",
						Name:                       "partition-1",
						Privatenetworkprefixlength: 24,
					})), nil).Return(&partition.CreatePartitionCreated{
						Payload: &models.V1PartitionResponse{
							Bootconfig: &models.V1PartitionBootConfiguration{
								Commandline: "commandline",
								Imageurl:    "imageurl",
								Kernelurl:   "kernelurl",
							},
							Description:                "partition 1",
							ID:                         pointer.Pointer("1"),
							Mgmtserviceaddress:         "mgmt",
							Name:                       "partition-1",
							Privatenetworkprefixlength: 24,
						},
					}, nil)
				},
			},
			want: &models.V1PartitionResponse{
				Bootconfig: &models.V1PartitionBootConfiguration{
					Commandline: "commandline",
					Imageurl:    "imageurl",
					Kernelurl:   "kernelurl",
				},
				Description:                "partition 1",
				ID:                         pointer.Pointer("1"),
				Mgmtserviceaddress:         "mgmt",
				Name:                       "partition-1",
				Privatenetworkprefixlength: 24,
			},
			wantTable: `
ID   NAME          DESCRIPTION
1    partition-1   partition 1
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
1 partition-1
`,
			wantMarkdown: `
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1PartitionResponse]{
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
					os.Args = append([]string{binaryName, "partition", "create", "-f", "/file.yaml"}, format.Args()...)

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

func Test_PartitionUpdateFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1PartitionResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "update partition from file",
			metalMocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("UpdatePartition", testcommon.MatchIgnoreContext(t, partition.NewUpdatePartitionParams().WithBody(&models.V1PartitionUpdateRequest{
						Bootconfig: &models.V1PartitionBootConfiguration{
							Commandline: "commandline",
							Imageurl:    "imageurl",
							Kernelurl:   "kernelurl",
						},
						Description:        "partition 1",
						ID:                 pointer.Pointer("1"),
						Mgmtserviceaddress: "mgmt",
						Name:               "partition-1",
					})), nil).Return(&partition.UpdatePartitionOK{
						Payload: &models.V1PartitionResponse{
							Bootconfig: &models.V1PartitionBootConfiguration{
								Commandline: "commandline",
								Imageurl:    "imageurl",
								Kernelurl:   "kernelurl",
							},
							Description:                "partition 1",
							ID:                         pointer.Pointer("1"),
							Mgmtserviceaddress:         "mgmt",
							Name:                       "partition-1",
							Privatenetworkprefixlength: 24,
						},
					}, nil)
				},
			},
			want: &models.V1PartitionResponse{
				Bootconfig: &models.V1PartitionBootConfiguration{
					Commandline: "commandline",
					Imageurl:    "imageurl",
					Kernelurl:   "kernelurl",
				},
				Description:                "partition 1",
				ID:                         pointer.Pointer("1"),
				Mgmtserviceaddress:         "mgmt",
				Name:                       "partition-1",
				Privatenetworkprefixlength: 24,
			},
			wantTable: `
ID   NAME          DESCRIPTION
1    partition-1   partition 1
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
1 partition-1
`,
			wantMarkdown: `
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1PartitionResponse]{
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
					os.Args = append([]string{binaryName, "partition", "update", "-f", "/file.yaml"}, format.Args()...)

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

func Test_PartitionApplyFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         []*models.V1PartitionResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "apply partitions from file",
			metalMocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("CreatePartition", testcommon.MatchIgnoreContext(t, partition.NewCreatePartitionParams().WithBody(&models.V1PartitionCreateRequest{
						Bootconfig: &models.V1PartitionBootConfiguration{
							Commandline: "commandline",
							Imageurl:    "imageurl",
							Kernelurl:   "kernelurl",
						},
						Description:                "partition 1",
						ID:                         pointer.Pointer("1"),
						Mgmtserviceaddress:         "mgmt",
						Name:                       "partition-1",
						Privatenetworkprefixlength: 24,
					})), nil).Return(nil, &partition.CreatePartitionConflict{}).Once()
					mock.On("UpdatePartition", testcommon.MatchIgnoreContext(t, partition.NewUpdatePartitionParams().WithBody(&models.V1PartitionUpdateRequest{
						Bootconfig: &models.V1PartitionBootConfiguration{
							Commandline: "commandline",
							Imageurl:    "imageurl",
							Kernelurl:   "kernelurl",
						},
						Description:        "partition 1",
						ID:                 pointer.Pointer("1"),
						Mgmtserviceaddress: "mgmt",
						Name:               "partition-1",
					})), nil).Return(&partition.UpdatePartitionOK{
						Payload: &models.V1PartitionResponse{
							Bootconfig: &models.V1PartitionBootConfiguration{
								Commandline: "commandline",
								Imageurl:    "imageurl",
								Kernelurl:   "kernelurl",
							},
							Description:                "partition 1",
							ID:                         pointer.Pointer("1"),
							Mgmtserviceaddress:         "mgmt",
							Name:                       "partition-1",
							Privatenetworkprefixlength: 24,
						},
					}, nil)
					mock.On("CreatePartition", testcommon.MatchIgnoreContext(t, partition.NewCreatePartitionParams().WithBody(&models.V1PartitionCreateRequest{
						Bootconfig: &models.V1PartitionBootConfiguration{
							Commandline: "commandline",
							Imageurl:    "imageurl",
							Kernelurl:   "kernelurl",
						},
						Description:                "partition 2",
						ID:                         pointer.Pointer("2"),
						Mgmtserviceaddress:         "mgmt",
						Name:                       "partition-2",
						Privatenetworkprefixlength: 24,
					})), nil).Return(&partition.CreatePartitionCreated{
						Payload: &models.V1PartitionResponse{
							Bootconfig: &models.V1PartitionBootConfiguration{
								Commandline: "commandline",
								Imageurl:    "imageurl",
								Kernelurl:   "kernelurl",
							},
							Description:                "partition 2",
							ID:                         pointer.Pointer("2"),
							Mgmtserviceaddress:         "mgmt",
							Name:                       "partition-2",
							Privatenetworkprefixlength: 24,
						},
					}, nil)
				},
			},
			want: []*models.V1PartitionResponse{
				{
					Bootconfig: &models.V1PartitionBootConfiguration{
						Commandline: "commandline",
						Imageurl:    "imageurl",
						Kernelurl:   "kernelurl",
					},
					Description:                "partition 1",
					ID:                         pointer.Pointer("1"),
					Mgmtserviceaddress:         "mgmt",
					Name:                       "partition-1",
					Privatenetworkprefixlength: 24,
				},
				{
					Bootconfig: &models.V1PartitionBootConfiguration{
						Commandline: "commandline",
						Imageurl:    "imageurl",
						Kernelurl:   "kernelurl",
					},
					Description:                "partition 2",
					ID:                         pointer.Pointer("2"),
					Mgmtserviceaddress:         "mgmt",
					Name:                       "partition-2",
					Privatenetworkprefixlength: 24,
				},
			},
			wantTable: `
ID   NAME          DESCRIPTION
1    partition-1   partition 1
2    partition-2   partition 2
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
1 partition-1
2 partition-2
`,
			wantMarkdown: `
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
|  2 | partition-2 | partition 2 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[[]*models.V1PartitionResponse]{
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
					os.Args = append([]string{binaryName, "partition", "apply", "-f", "/file.yaml"}, format.Args()...)

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
