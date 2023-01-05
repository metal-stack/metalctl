package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/net"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	machine1 = &models.V1MachineResponse{
		Allocation: &models.V1MachineAllocation{
			BootInfo: &models.V1BootInfo{
				Bootloaderid: pointer.Pointer("bootloaderid"),
				Cmdline:      pointer.Pointer("cmdline"),
				ImageID:      pointer.Pointer("imageid"),
				Initrd:       pointer.Pointer("initrd"),
				Kernel:       pointer.Pointer("kernel"),
				OsPartition:  pointer.Pointer("ospartition"),
				PrimaryDisk:  pointer.Pointer("primarydisk"),
			},
			Created:          pointer.Pointer(strfmt.DateTime(testTime.Add(-14 * 24 * time.Hour))),
			Creator:          pointer.Pointer("creator"),
			Description:      "machine allocation 1",
			Filesystemlayout: fsl1,
			Hostname:         pointer.Pointer("machine-hostname-1"),
			Image:            image1,
			Name:             pointer.Pointer("machine-1"),
			Networks: []*models.V1MachineNetwork{
				{
					Asn:                 pointer.Pointer(int64(200)),
					Destinationprefixes: []string{"2.2.2.2"},
					Ips:                 []string{"1.1.1.1"},
					Nat:                 pointer.Pointer(false),
					Networkid:           pointer.Pointer("private"),
					Networktype:         pointer.Pointer(net.PrivatePrimaryUnshared),
					Prefixes:            []string{"prefixes"},
					Private:             pointer.Pointer(true),
					Underlay:            pointer.Pointer(false),
					Vrf:                 pointer.Pointer(int64(100)),
				},
			},
			Project:    pointer.Pointer("project-1"),
			Reinstall:  pointer.Pointer(false),
			Role:       pointer.Pointer(models.V1MachineAllocationRoleMachine),
			SSHPubKeys: []string{"sshpubkey"},
			Succeeded:  pointer.Pointer(true),
			UserData:   "---userdata---",
		},
		Bios: &models.V1MachineBIOS{
			Date:    pointer.Pointer("biosdata"),
			Vendor:  pointer.Pointer("biosvendor"),
			Version: pointer.Pointer("biosversion"),
		},
		Description: "machine 1",
		Events: &models.V1MachineRecentProvisioningEvents{
			CrashLoop:            pointer.Pointer(false),
			FailedMachineReclaim: pointer.Pointer(false),
			LastErrorEvent: &models.V1MachineProvisioningEvent{
				Event:   pointer.Pointer("Crashed"),
				Message: "crash",
				Time:    strfmt.DateTime(testTime.Add(-10 * 24 * time.Hour)),
			},
			LastEventTime: strfmt.DateTime(testTime.Add(-7 * 24 * time.Hour)),
			Log: []*models.V1MachineProvisioningEvent{
				{
					Event:   pointer.Pointer("Phoned Home"),
					Message: "phoning home",
					Time:    strfmt.DateTime(testTime.Add(-7 * 24 * time.Hour)),
				},
			},
		},
		Hardware: &models.V1MachineHardware{
			CPUCores: pointer.Pointer(int32(16)),
			Disks:    []*models.V1MachineBlockDevice{},
			Memory:   pointer.Pointer(int64(32)),
			Nics:     []*models.V1MachineNic{},
		},
		ID: pointer.Pointer("1"),
		Ledstate: &models.V1ChassisIdentifyLEDState{
			Description: pointer.Pointer(""),
			Value:       pointer.Pointer(""),
		},
		Liveliness: pointer.Pointer("Alive"),
		Name:       "machine-1",
		Partition:  partition1,
		Rackid:     "rack-1",
		Size:       size1,
		State: &models.V1MachineState{
			Description:        pointer.Pointer("state"),
			Issuer:             "issuer",
			MetalHammerVersion: pointer.Pointer("version"),
			Value:              pointer.Pointer(""),
		},
		Tags: []string{"a"},
	}
	machine2 = &models.V1MachineResponse{
		Bios: &models.V1MachineBIOS{
			Date:    pointer.Pointer("biosdata"),
			Vendor:  pointer.Pointer("biosvendor"),
			Version: pointer.Pointer("biosversion"),
		},
		Description: "machine 2",
		Events: &models.V1MachineRecentProvisioningEvents{
			CrashLoop:            pointer.Pointer(false),
			FailedMachineReclaim: pointer.Pointer(false),
			LastErrorEvent:       &models.V1MachineProvisioningEvent{},
			LastEventTime:        strfmt.DateTime(testTime.Add(-1 * time.Minute)),
			Log: []*models.V1MachineProvisioningEvent{
				{
					Event:   pointer.Pointer("Waiting"),
					Message: "waiting",
					Time:    strfmt.DateTime{},
				},
			},
		},
		Hardware: &models.V1MachineHardware{
			CPUCores: pointer.Pointer(int32(16)),
			Disks:    []*models.V1MachineBlockDevice{},
			Memory:   pointer.Pointer(int64(32)),
			Nics:     []*models.V1MachineNic{},
		},
		ID: pointer.Pointer("2"),
		Ledstate: &models.V1ChassisIdentifyLEDState{
			Description: pointer.Pointer(""),
			Value:       pointer.Pointer(""),
		},
		Liveliness: pointer.Pointer("Alive"),
		Name:       "machine-2",
		Partition:  partition1,
		Rackid:     "rack-1",
		Size:       size1,
		State: &models.V1MachineState{
			Description:        pointer.Pointer("state"),
			Issuer:             "issuer",
			MetalHammerVersion: pointer.Pointer("version"),
			Value:              pointer.Pointer(""),
		},
		Tags: []string{"b"},
	}
)

