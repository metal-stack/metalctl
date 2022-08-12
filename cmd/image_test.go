package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_ImageListCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         []*models.V1ImageResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "list images",
			metalMocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("ListImages", testcommon.MatchIgnoreContext(t, image.NewListImagesParams().WithShowUsage(pointer.Pointer(false))), nil).Return(&image.ListImagesOK{
						Payload: []*models.V1ImageResponse{
							{
								Features:    []string{"machine"},
								ID:          pointer.Pointer("ubuntu"),
								Name:        "ubuntu-name",
								Description: "ubuntu-description",
								Usedby:      []string{},
							},
							{
								Features:    []string{"machine"},
								ID:          pointer.Pointer("debian"),
								Name:        "debian-name",
								Description: "debian-description",
								Usedby:      []string{},
							},
						},
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				{
					Features:    []string{"machine"},
					ID:          pointer.Pointer("debian"),
					Name:        "debian-name",
					Description: "debian-description",
					Usedby:      []string{},
				},
				{
					Features:    []string{"machine"},
					ID:          pointer.Pointer("ubuntu"),
					Name:        "ubuntu-name",
					Description: "ubuntu-description",
					Usedby:      []string{},
				},
			},
			wantTable: `
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS   USEDBY
debian   debian-name   debian-description   machine                          0
ubuntu   ubuntu-name   ubuntu-description   machine                          0
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
debian debian-name
ubuntu ubuntu-name
`,
			wantMarkdown: `
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION | STATUS | USEDBY |
|--------|-------------|--------------------|----------|------------|--------|--------|
| debian | debian-name | debian-description | machine  |            |        |      0 |
| ubuntu | ubuntu-name | ubuntu-description | machine  |            |        |      0 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[[]*models.V1ImageResponse]{
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
					os.Args = append([]string{binaryName, "image", "list"}, format.Args()...)

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

func Test_ImageDescribeCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ImageResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "describe image",
			metalMocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("FindImage", testcommon.MatchIgnoreContext(t, image.NewFindImageParams().WithID("ubuntu")), nil).Return(&image.FindImageOK{
						Payload: &models.V1ImageResponse{
							Features:    []string{"machine"},
							ID:          pointer.Pointer("ubuntu"),
							Name:        "ubuntu-name",
							Description: "ubuntu-description",
							Usedby:      []string{},
						},
					}, nil)
				},
			},
			want: &models.V1ImageResponse{
				Features:    []string{"machine"},
				ID:          pointer.Pointer("ubuntu"),
				Name:        "ubuntu-name",
				Description: "ubuntu-description",
				Usedby:      []string{},
			},
			wantTable: `
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS   USEDBY
ubuntu   ubuntu-name   ubuntu-description   machine                          0
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
ubuntu ubuntu-name
`,
			wantMarkdown: `
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION | STATUS | USEDBY |
|--------|-------------|--------------------|----------|------------|--------|--------|
| ubuntu | ubuntu-name | ubuntu-description | machine  |            |        |      0 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ImageResponse]{
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
					os.Args = append([]string{binaryName, "image", "describe", *tt.want.ID}, format.Args()...)

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

func Test_ImageDeleteCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ImageResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "remove image",
			metalMocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("DeleteImage", testcommon.MatchIgnoreContext(t, image.NewDeleteImageParams().WithID("ubuntu")), nil).Return(&image.DeleteImageOK{
						Payload: &models.V1ImageResponse{
							Features:    []string{"machine"},
							ID:          pointer.Pointer("ubuntu"),
							Name:        "ubuntu-name",
							Description: "ubuntu-description",
							Usedby:      []string{},
						},
					}, nil)
				},
			},
			want: &models.V1ImageResponse{
				Features:    []string{"machine"},
				ID:          pointer.Pointer("ubuntu"),
				Name:        "ubuntu-name",
				Description: "ubuntu-description",
				Usedby:      []string{},
			},
			wantTable: `
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS   USEDBY
ubuntu   ubuntu-name   ubuntu-description   machine                          0
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
ubuntu ubuntu-name
`,
			wantMarkdown: `
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION | STATUS | USEDBY |
|--------|-------------|--------------------|----------|------------|--------|--------|
| ubuntu | ubuntu-name | ubuntu-description | machine  |            |        |      0 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ImageResponse]{
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
					os.Args = append([]string{binaryName, "image", "rm", *tt.want.ID}, format.Args()...)

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

func Test_ImageCreateFromCLICmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ImageResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "create image",
			metalMocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("CreateImage", testcommon.MatchIgnoreContext(t, image.NewCreateImageParams().WithBody(&models.V1ImageCreateRequest{
						ID:          pointer.Pointer("ubuntu"),
						Name:        "ubuntu-name",
						Description: "ubuntu-description",
						Features:    []string{"machine"},
						URL:         pointer.Pointer("url"),
					})), nil).Return(&image.CreateImageCreated{
						Payload: &models.V1ImageResponse{
							ID:          pointer.Pointer("ubuntu"),
							Name:        "ubuntu-name",
							Description: "ubuntu-description",
							Features:    []string{"machine"},
							URL:         "url",
							Usedby:      []string{},
						},
					}, nil)
				},
			},
			want: &models.V1ImageResponse{
				ID:          pointer.Pointer("ubuntu"),
				Name:        "ubuntu-name",
				Description: "ubuntu-description",
				Features:    []string{"machine"},
				URL:         "url",
				Usedby:      []string{},
			},
			wantTable: `
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS   USEDBY
ubuntu   ubuntu-name   ubuntu-description   machine                          0
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
ubuntu ubuntu-name
`,
			wantMarkdown: `
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION | STATUS | USEDBY |
|--------|-------------|--------------------|----------|------------|--------|--------|
| ubuntu | ubuntu-name | ubuntu-description | machine  |            |        |      0 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ImageResponse]{
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
					os.Args = append([]string{binaryName, "image", "create",
						"--id", *tt.want.ID,
						"--name", tt.want.Name,
						"--description", tt.want.Description,
						"--url", tt.want.URL,
						"--features", tt.want.Features[0],
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

func Test_ImageCreateFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ImageResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "create image",
			metalMocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("CreateImage", testcommon.MatchIgnoreContext(t, image.NewCreateImageParams().WithBody(&models.V1ImageCreateRequest{
						ID:          pointer.Pointer("ubuntu"),
						Name:        "ubuntu-name",
						Description: "ubuntu-description",
						Features:    []string{"machine"},
						URL:         pointer.Pointer("url"),
					})), nil).Return(&image.CreateImageCreated{
						Payload: &models.V1ImageResponse{
							ID:          pointer.Pointer("ubuntu"),
							Name:        "ubuntu-name",
							Description: "ubuntu-description",
							Features:    []string{"machine"},
							URL:         "url",
							Usedby:      []string{},
						},
					}, nil)
				},
			},
			want: &models.V1ImageResponse{
				ID:          pointer.Pointer("ubuntu"),
				Name:        "ubuntu-name",
				Description: "ubuntu-description",
				Features:    []string{"machine"},
				URL:         "url",
				Usedby:      []string{},
			},
			wantTable: `
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS   USEDBY
ubuntu   ubuntu-name   ubuntu-description   machine                          0
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
ubuntu ubuntu-name
`,
			wantMarkdown: `
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION | STATUS | USEDBY |
|--------|-------------|--------------------|----------|------------|--------|--------|
| ubuntu | ubuntu-name | ubuntu-description | machine  |            |        |      0 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ImageResponse]{
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
					os.Args = append([]string{binaryName, "image", "create", "-f", "/file.yaml"}, format.Args()...)

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

func Test_ImageUpdateFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         *models.V1ImageResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "update image from file",
			metalMocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("UpdateImage", testcommon.MatchIgnoreContext(t, image.NewUpdateImageParams().WithBody(&models.V1ImageUpdateRequest{
						ID:          pointer.Pointer("ubuntu"),
						Name:        "ubuntu-name",
						Description: "ubuntu-description",
						Features:    []string{"machine"},
						URL:         "url",
						Usedby:      []string{},
					})), nil).Return(&image.UpdateImageOK{
						Payload: &models.V1ImageResponse{
							ID:          pointer.Pointer("ubuntu"),
							Name:        "ubuntu-name",
							Description: "ubuntu-description",
							Features:    []string{"machine"},
							URL:         "url",
							Usedby:      []string{},
						},
					}, nil)
				},
			},
			want: &models.V1ImageResponse{
				ID:          pointer.Pointer("ubuntu"),
				Name:        "ubuntu-name",
				Description: "ubuntu-description",
				Features:    []string{"machine"},
				URL:         "url",
				Usedby:      []string{},
			},
			wantTable: `
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS   USEDBY
ubuntu   ubuntu-name   ubuntu-description   machine                          0
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
ubuntu ubuntu-name
`,
			wantMarkdown: `
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION | STATUS | USEDBY |
|--------|-------------|--------------------|----------|------------|--------|--------|
| ubuntu | ubuntu-name | ubuntu-description | machine  |            |        |      0 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[*models.V1ImageResponse]{
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
					os.Args = append([]string{binaryName, "image", "update", "-f", "/file.yaml"}, format.Args()...)

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

func Test_ImageApplyFromFileCmd(t *testing.T) {
	tests := []struct {
		name         string
		metalMocks   *client.MetalMockFns
		want         []*models.V1ImageResponse
		wantTable    string
		template     string
		wantTemplate string
		wantMarkdown string
		wantErr      error
	}{
		{
			name: "apply images from file",
			metalMocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("CreateImage", testcommon.MatchIgnoreContext(t, image.NewCreateImageParams().WithBody(&models.V1ImageCreateRequest{
						ID:          pointer.Pointer("debian"),
						Name:        "debian-name",
						Description: "debian-description",
						Features:    []string{"machine"},
						URL:         pointer.Pointer("debian-url"),
					})), nil).Return(nil, &image.CreateImageConflict{}).Once()
					mock.On("UpdateImage", testcommon.MatchIgnoreContext(t, image.NewUpdateImageParams().WithBody(&models.V1ImageUpdateRequest{
						ID:          pointer.Pointer("debian"),
						Name:        "debian-name",
						Description: "debian-description",
						Features:    []string{"machine"},
						URL:         "debian-url",
						Usedby:      []string{},
					})), nil).Return(&image.UpdateImageOK{
						Payload: &models.V1ImageResponse{
							ID:          pointer.Pointer("debian"),
							Name:        "debian-name",
							Description: "debian-description",
							Features:    []string{"machine"},
							URL:         "debian-url",
							Usedby:      []string{},
						},
					}, nil)
					mock.On("CreateImage", testcommon.MatchIgnoreContext(t, image.NewCreateImageParams().WithBody(&models.V1ImageCreateRequest{
						ID:          pointer.Pointer("ubuntu"),
						Name:        "ubuntu-name",
						Description: "ubuntu-description",
						Features:    []string{"machine"},
						URL:         pointer.Pointer("ubuntu-url"),
					})), nil).Return(&image.CreateImageCreated{
						Payload: &models.V1ImageResponse{
							ID:          pointer.Pointer("ubuntu"),
							Name:        "ubuntu-name",
							Description: "ubuntu-description",
							Features:    []string{"machine"},
							URL:         "ubuntu-url",
							Usedby:      []string{},
						},
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				{
					ID:          pointer.Pointer("debian"),
					Name:        "debian-name",
					Features:    []string{"machine"},
					Description: "debian-description",
					URL:         "debian-url",
					Usedby:      []string{},
				},
				{
					ID:          pointer.Pointer("ubuntu"),
					Name:        "ubuntu-name",
					Description: "ubuntu-description",
					Features:    []string{"machine"},
					URL:         "ubuntu-url",
					Usedby:      []string{},
				},
			},
			wantTable: `
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS   USEDBY
debian   debian-name   debian-description   machine                          0
ubuntu   ubuntu-name   ubuntu-description   machine                          0
`,
			template: "{{ .id }} {{ .name }}",
			wantTemplate: `
debian debian-name
ubuntu ubuntu-name
`,
			wantMarkdown: `
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION | STATUS | USEDBY |
|--------|-------------|--------------------|----------|------------|--------|--------|
| debian | debian-name | debian-description | machine  |            |        |      0 |
| ubuntu | ubuntu-name | ubuntu-description | machine  |            |        |      0 |
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		formats := &outputsFormatConfig[[]*models.V1ImageResponse]{
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
					os.Args = append([]string{binaryName, "image", "apply", "-f", "/file.yaml"}, format.Args()...)

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
