package cmd

import (
	"testing"

	fsmodel "github.com/metal-stack/metal-go/api/client/filesystemlayout"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	fsl1 = &models.V1FilesystemLayoutResponse{
		Constraints: &models.V1FilesystemLayoutConstraints{
			Images: map[string]string{
				"os-image": "*",
			},
			Sizes: []string{"size1"},
		},
		Description: "fsl 1",
		Disks: []*models.V1Disk{
			{
				Device: pointer.Pointer("/dev/sda"),
				Partitions: []*models.V1DiskPartition{
					{
						Gpttype: pointer.Pointer("ef00"),
						Label:   "efi",
						Number:  pointer.Pointer(int64(1)),
						Size:    pointer.Pointer(int64(1000)),
					},
				},
				Wipeonreinstall: pointer.Pointer(true),
			},
		},
		Filesystems: []*models.V1Filesystem{
			{
				Createoptions: []string{"-F 32"},
				Device:        pointer.Pointer("/dev/sda1"),
				Format:        pointer.Pointer("vfat"),
				Label:         "efi",
				Mountoptions:  []string{"noexec"},
				Path:          "/boot/efi",
			},
			{
				Createoptions: []string{},
				Device:        pointer.Pointer("tmpfs"),
				Format:        pointer.Pointer("tmpfs"),
				Label:         "",
				Mountoptions:  []string{"noexec"},
				Path:          "/tmp",
			},
		},
		ID: pointer.Pointer("1"),
		Logicalvolumes: []*models.V1LogicalVolume{
			{
				Lvmtype:     pointer.Pointer("linear"),
				Name:        pointer.Pointer("varlib"),
				Size:        pointer.Pointer(int64(5000)),
				Volumegroup: pointer.Pointer("lvm"),
			},
		},
		Name: "fsl1",
		Raid: []*models.V1Raid{},
		Volumegroups: []*models.V1VolumeGroup{
			{
				Devices: []string{"/dev/nvme0n1"},
				Name:    pointer.Pointer("lvm"),
				Tags:    []string{},
			},
		},
	}
	fsl2 = &models.V1FilesystemLayoutResponse{
		Constraints: &models.V1FilesystemLayoutConstraints{
			Images: map[string]string{
				"os-image": "*",
			},
			Sizes: []string{"size1"},
		},
		Description: "fsl 2",
		Disks: []*models.V1Disk{
			{
				Device: pointer.Pointer("/dev/sda"),
				Partitions: []*models.V1DiskPartition{
					{
						Gpttype: pointer.Pointer("ef00"),
						Label:   "efi",
						Number:  pointer.Pointer(int64(1)),
						Size:    pointer.Pointer(int64(1000)),
					},
				},
				Wipeonreinstall: pointer.Pointer(true),
			},
		},
		Filesystems: []*models.V1Filesystem{
			{
				Createoptions: []string{},
				Device:        pointer.Pointer("tmpfs"),
				Format:        pointer.Pointer("tmpfs"),
				Label:         "",
				Mountoptions:  []string{"noexec"},
				Path:          "/tmp",
			},
		},
		ID: pointer.Pointer("2"),
		Logicalvolumes: []*models.V1LogicalVolume{
			{
				Lvmtype:     pointer.Pointer("linear"),
				Name:        pointer.Pointer("varlib"),
				Size:        pointer.Pointer(int64(5000)),
				Volumegroup: pointer.Pointer("lvm"),
			},
		},
		Name: "fsl2",
		Raid: []*models.V1Raid{},
		Volumegroups: []*models.V1VolumeGroup{
			{
				Devices: []string{"/dev/nvme0n1"},
				Name:    pointer.Pointer("lvm"),
				Tags:    []string{},
			},
		},
	}
)

