package cmd

import (
	"testing"

	"github.com/metal-stack/metal-go/api/client/sizeimageconstraint"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	sic1 = &models.V1SizeImageConstraintResponse{
		Constraints: &models.V1SizeImageConstraintBase{
			Images: map[string]string{
				"os-image": "*",
			},
		},
		Description: "sic 1",
		ID:          new("1"),
		Name:        "sic-1",
	}
	sic2 = &models.V1SizeImageConstraintResponse{
		Constraints: &models.V1SizeImageConstraintBase{
			Images: map[string]string{
				"os-image": "*",
			},
		},
		Description: "sic 2",
		ID:          new("2"),
		Name:        "sic-2",
	}
)

func Test_SizeImageConstraintCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1SizeImageConstraintResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1SizeImageConstraintResponse) []string {
				return []string{"size", "imageconstraint", "list"}
			},
			mocks: &client.MetalMockFns{
				Sizeimageconstraint: func(mock *mock.Mock) {
					mock.On("ListSizeImageConstraints", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewListSizeImageConstraintsParams()), nil).Return(&sizeimageconstraint.ListSizeImageConstraintsOK{
						Payload: []*models.V1SizeImageConstraintResponse{
							sic2,
							sic1,
						},
					}, nil)
				},
			},
			want: []*models.V1SizeImageConstraintResponse{
				sic1,
				sic2,
			},
			wantTable: new(`
ID  NAME   DESCRIPTION  IMAGE     CONSTRAINT
1   sic-1  sic 1        os-image  *
2   sic-2  sic 2        os-image  *
`),
			wantWideTable: new(`
ID  NAME   DESCRIPTION  IMAGE     CONSTRAINT
1   sic-1  sic 1        os-image  *
2   sic-2  sic 2        os-image  *
`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
1 sic-1
2 sic-2
`),
			wantMarkdown: new(`
| ID | NAME  | DESCRIPTION | IMAGE    | CONSTRAINT |
|----|-------|-------------|----------|------------|
| 1  | sic-1 | sic 1       | os-image | *          |
| 2  | sic-2 | sic 2       | os-image | *          |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1SizeImageConstraintResponse) []string {
				return appendFromFileCommonArgs("size", "imageconstraint", "apply")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeImageConstraintResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Sizeimageconstraint: func(mock *mock.Mock) {
					mock.On("CreateSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewCreateSizeImageConstraintParams().WithBody(sizeImageConstraintResponseToCreate(sic1))), nil).Return(nil, &sizeimageconstraint.CreateSizeImageConstraintConflict{}).Once()
					mock.On("UpdateSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewUpdateSizeImageConstraintParams().WithBody(sizeImageConstraintResponseToUpdate(sic1))), nil).Return(&sizeimageconstraint.UpdateSizeImageConstraintOK{
						Payload: sic1,
					}, nil)
					mock.On("CreateSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewCreateSizeImageConstraintParams().WithBody(sizeImageConstraintResponseToCreate(sic2))), nil).Return(&sizeimageconstraint.CreateSizeImageConstraintCreated{
						Payload: sic2,
					}, nil)
				},
			},
			want: []*models.V1SizeImageConstraintResponse{
				sic1,
				sic2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1SizeImageConstraintResponse) []string {
				return appendFromFileCommonArgs("size", "imageconstraint", "create")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeImageConstraintResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Sizeimageconstraint: func(mock *mock.Mock) {
					mock.On("CreateSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewCreateSizeImageConstraintParams().WithBody(sizeImageConstraintResponseToCreate(sic1))), nil).Return(&sizeimageconstraint.CreateSizeImageConstraintCreated{
						Payload: sic1,
					}, nil)
				},
			},
			want: []*models.V1SizeImageConstraintResponse{
				sic1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1SizeImageConstraintResponse) []string {
				return appendFromFileCommonArgs("size", "imageconstraint", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeImageConstraintResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Sizeimageconstraint: func(mock *mock.Mock) {
					mock.On("UpdateSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewUpdateSizeImageConstraintParams().WithBody(sizeImageConstraintResponseToUpdate(sic1))), nil).Return(&sizeimageconstraint.UpdateSizeImageConstraintOK{
						Payload: sic1,
					}, nil)
				},
			},
			want: []*models.V1SizeImageConstraintResponse{
				sic1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_SizeImageConstraintCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1SizeImageConstraintResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1SizeImageConstraintResponse) []string {
				return []string{"size", "imageconstraint", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Sizeimageconstraint: func(mock *mock.Mock) {
					mock.On("FindSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewFindSizeImageConstraintParams().WithID(*sic1.ID)), nil).Return(&sizeimageconstraint.FindSizeImageConstraintOK{
						Payload: sic1,
					}, nil)
				},
			},
			want: sic1,
			wantTable: new(`
ID  NAME   DESCRIPTION  IMAGE     CONSTRAINT
1   sic-1  sic 1        os-image  *
		`),
			wantWideTable: new(`
ID  NAME   DESCRIPTION  IMAGE     CONSTRAINT
1   sic-1  sic 1        os-image  *
		`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
1 sic-1
		`),
			wantMarkdown: new(`
| ID | NAME  | DESCRIPTION | IMAGE    | CONSTRAINT |
|----|-------|-------------|----------|------------|
| 1  | sic-1 | sic 1       | os-image | *          |
		`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1SizeImageConstraintResponse) []string {
				return []string{"size", "imageconstraint", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Sizeimageconstraint: func(mock *mock.Mock) {
					mock.On("DeleteSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewDeleteSizeImageConstraintParams().WithID(*sic1.ID)), nil).Return(&sizeimageconstraint.DeleteSizeImageConstraintOK{
						Payload: sic1,
					}, nil)
				},
			},
			want: sic1,
		},
		{
			name: "create from file",
			cmd: func(want *models.V1SizeImageConstraintResponse) []string {
				return []string{"size", "imageconstraint", "create", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1SizeImageConstraintResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Sizeimageconstraint: func(mock *mock.Mock) {
					mock.On("CreateSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewCreateSizeImageConstraintParams().WithBody(sizeImageConstraintResponseToCreate(sic1))), nil).Return(&sizeimageconstraint.CreateSizeImageConstraintCreated{
						Payload: sic1,
					}, nil)
				},
			},
			want: sic1,
		},
		{
			name: "update from file",
			cmd: func(want *models.V1SizeImageConstraintResponse) []string {
				return []string{"size", "imageconstraint", "update", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1SizeImageConstraintResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Sizeimageconstraint: func(mock *mock.Mock) {
					mock.On("UpdateSizeImageConstraint", testcommon.MatchIgnoreContext(t, sizeimageconstraint.NewUpdateSizeImageConstraintParams().WithBody(sizeImageConstraintResponseToUpdate(sic1))), nil).Return(&sizeimageconstraint.UpdateSizeImageConstraintOK{
						Payload: sic1,
					}, nil)
				},
			},
			want: sic1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
