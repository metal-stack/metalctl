package cmd

import (
	"strconv"
	"testing"

	"github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	size1 = &models.V1SizeResponse{
		Constraints: []*models.V1SizeConstraint{
			{
				Max:  pointer.Pointer(int64(2)),
				Min:  pointer.Pointer(int64(1)),
				Type: pointer.Pointer("storage"),
			},
			{
				Max:  pointer.Pointer(int64(4)),
				Min:  pointer.Pointer(int64(3)),
				Type: pointer.Pointer("memory"),
			},
			{
				Max:  pointer.Pointer(int64(6)),
				Min:  pointer.Pointer(int64(5)),
				Type: pointer.Pointer("cores"),
			},
		},
		Description: "size 1",
		ID:          pointer.Pointer("1"),
		Name:        "size-1",
	}
	size2 = &models.V1SizeResponse{
		Constraints: []*models.V1SizeConstraint{
			{
				Max:  pointer.Pointer(int64(2)),
				Min:  pointer.Pointer(int64(1)),
				Type: pointer.Pointer("storage"),
			},
			{
				Max:  pointer.Pointer(int64(4)),
				Min:  pointer.Pointer(int64(3)),
				Type: pointer.Pointer("memory"),
			},
			{
				Max:  pointer.Pointer(int64(6)),
				Min:  pointer.Pointer(int64(5)),
				Type: pointer.Pointer("cores"),
			},
		},
		Description: "size 2",
		ID:          pointer.Pointer("2"),
		Name:        "size-2",
	}
)