func Test_FilesystemLayoutCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1FilesystemLayoutResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1FilesystemLayoutResponse) []string {
				return []string{"fsl", "list"}
			},
			mocks: &client.MetalMockFns{
				Filesystemlayout: func(mock *mock.Mock) {
					mock.On("ListFilesystemLayouts", testcommon.MatchIgnoreContext(t, fsmodel.NewListFilesystemLayoutsParams()), nil).Return(&fsmodel.ListFilesystemLayoutsOK{
						Payload: []*models.V1FilesystemLayoutResponse{
							fsl2,
							fsl1,
						},
					}, nil)
				},
			},
			want: []*models.V1FilesystemLayoutResponse{
				fsl1,
				fsl2,
			},
			wantTable: pointer.Pointer(`
ID   DESCRIPTION   FILESYSTEMS            SIZES   IMAGES
1    fsl 1         /tmp       tmpfs       size1   os-image *
                   /boot/efi  /dev/sda1
2    fsl 2         /tmp  tmpfs            size1   os-image *
`),
			wantWideTable: pointer.Pointer(`
ID   DESCRIPTION   FILESYSTEMS            SIZES   IMAGES
1    fsl 1         /tmp       tmpfs       size1   os-image *
                   /boot/efi  /dev/sda1
2    fsl 2         /tmp  tmpfs            size1   os-image *
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 fsl1
2 fsl2
`),
			wantMarkdown: pointer.Pointer(`
| ID | DESCRIPTION |     FILESYSTEMS      | SIZES |   IMAGES   |
|----|-------------|----------------------|-------|------------|
|  1 | fsl 1       | /tmp       tmpfs     | size1 | os-image * |
|    |             | /boot/efi  /dev/sda1 |       |            |
|  2 | fsl 2       | /tmp  tmpfs          | size1 | os-image * |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1FilesystemLayoutResponse) []string {
				return []string{"fsl", "apply", "--bulk-output", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1FilesystemLayoutResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Filesystemlayout: func(mock *mock.Mock) {
					mock.On("CreateFilesystemLayout", testcommon.MatchIgnoreContext(t, fsmodel.NewCreateFilesystemLayoutParams().WithBody(filesystemLayoutResponseToCreate(fsl1))), nil).Return(nil, &fsmodel.CreateFilesystemLayoutConflict{}).Once()
					mock.On("UpdateFilesystemLayout", testcommon.MatchIgnoreContext(t, fsmodel.NewUpdateFilesystemLayoutParams().WithBody(filesystemLayoutResponseToUpdate(fsl1))), nil).Return(&fsmodel.UpdateFilesystemLayoutOK{
						Payload: fsl1,
					}, nil)
					mock.On("CreateFilesystemLayout", testcommon.MatchIgnoreContext(t, fsmodel.NewCreateFilesystemLayoutParams().WithBody(filesystemLayoutResponseToCreate(fsl2))), nil).Return(&fsmodel.CreateFilesystemLayoutCreated{
						Payload: fsl2,
					}, nil)
				},
			},
			want: []*models.V1FilesystemLayoutResponse{
				fsl1,
				fsl2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1FilesystemLayoutResponse) []string {
				return []string{"fsl", "create", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1FilesystemLayoutResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Filesystemlayout: func(mock *mock.Mock) {
					mock.On("CreateFilesystemLayout", testcommon.MatchIgnoreContext(t, fsmodel.NewCreateFilesystemLayoutParams().WithBody(filesystemLayoutResponseToCreate(fsl1))), nil).Return(&fsmodel.CreateFilesystemLayoutCreated{
						Payload: fsl1,
					}, nil)
				},
			},
			want: []*models.V1FilesystemLayoutResponse{
				fsl1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1FilesystemLayoutResponse) []string {
				return []string{"fsl", "update", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1FilesystemLayoutResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Filesystemlayout: func(mock *mock.Mock) {
					mock.On("UpdateFilesystemLayout", testcommon.MatchIgnoreContext(t, fsmodel.NewUpdateFilesystemLayoutParams().WithBody(filesystemLayoutResponseToUpdate(fsl1))), nil).Return(&fsmodel.UpdateFilesystemLayoutOK{
						Payload: fsl1,
					}, nil)
				},
			},
			want: []*models.V1FilesystemLayoutResponse{
				fsl1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1FilesystemLayoutResponse) []string {
				return []string{"fsl", "delete", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want []*models.V1FilesystemLayoutResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Filesystemlayout: func(mock *mock.Mock) {
					mock.On("DeleteFilesystemLayout", testcommon.MatchIgnoreContext(t, fsmodel.NewDeleteFilesystemLayoutParams().WithID(*fsl1.ID)), nil).Return(&fsmodel.DeleteFilesystemLayoutOK{
						Payload: fsl1,
					}, nil)
				},
			},
			want: []*models.V1FilesystemLayoutResponse{
				fsl1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_FilesystemLayoutCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1FilesystemLayoutResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1FilesystemLayoutResponse) []string {
				return []string{"fsl", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Filesystemlayout: func(mock *mock.Mock) {
					mock.On("GetFilesystemLayout", testcommon.MatchIgnoreContext(t, fsmodel.NewGetFilesystemLayoutParams().WithID(*fsl1.ID)), nil).Return(&fsmodel.GetFilesystemLayoutOK{
						Payload: fsl1,
					}, nil)
				},
			},
			want: fsl1,
			wantTable: pointer.Pointer(`
ID   DESCRIPTION   FILESYSTEMS            SIZES   IMAGES
1    fsl 1         /tmp       tmpfs       size1   os-image *
					/boot/efi  /dev/sda1
		`),
			wantWideTable: pointer.Pointer(`
ID   DESCRIPTION   FILESYSTEMS            SIZES   IMAGES
1    fsl 1         /tmp       tmpfs       size1   os-image *
                   /boot/efi  /dev/sda1
		`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 fsl1
		`),
			wantMarkdown: pointer.Pointer(`
| ID | DESCRIPTION |     FILESYSTEMS      | SIZES |   IMAGES   |
|----|-------------|----------------------|-------|------------|
|  1 | fsl 1       | /tmp       tmpfs     | size1 | os-image * |
|    |             | /boot/efi  /dev/sda1 |       |            |
		`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1FilesystemLayoutResponse) []string {
				return []string{"fsl", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Filesystemlayout: func(mock *mock.Mock) {
					mock.On("DeleteFilesystemLayout", testcommon.MatchIgnoreContext(t, fsmodel.NewDeleteFilesystemLayoutParams().WithID(*fsl1.ID)), nil).Return(&fsmodel.DeleteFilesystemLayoutOK{
						Payload: fsl1,
					}, nil)
				},
			},
			want: fsl1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
