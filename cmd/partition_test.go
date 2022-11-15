package cmd

import (
	"testing"

	"github.com/metal-stack/metal-go/api/client/partition"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	partition1 = &models.V1PartitionResponse{
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
	}
	partition2 = &models.V1PartitionResponse{
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
	}
)

func Test_PartitionCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1PartitionResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1PartitionResponse) []string {
				return []string{"partition", "list"}
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("ListPartitions", testcommon.MatchIgnoreContext(t, partition.NewListPartitionsParams()), nil).Return(&partition.ListPartitionsOK{
						Payload: []*models.V1PartitionResponse{
							partition2,
							partition1,
						},
					}, nil)
				},
			},
			want: []*models.V1PartitionResponse{
				partition1,
				partition2,
			},
			wantTable: pointer.Pointer(`
ID   NAME          DESCRIPTION
1    partition-1   partition 1
2    partition-2   partition 2
`),
			wantWideTable: pointer.Pointer(`
ID   NAME          DESCRIPTION
1    partition-1   partition 1
2    partition-2   partition 2
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 partition-1
2 partition-2
`),
			wantMarkdown: pointer.Pointer(`
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
|  2 | partition-2 | partition 2 |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1PartitionResponse) []string {
				return []string{"partition", "apply", "--bulk-output", "-f", "/file.yaml", "--force"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1PartitionResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("CreatePartition", testcommon.MatchIgnoreContext(t, partition.NewCreatePartitionParams().WithBody(partitionResponseToCreate(partition1))), nil).Return(nil, &partition.CreatePartitionConflict{}).Once()
					mock.On("UpdatePartition", testcommon.MatchIgnoreContext(t, partition.NewUpdatePartitionParams().WithBody(partitionResponseToUpdate(partition1))), nil).Return(&partition.UpdatePartitionOK{
						Payload: partition1,
					}, nil)
					mock.On("CreatePartition", testcommon.MatchIgnoreContext(t, partition.NewCreatePartitionParams().WithBody(partitionResponseToCreate(partition2))), nil).Return(&partition.CreatePartitionCreated{
						Payload: partition2,
					}, nil)
				},
			},
			want: []*models.V1PartitionResponse{
				partition1,
				partition2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1PartitionResponse) []string {
				return []string{"partition", "create", "-f", "/file.yaml", "--force"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1PartitionResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("CreatePartition", testcommon.MatchIgnoreContext(t, partition.NewCreatePartitionParams().WithBody(partitionResponseToCreate(partition1))), nil).Return(&partition.CreatePartitionCreated{
						Payload: partition1,
					}, nil)
				},
			},
			want: []*models.V1PartitionResponse{
				partition1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1PartitionResponse) []string {
				return []string{"partition", "update", "-f", "/file.yaml", "--force"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1PartitionResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("UpdatePartition", testcommon.MatchIgnoreContext(t, partition.NewUpdatePartitionParams().WithBody(partitionResponseToUpdate(partition1))), nil).Return(&partition.UpdatePartitionOK{
						Payload: partition1,
					}, nil)
				},
			},
			want: []*models.V1PartitionResponse{
				partition1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1PartitionResponse) []string {
				return []string{"partition", "delete", "-f", "/file.yaml", "--force"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1PartitionResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("DeletePartition", testcommon.MatchIgnoreContext(t, partition.NewDeletePartitionParams().WithID(*partition1.ID)), nil).Return(&partition.DeletePartitionOK{
						Payload: partition1,
					}, nil)
				},
			},
			want: []*models.V1PartitionResponse{
				partition1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_PartitionCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1PartitionResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1PartitionResponse) []string {
				return []string{"partition", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("FindPartition", testcommon.MatchIgnoreContext(t, partition.NewFindPartitionParams().WithID(*partition1.ID)), nil).Return(&partition.FindPartitionOK{
						Payload: partition1,
					}, nil)
				},
			},
			want: partition1,
			wantTable: pointer.Pointer(`
ID   NAME          DESCRIPTION
1    partition-1   partition 1
`),
			wantWideTable: pointer.Pointer(`
ID   NAME          DESCRIPTION
1    partition-1   partition 1
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 partition-1
`),
			wantMarkdown: pointer.Pointer(`
| ID |    NAME     | DESCRIPTION |
|----|-------------|-------------|
|  1 | partition-1 | partition 1 |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1PartitionResponse) []string {
				return []string{"partition", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("DeletePartition", testcommon.MatchIgnoreContext(t, partition.NewDeletePartitionParams().WithID(*partition1.ID)), nil).Return(&partition.DeletePartitionOK{
						Payload: partition1,
					}, nil)
				},
			},
			want: partition1,
		},
		{
			name: "create",
			cmd: func(want *models.V1PartitionResponse) []string {
				args := []string{"partition", "create",
					"--id", *want.ID,
					"--name", want.Name,
					"--description", want.Description,
					"--cmdline", want.Bootconfig.Commandline,
					"--kernelurl", want.Bootconfig.Kernelurl,
					"--imageurl", want.Bootconfig.Imageurl,
					"--mgmtserver", want.Mgmtserviceaddress,
				}
				assertExhaustiveArgs(t, args, "file", "bulk-output", "force")
				return args
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					p := partition1
					p.Privatenetworkprefixlength = 0
					mock.On("CreatePartition", testcommon.MatchIgnoreContext(t, partition.NewCreatePartitionParams().WithBody(partitionResponseToCreate(p))), nil).Return(&partition.CreatePartitionCreated{
						Payload: partition1,
					}, nil)
				},
			},
			want: partition1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_PartitionCapacityCmd(t *testing.T) {
	tests := []*test[[]*models.V1PartitionCapacity]{
		{
			name: "capacity",
			cmd: func(want []*models.V1PartitionCapacity) []string {
				return []string{"partition", "capacity"}
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("PartitionCapacity", testcommon.MatchIgnoreContext(t, partition.NewPartitionCapacityParams().WithBody(&models.V1PartitionCapacityRequest{})), nil).Return(&partition.PartitionCapacityOK{
						Payload: []*models.V1PartitionCapacity{
							{
								Description: "partition 1",
								ID:          pointer.Pointer("1"),
								Name:        "partition-1",
								Servers: []*models.V1ServerCapacity{
									{
										Allocated:      pointer.Pointer(int32(1)),
										Faulty:         pointer.Pointer(int32(2)),
										Faultymachines: []string{"abc"},
										Free:           pointer.Pointer(int32(3)),
										Other:          pointer.Pointer(int32(4)),
										Othermachines:  []string{"def"},
										Size:           pointer.Pointer("size-1"),
										Total:          pointer.Pointer(int32(5)),
									},
								},
							},
						},
					}, nil)
				},
			},
			want: []*models.V1PartitionCapacity{
				{
					Description: "partition 1",
					ID:          pointer.Pointer("1"),
					Name:        "partition-1",
					Servers: []*models.V1ServerCapacity{
						{
							Allocated:      pointer.Pointer(int32(1)),
							Faulty:         pointer.Pointer(int32(2)),
							Faultymachines: []string{"abc"},
							Free:           pointer.Pointer(int32(3)),
							Other:          pointer.Pointer(int32(4)),
							Othermachines:  []string{"def"},
							Size:           pointer.Pointer("size-1"),
							Total:          pointer.Pointer(int32(5)),
						},
					},
				},
			},
			wantTable: pointer.Pointer(`
PARTITION   SIZE     TOTAL   FREE   ALLOCATED   OTHER   FAULTY
1           size-1   5       3      1           4       2
Total                5       3      1           4       2
`),
			wantWideTable: pointer.Pointer(`
PARTITION   SIZE     TOTAL   FREE   ALLOCATED   OTHER   FAULTY
1           size-1   5       3      1           def     abc
Total                5       3      1           4       2
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 partition-1
`),
			wantMarkdown: pointer.Pointer(`
| PARTITION |  SIZE  | TOTAL | FREE | ALLOCATED | OTHER | FAULTY |
|-----------|--------|-------|------|-----------|-------|--------|
|         1 | size-1 |     5 |    3 |         1 |     4 |      2 |
| Total     |        |     5 |    3 |         1 |     4 |      2 |
`),
		},
		{
			name: "capacity with filters",
			cmd: func(want []*models.V1PartitionCapacity) []string {
				args := []string{"partition", "capacity", "--id", "1", "--size", "size-1"}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &client.MetalMockFns{
				Partition: func(mock *mock.Mock) {
					mock.On("PartitionCapacity", testcommon.MatchIgnoreContext(t, partition.NewPartitionCapacityParams().WithBody(&models.V1PartitionCapacityRequest{
						ID:     "1",
						Sizeid: "size-1",
					})), nil).Return(&partition.PartitionCapacityOK{
						Payload: []*models.V1PartitionCapacity{
							{
								Description: "partition 1",
								ID:          pointer.Pointer("1"),
								Name:        "partition-1",
								Servers: []*models.V1ServerCapacity{
									{
										Allocated:      pointer.Pointer(int32(1)),
										Faulty:         pointer.Pointer(int32(2)),
										Faultymachines: []string{"abc"},
										Free:           pointer.Pointer(int32(3)),
										Other:          pointer.Pointer(int32(4)),
										Othermachines:  []string{"def"},
										Size:           pointer.Pointer("size-1"),
										Total:          pointer.Pointer(int32(5)),
									},
								},
							},
						},
					}, nil)
				},
			},
			want: []*models.V1PartitionCapacity{
				{
					Description: "partition 1",
					ID:          pointer.Pointer("1"),
					Name:        "partition-1",
					Servers: []*models.V1ServerCapacity{
						{
							Allocated:      pointer.Pointer(int32(1)),
							Faulty:         pointer.Pointer(int32(2)),
							Faultymachines: []string{"abc"},
							Free:           pointer.Pointer(int32(3)),
							Other:          pointer.Pointer(int32(4)),
							Othermachines:  []string{"def"},
							Size:           pointer.Pointer("size-1"),
							Total:          pointer.Pointer(int32(5)),
						},
					},
				},
			},
			wantTable: pointer.Pointer(`
PARTITION   SIZE     TOTAL   FREE   ALLOCATED   OTHER   FAULTY
1           size-1   5       3      1           4       2
Total                5       3      1           4       2
`),
			wantWideTable: pointer.Pointer(`
PARTITION   SIZE     TOTAL   FREE   ALLOCATED   OTHER   FAULTY
1           size-1   5       3      1           def     abc
Total                5       3      1           4       2
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 partition-1
`),
			wantMarkdown: pointer.Pointer(`
| PARTITION |  SIZE  | TOTAL | FREE | ALLOCATED | OTHER | FAULTY |
|-----------|--------|-------|------|-----------|-------|--------|
|         1 | size-1 |     5 |    3 |         1 |     4 |      2 |
| Total     |        |     5 |    3 |         1 |     4 |      2 |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