func Test_MachineCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1MachineResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1MachineResponse) []string {
				return []string{"machine", "list"}
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("FindMachines", testcommon.MatchIgnoreContext(t, machine.NewFindMachinesParams().WithBody(&models.V1MachineFindRequest{
						NicsMacAddresses: []string{},
						Tags:             []string{},
					})), nil).Return(&machine.FindMachinesOK{
						Payload: []*models.V1MachineResponse{
							machine1,
							machine2,
						},
					}, nil)
				},
			},
			want: []*models.V1MachineResponse{
				machine2,
				machine1,
			},
			wantTable: pointer.Pointer(`
ID      LAST EVENT    WHEN   AGE   HOSTNAME             PROJECT     SIZE   IMAGE         PARTITION   RACK
2       Waiting       1m                                            1                    1           rack-1
1       Phoned Home   7d     14d   machine-hostname-1   project-1   1      debian-name   1           rack-1
`),
			wantWideTable: pointer.Pointer(`
ID   LAST EVENT    WHEN   AGE   DESCRIPTION            NAME        HOSTNAME             PROJECT     IPS       SIZE   IMAGE         PARTITION   STARTED                TAGS   LOCK/RESERVE
2    Waiting       1m                                                                                         1                    1                                  b
1    Phoned Home   7d     14d   machine allocation 1   machine-1   machine-hostname-1   project-1   1.1.1.1   1      debian-name   1           2022-05-05T01:02:03Z   a
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
2 machine-2
1 machine-1
`),
			wantMarkdown: pointer.Pointer(`
| ID |  | LAST EVENT  | WHEN | AGE |      HOSTNAME      |  PROJECT  | SIZE |    IMAGE    | PARTITION |  RACK  |
|----|--|-------------|------|-----|--------------------|-----------|------|-------------|-----------|--------|
|  2 |  | Waiting     | 1m   |     |                    |           |    1 |             |         1 | rack-1 |
|  1 |  | Phoned Home | 7d   | 14d | machine-hostname-1 | project-1 |    1 | debian-name |         1 | rack-1 |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_MachineCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1MachineResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1MachineResponse) []string {
				return []string{"machine", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("FindMachine", testcommon.MatchIgnoreContext(t, machine.NewFindMachineParams().WithID(*machine1.ID)), nil).Return(&machine.FindMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: machine1,
			wantTable: pointer.Pointer(`
ID      LAST EVENT    WHEN   AGE   HOSTNAME             PROJECT     SIZE   IMAGE         PARTITION   RACK
1       Phoned Home   7d     14d   machine-hostname-1   project-1   1      debian-name   1           rack-1
`),
			wantWideTable: pointer.Pointer(`
ID   LAST EVENT    WHEN   AGE   DESCRIPTION            NAME        HOSTNAME             PROJECT     IPS       SIZE   IMAGE         PARTITION   STARTED                TAGS   LOCK/RESERVE
1    Phoned Home   7d     14d   machine allocation 1   machine-1   machine-hostname-1   project-1   1.1.1.1   1      debian-name   1           2022-05-05T01:02:03Z   a
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 machine-1
`),
			wantMarkdown: pointer.Pointer(`
| ID |  | LAST EVENT  | WHEN | AGE |      HOSTNAME      |  PROJECT  | SIZE |    IMAGE    | PARTITION |  RACK  |
|----|--|-------------|------|-----|--------------------|-----------|------|-------------|-----------|--------|
|  1 |  | Phoned Home | 7d   | 14d | machine-hostname-1 | project-1 |    1 | debian-name |         1 | rack-1 |
`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1MachineResponse) []string {
				return []string{"machine", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("FreeMachine", testcommon.MatchIgnoreContext(t, machine.NewFreeMachineParams().WithID(*machine1.ID)), nil).Return(&machine.FreeMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: machine1,
		},
		{
			name: "create",
			cmd: func(want *models.V1MachineResponse) []string {
				var (
					ips      []string
					networks []string
				)
				for _, s := range want.Allocation.Networks {
					ips = append(ips, s.Ips...)
					networks = append(networks, *s.Networkid+":noauto")
				}

				args := []string{"machine", "create",
					"--id", *want.ID,
					"--name", want.Name,
					"--description", want.Allocation.Description,
					"--filesystemlayout", *want.Allocation.Filesystemlayout.ID,
					"--hostname", *want.Allocation.Hostname,
					"--image", *want.Allocation.Image.ID,
					"--ips", strings.Join(ips, ","),
					"--networks", strings.Join(networks, ","),
					"--partition", *want.Partition.ID,
					"--project", *want.Allocation.Project,
					"--size", *want.Size.ID,
					"--sshpublickey", pointer.FirstOrZero(want.Allocation.SSHPubKeys),
					"--tags", strings.Join(want.Tags, ","),
					"--userdata", want.Allocation.UserData,
				}
				assertExhaustiveArgs(t, args, "file")
				return args
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("AllocateMachine", testcommon.MatchIgnoreContext(t, machine.NewAllocateMachineParams().WithBody(machineResponseToCreate(machine1))), nil).Return(&machine.AllocateMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: machine1,
		},
		{
			name: "update",
			cmd: func(want *models.V1MachineResponse) []string {
				args := []string{"machine", "update", *want.ID,
					"--description", want.Allocation.Description,
					"--add-tags", strings.Join(want.Tags, ","),
					"--remove-tags", "z",
				}
				assertExhaustiveArgs(t, args, "file")
				return args
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					machineToUpdate := machine1
					machineToUpdate.Tags = []string{"z"}
					mock.On("FindMachine", testcommon.MatchIgnoreContext(t, machine.NewFindMachineParams().WithID(*machine1.ID)), nil).Return(&machine.FindMachineOK{
						Payload: machineToUpdate,
					}, nil)
					mock.On("UpdateMachine", testcommon.MatchIgnoreContext(t, machine.NewUpdateMachineParams().WithBody(machineResponseToUpdate(machine1))), nil).Return(&machine.UpdateMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: machine1,
		},
		{
			name: "create from file",
			cmd: func(want *models.V1MachineResponse) []string {
				return []string{"machine", "create", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1MachineResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("AllocateMachine", testcommon.MatchIgnoreContext(t, machine.NewAllocateMachineParams().WithBody(machineResponseToCreate(machine1))), nil).Return(&machine.AllocateMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: machine1,
		},
		{
			name: "update from file",
			cmd: func(want *models.V1MachineResponse) []string {
				return []string{"machine", "update", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1MachineResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("UpdateMachine", testcommon.MatchIgnoreContext(t, machine.NewUpdateMachineParams().WithBody(machineResponseToUpdate(machine1))), nil).Return(&machine.UpdateMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: machine1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
