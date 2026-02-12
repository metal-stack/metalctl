package cmd

import (
	"testing"

	"github.com/metal-stack/metal-go/api/client/size"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	rv1 = &models.V1SizeReservationResponse{
		Amount:      new(int32(3)),
		Description: "this is reservation 1",
		ID:          new("r1"),
		Labels: map[string]string{
			"a": "b",
		},
		Name:         "reservation 1",
		Partitionids: []string{"partition-a", "partition-b"},
		Projectid:    new("project-a"),
		Sizeid:       new("size-a"),
	}
	rv2 = &models.V1SizeReservationResponse{
		Amount:      new(int32(2)),
		Description: "this is reservation 2",
		ID:          new("r2"),
		Labels: map[string]string{
			"b": "c",
		},
		Name:         "reservation 2",
		Partitionids: []string{"partition-b"},
		Projectid:    new("project-b"),
		Sizeid:       new("size-b"),
	}
)

func Test_SizeReservationCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1SizeReservationResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1SizeReservationResponse) []string {
				return []string{"size", "reservation", "list"}
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("FindSizeReservations", testcommon.MatchIgnoreContext(t, size.NewFindSizeReservationsParams().WithBody(&models.V1SizeReservationListRequest{})), nil).Return(&size.FindSizeReservationsOK{
						Payload: []*models.V1SizeReservationResponse{
							rv2,
							rv1,
						},
					}, nil)
				},
			},
			want: []*models.V1SizeReservationResponse{
				rv1,
				rv2,
			},
			wantTable: new(`
ID  SIZE    PROJECT    PARTITIONS                DESCRIPTION            AMOUNT
r1  size-a  project-a  partition-a, partition-b  this is reservation 1  3
r2  size-b  project-b  partition-b               this is reservation 2  2
`),
			wantWideTable: new(`
ID  SIZE    PROJECT    PARTITIONS                DESCRIPTION            AMOUNT  LABELS
r1  size-a  project-a  partition-a, partition-b  this is reservation 1  3       a=b
r2  size-b  project-b  partition-b               this is reservation 2  2       b=c
`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
r1 reservation 1
r2 reservation 2
`),
			wantMarkdown: new(`
| ID | SIZE   | PROJECT   | PARTITIONS               | DESCRIPTION           | AMOUNT |
|----|--------|-----------|--------------------------|-----------------------|--------|
| r1 | size-a | project-a | partition-a, partition-b | this is reservation 1 | 3      |
| r2 | size-b | project-b | partition-b              | this is reservation 2 | 2      |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1SizeReservationResponse) []string {
				return appendFromFileCommonArgs("size", "reservation", "apply")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("CreateSizeReservation", testcommon.MatchIgnoreContext(t, size.NewCreateSizeReservationParams().WithBody(sizeReservationResponseToCreate(rv1))), nil).Return(nil, &size.CreateSizeReservationConflict{}).Once()
					mock.On("UpdateSizeReservation", testcommon.MatchIgnoreContext(t, size.NewUpdateSizeReservationParams().WithBody(sizeReservationResponseToUpdate(rv1))), nil).Return(&size.UpdateSizeReservationOK{
						Payload: rv1,
					}, nil)
					mock.On("CreateSizeReservation", testcommon.MatchIgnoreContext(t, size.NewCreateSizeReservationParams().WithBody(sizeReservationResponseToCreate(rv2))), nil).Return(&size.CreateSizeReservationCreated{
						Payload: rv2,
					}, nil)
				},
			},
			want: []*models.V1SizeReservationResponse{
				rv1,
				rv2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1SizeReservationResponse) []string {
				return appendFromFileCommonArgs("size", "reservation", "create")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("CreateSizeReservation", testcommon.MatchIgnoreContext(t, size.NewCreateSizeReservationParams().WithBody(sizeReservationResponseToCreate(rv1))), nil).Return(&size.CreateSizeReservationCreated{
						Payload: rv1,
					}, nil)
				},
			},
			want: []*models.V1SizeReservationResponse{
				rv1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1SizeReservationResponse) []string {
				return appendFromFileCommonArgs("size", "reservation", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SizeReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("UpdateSizeReservation", testcommon.MatchIgnoreContext(t, size.NewUpdateSizeReservationParams().WithBody(sizeReservationResponseToUpdate(rv1))), nil).Return(&size.UpdateSizeReservationOK{
						Payload: rv1,
					}, nil)
				},
			},
			want: []*models.V1SizeReservationResponse{
				rv1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_SizeReservationCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1SizeReservationResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1SizeReservationResponse) []string {
				return []string{"size", "reservation", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("GetSizeReservation", testcommon.MatchIgnoreContext(t, size.NewGetSizeReservationParams().WithID(*rv1.ID)), nil).Return(&size.GetSizeReservationOK{
						Payload: rv1,
					}, nil)
				},
			},
			want: rv1,
			wantTable: new(`
ID  SIZE    PROJECT    PARTITIONS                DESCRIPTION            AMOUNT
r1  size-a  project-a  partition-a, partition-b  this is reservation 1  3
		`),
			wantWideTable: new(`
ID  SIZE    PROJECT    PARTITIONS                DESCRIPTION            AMOUNT  LABELS
r1  size-a  project-a  partition-a, partition-b  this is reservation 1  3       a=b
		`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
r1 reservation 1
		`),
			wantMarkdown: new(`
| ID | SIZE   | PROJECT   | PARTITIONS               | DESCRIPTION           | AMOUNT |
|----|--------|-----------|--------------------------|-----------------------|--------|
| r1 | size-a | project-a | partition-a, partition-b | this is reservation 1 | 3      |
		`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1SizeReservationResponse) []string {
				return []string{"size", "reservation", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("DeleteSizeReservation", testcommon.MatchIgnoreContext(t, size.NewDeleteSizeReservationParams().WithID(*rv1.ID)), nil).Return(&size.DeleteSizeReservationOK{
						Payload: rv1,
					}, nil)
				},
			},
			want: rv1,
		},
		{
			name: "create from file",
			cmd: func(want *models.V1SizeReservationResponse) []string {
				return []string{"size", "reservation", "create", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1SizeReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("CreateSizeReservation", testcommon.MatchIgnoreContext(t, size.NewCreateSizeReservationParams().WithBody(sizeReservationResponseToCreate(rv1))), nil).Return(&size.CreateSizeReservationCreated{
						Payload: rv1,
					}, nil)
				},
			},
			want: rv1,
		},
		{
			name: "update from file",
			cmd: func(want *models.V1SizeReservationResponse) []string {
				return []string{"size", "reservation", "update", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1SizeReservationResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Size: func(mock *mock.Mock) {
					mock.On("UpdateSizeReservation", testcommon.MatchIgnoreContext(t, size.NewUpdateSizeReservationParams().WithBody(sizeReservationResponseToUpdate(rv1))), nil).Return(&size.UpdateSizeReservationOK{
						Payload: rv1,
					}, nil)
				},
			},
			want: rv1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
