package cmd

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	imageExpiration = pointer.Pointer(strfmt.DateTime(testTime.Add(3 * 24 * time.Hour)))
	image1          = &models.V1ImageResponse{
		Classification: models.DatastoreImageSearchQueryClassificationSupported,
		Description:    "debian-description",
		ExpirationDate: imageExpiration,
		Features:       []string{"machine"},
		ID:             pointer.Pointer("debian"),
		Name:           "debian-name",
		URL:            "debian-url",
		Usedby:         []string{"abc-def"},
	}
	image2 = &models.V1ImageResponse{
		Classification: models.DatastoreImageSearchQueryClassificationSupported,
		Description:    "ubuntu-description",
		ExpirationDate: imageExpiration,
		Features:       []string{"machine"},
		ID:             pointer.Pointer("ubuntu"),
		Name:           "ubuntu-name",
		URL:            "ubuntu-url",
		Usedby:         []string{"123"},
	}
)

func Test_ImageCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1ImageResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1ImageResponse) []string {
				return []string{"image", "list", "--show-usage"}
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("FindImages", testcommon.MatchIgnoreContext(t, image.NewFindImagesParams().WithBody(&models.V1ImageFindRequest{
						Features: []string{},
					}).WithShowUsage(pointer.Pointer(true))), nil).Return(&image.FindImagesOK{
						Payload: []*models.V1ImageResponse{
							image2,
							image1,
						},
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				image1,
				image2,
			},
			wantTable: pointer.Pointer(`
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS      USEDBY
debian   debian-name   debian-description   machine    3d           supported   1
ubuntu   ubuntu-name   ubuntu-description   machine    3d           supported   1
`),
			wantWideTable: pointer.Pointer(`
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS      USEDBY
debian   debian-name   debian-description   machine    3d           supported   abc-def
ubuntu   ubuntu-name   ubuntu-description   machine    3d           supported   123
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
debian debian-name
ubuntu ubuntu-name
`),
			wantMarkdown: pointer.Pointer(`
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION |  STATUS   | USEDBY |
|--------|-------------|--------------------|----------|------------|-----------|--------|
| debian | debian-name | debian-description | machine  | 3d         | supported |      1 |
| ubuntu | ubuntu-name | ubuntu-description | machine  | 3d         | supported |      1 |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1ImageResponse) []string {
				args := []string{"image", "list",
					"--id", *want[0].ID,
					"--name", want[0].Name,
					"--classification", want[0].Classification,
					"--features", want[0].Features[0],
					"--os", "debian", "--version", "10",
					"--show-usage",
				}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("FindImages", testcommon.MatchIgnoreContext(t, image.NewFindImagesParams().WithBody(&models.V1ImageFindRequest{
						Classification: image1.Classification,
						Features:       image1.Features,
						ID:             *image1.ID,
						Name:           image1.Name,
						Os:             "debian",
						Version:        "10",
					}).WithShowUsage(pointer.Pointer(true))), nil).Return(&image.FindImagesOK{
						Payload: []*models.V1ImageResponse{
							image1,
						},
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				image1,
			},
			wantTable: pointer.Pointer(`
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS      USEDBY
debian   debian-name   debian-description   machine    3d           supported   1
`),
			wantWideTable: pointer.Pointer(`
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS      USEDBY
debian   debian-name   debian-description   machine    3d           supported   abc-def
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
debian debian-name
`),
			wantMarkdown: pointer.Pointer(`
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION |  STATUS   | USEDBY |
|--------|-------------|--------------------|----------|------------|-----------|--------|
| debian | debian-name | debian-description | machine  | 3d         | supported |      1 |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1ImageResponse) []string {
				return appendFromFileCommonArgs("image", "apply")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1ImageResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("CreateImage", testcommon.MatchIgnoreContext(t, image.NewCreateImageParams().WithBody(imageResponseToCreate(image1)), testcommon.StrFmtDateComparer()), nil).Return(nil, &image.CreateImageConflict{}).Once()
					mock.On("UpdateImage", testcommon.MatchIgnoreContext(t, image.NewUpdateImageParams().WithBody(imageResponseToUpdate(image1)), testcommon.StrFmtDateComparer()), nil).Return(&image.UpdateImageOK{
						Payload: image1,
					}, nil)
					mock.On("CreateImage", testcommon.MatchIgnoreContext(t, image.NewCreateImageParams().WithBody(imageResponseToCreate(image2)), testcommon.StrFmtDateComparer()), nil).Return(&image.CreateImageCreated{
						Payload: image2,
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				image1,
				image2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1ImageResponse) []string {
				return appendFromFileCommonArgs("image", "create")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1ImageResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("CreateImage", testcommon.MatchIgnoreContext(t, image.NewCreateImageParams().WithBody(imageResponseToCreate(image1)), testcommon.StrFmtDateComparer()), nil).Return(&image.CreateImageCreated{
						Payload: image1,
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				image1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1ImageResponse) []string {
				return appendFromFileCommonArgs("image", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1ImageResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("UpdateImage", testcommon.MatchIgnoreContext(t, image.NewUpdateImageParams().WithBody(imageResponseToUpdate(image1)), testcommon.StrFmtDateComparer()), nil).Return(&image.UpdateImageOK{
						Payload: image1,
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				image1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1ImageResponse) []string {
				return appendFromFileCommonArgs("image", "delete")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1ImageResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("DeleteImage", testcommon.MatchIgnoreContext(t, image.NewDeleteImageParams().WithID(*image1.ID)), nil).Return(&image.DeleteImageOK{
						Payload: image1,
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				image1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_ImageCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1ImageResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1ImageResponse) []string {
				return []string{"image", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("FindImage", testcommon.MatchIgnoreContext(t, image.NewFindImageParams().WithID(*image1.ID)), nil).Return(&image.FindImageOK{
						Payload: image1,
					}, nil)
				},
			},
			want: image1,
			wantTable: pointer.Pointer(`
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS      USEDBY
debian   debian-name   debian-description   machine    3d           supported
		`),
			wantWideTable: pointer.Pointer(`
ID       NAME          DESCRIPTION          FEATURES   EXPIRATION   STATUS      USEDBY
debian   debian-name   debian-description   machine    3d           supported
		`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
debian debian-name
		`),
			wantMarkdown: pointer.Pointer(`
|   ID   |    NAME     |    DESCRIPTION     | FEATURES | EXPIRATION |  STATUS   | USEDBY |
|--------|-------------|--------------------|----------|------------|-----------|--------|
| debian | debian-name | debian-description | machine  | 3d         | supported |        |
		`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1ImageResponse) []string {
				return []string{"image", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("DeleteImage", testcommon.MatchIgnoreContext(t, image.NewDeleteImageParams().WithID(*image1.ID)), nil).Return(&image.DeleteImageOK{
						Payload: image1,
					}, nil)
				},
			},
			want: image1,
		},
		{
			name: "create",
			cmd: func(want *models.V1ImageResponse) []string {
				args := []string{"image", "create",
					"--id", *want.ID,
					"--name", want.Name,
					"--description", want.Description,
					"--url", want.URL,
					"--features", want.Features[0],
				}
				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
				return args
			},
			mocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					i := image1
					i.Classification = ""
					i.ExpirationDate = &strfmt.DateTime{}
					mock.On("CreateImage", testcommon.MatchIgnoreContext(t, image.NewCreateImageParams().WithBody(imageResponseToCreate(i))), nil).Return(&image.CreateImageCreated{
						Payload: image1,
					}, nil)
				},
			},
			want: image1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