func Test_SizeCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1SizeResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1SizeResponse) []string {
				return []string{"size", "list"}
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("ListSizes", testcommon.MatchIgnoreContext(t, size.NewListSizesParams()), nil).Return(&size.ListSizesOK{
						Payload: []*models.V1SizeResponse{
							size2,
							size1,
						},
					}, nil)
				},
			},
			want: []*models.V1SizeResponse{
				size1,
				size2,
			},
			wantTable: pointer.Pointer(`
ID   NAME     DESCRIPTION   CPU RANGE   MEMORY RANGE   STORAGE RANGE
1    size-1   size 1        5 - 6       3 B - 4 B      1 B - 2 B
2    size-2   size 2        5 - 6       3 B - 4 B      1 B - 2 B
`),
			wantWideTable: pointer.Pointer(`
ID   NAME     DESCRIPTION   CPU RANGE   MEMORY RANGE   STORAGE RANGE
1    size-1   size 1        5 - 6       3 B - 4 B      1 B - 2 B
2    size-2   size 2        5 - 6       3 B - 4 B      1 B - 2 B
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 size-1
2 size-2
`),
			wantMarkdown: pointer.Pointer(`
| ID |  NAME  | DESCRIPTION | CPU RANGE | MEMORY RANGE | STORAGE RANGE |
|----|--------|-------------|-----------|--------------|---------------|
|  1 | size-1 | size 1      | 5 - 6     | 3 B - 4 B    | 1 B - 2 B     |
|  2 | size-2 | size 2      | 5 - 6     | 3 B - 4 B    | 1 B - 2 B     |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1SizeResponse) []string {
				return []string{"size", "apply", "--bulk-output", "-f", "/file.yaml", "--force"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("CreateSize", testcommon.MatchIgnoreContext(t, size.NewCreateSizeParams().WithBody(sizeResponseToCreate(size1))), nil).Return(nil, &size.CreateSizeConflict{}).Once()
					mock.On("UpdateSize", testcommon.MatchIgnoreContext(t, size.NewUpdateSizeParams().WithBody(sizeResponseToUpdate(size1))), nil).Return(&size.UpdateSizeOK{
						Payload: size1,
					}, nil)
					mock.On("CreateSize", testcommon.MatchIgnoreContext(t, size.NewCreateSizeParams().WithBody(sizeResponseToCreate(size2))), nil).Return(&size.CreateSizeCreated{
						Payload: size2,
					}, nil)
				},
			},
			want: []*models.V1SizeResponse{
				size1,
				size2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1SizeResponse) []string {
				return []string{"size", "create", "-f", "/file.yaml", "--force"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("CreateSize", testcommon.MatchIgnoreContext(t, size.NewCreateSizeParams().WithBody(sizeResponseToCreate(size1))), nil).Return(&size.CreateSizeCreated{
						Payload: size1,
					}, nil)
				},
			},
			want: []*models.V1SizeResponse{
				size1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1SizeResponse) []string {
				return []string{"size", "update", "-f", "/file.yaml", "--force"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("UpdateSize", testcommon.MatchIgnoreContext(t, size.NewUpdateSizeParams().WithBody(sizeResponseToUpdate(size1))), nil).Return(&size.UpdateSizeOK{
						Payload: size1,
					}, nil)
				},
			},
			want: []*models.V1SizeResponse{
				size1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1SizeResponse) []string {
				return []string{"size", "delete", "-f", "/file.yaml", "--force"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("DeleteSize", testcommon.MatchIgnoreContext(t, size.NewDeleteSizeParams().WithID(*size1.ID)), nil).Return(&size.DeleteSizeOK{
						Payload: size1,
					}, nil)
				},
			},
			want: []*models.V1SizeResponse{
				size1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_SizeCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1SizeResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1SizeResponse) []string {
				return []string{"size", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("FindSize", testcommon.MatchIgnoreContext(t, size.NewFindSizeParams().WithID(*size1.ID)), nil).Return(&size.FindSizeOK{
						Payload: size1,
					}, nil)
				},
			},
			want: size1,
			wantTable: pointer.Pointer(`
ID   NAME     DESCRIPTION   CPU RANGE   MEMORY RANGE   STORAGE RANGE
1    size-1   size 1        5 - 6       3 B - 4 B      1 B - 2 B
`),
			wantWideTable: pointer.Pointer(`
ID   NAME     DESCRIPTION   CPU RANGE   MEMORY RANGE   STORAGE RANGE
1    size-1   size 1        5 - 6       3 B - 4 B      1 B - 2 B
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 size-1
`),
			wantMarkdown: pointer.Pointer(`
| ID |  NAME  | DESCRIPTION | CPU RANGE | MEMORY RANGE | STORAGE RANGE |
|----|--------|-------------|-----------|--------------|---------------|
|  1 | size-1 | size 1      | 5 - 6     | 3 B - 4 B    | 1 B - 2 B     |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1SizeResponse) []string {
				return []string{"size", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("DeleteSize", testcommon.MatchIgnoreContext(t, size.NewDeleteSizeParams().WithID(*size1.ID)), nil).Return(&size.DeleteSizeOK{
						Payload: size1,
					}, nil)
				},
			},
			want: size1,
		},
		{
			name: "create",
			cmd: func(want *models.V1SizeResponse) []string {
				args := []string{"size", "create",
					"--id", *want.ID,
					"--name", want.Name,
					"--description", want.Description,
					"--max", strconv.FormatInt(*want.Constraints[0].Max, 10),
					"--min", strconv.FormatInt(*want.Constraints[0].Min, 10),
					"--type", *want.Constraints[0].Type,
				}
				assertExhaustiveArgs(t, args, "file", "bulk-output", "force")
				return args
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					s := size1
					s.Constraints = []*models.V1SizeConstraint{
						{
							Max:  size1.Constraints[0].Max,
							Min:  size1.Constraints[0].Min,
							Type: size1.Constraints[0].Type,
						},
					}
					mock.On("CreateSize", testcommon.MatchIgnoreContext(t, size.NewCreateSizeParams().WithBody(sizeResponseToCreate(size1))), nil).Return(&size.CreateSizeCreated{
						Payload: size1,
					}, nil)
				},
			},
			want: size1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
