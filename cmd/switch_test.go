package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/metal-stack/metal-go/api/client/machine"
	"github.com/metal-stack/metal-go/api/client/switch_operations"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	switch1 = &models.V1SwitchResponse{
		Connections: []*models.V1SwitchConnection{
			{
				MachineID: "machine-1",
				Nic: &models.V1SwitchNic{
					Filter: &models.V1BGPFilter{
						Cidrs: []string{"cidr"},
						Vnis:  []string{"vni"},
					},
					Mac:    pointer.Pointer("a-mac"),
					Name:   pointer.Pointer("a-name"),
					Vrf:    "100",
					Actual: pointer.Pointer("UP"),
				},
			},
		},
		Description: "switch 1",
		ID:          pointer.Pointer("1"),
		LastSync: &models.V1SwitchSync{
			Duration: pointer.Pointer(int64(1 * time.Second)),
			Error:    "",
			Time:     pointer.Pointer(strfmt.DateTime(testTime)),
		},
		LastSyncError: &models.V1SwitchSync{
			Duration: pointer.Pointer(int64(2 * time.Second)),
			Error:    "error",
			Time:     pointer.Pointer(strfmt.DateTime(testTime.Add(-5 * time.Minute))),
		},
		Mode: "operational",
		Name: "switch-1",
		Nics: []*models.V1SwitchNic{
			{
				Filter: &models.V1BGPFilter{
					Cidrs: []string{"cidr"},
					Vnis:  []string{"vni"},
				},
				Mac:    pointer.Pointer("a-mac"),
				Name:   pointer.Pointer("a-name"),
				Vrf:    "100",
				Actual: pointer.Pointer("UP"),
			},
		},
		Partition: partition1,
		RackID:    pointer.Pointer("rack-1"),
		Os: &models.V1SwitchOS{
			Vendor:           "SONiC",
			Version:          "1",
			MetalCoreVersion: "1.2.3",
		},
		ManagementIP:   "1.2.3.4",
		ManagementUser: "root",
	}
	switch2 = &models.V1SwitchResponse{
		Connections: []*models.V1SwitchConnection{
			{
				MachineID: "machine-1",
				Nic: &models.V1SwitchNic{
					Filter: &models.V1BGPFilter{
						Cidrs: []string{"cidr"},
						Vnis:  []string{"vni"},
					},
					Mac:    pointer.Pointer("a-mac"),
					Name:   pointer.Pointer("a-name"),
					Vrf:    "100",
					Actual: pointer.Pointer("DOWN"),
				},
			},
		},
		Description: "switch 2",
		ID:          pointer.Pointer("2"),
		LastSync: &models.V1SwitchSync{
			Duration: pointer.Pointer(int64(1 * time.Second)),
			Error:    "",
			Time:     pointer.Pointer(strfmt.DateTime(testTime)),
		},
		LastSyncError: &models.V1SwitchSync{
			Duration: pointer.Pointer(int64(2 * time.Second)),
			Error:    "error",
			Time:     pointer.Pointer(strfmt.DateTime(testTime.Add(-5 * time.Minute))),
		},
		Mode: "operational",
		Name: "switch-2",
		Nics: []*models.V1SwitchNic{
			{
				Filter: &models.V1BGPFilter{
					Cidrs: []string{"cidr"},
					Vnis:  []string{"vni"},
				},
				Mac:    pointer.Pointer("a-mac"),
				Name:   pointer.Pointer("a-name"),
				Vrf:    "100",
				Actual: pointer.Pointer("UP"),
			},
		},
		Partition: partition1,
		RackID:    pointer.Pointer("rack-1"),
		Os: &models.V1SwitchOS{
			Vendor:  "Cumulus",
			Version: "2",
		},
	}
)

func Test_SwitchCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1SwitchResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1SwitchResponse) []string {
				return []string{"switch", "list"}
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("FindSwitches", testcommon.MatchIgnoreContext(t, switch_operations.NewFindSwitchesParams().WithBody(&models.V1SwitchFindRequest{})), nil).Return(&switch_operations.FindSwitchesOK{
						Payload: []*models.V1SwitchResponse{
							switch2,
							switch1,
						},
					}, nil)
				},
			},
			want: []*models.V1SwitchResponse{
				switch1,
				switch2,
			},
			wantTable: pointer.Pointer(`
ID  PARTITION  RACK    OS  STATUS  LAST SYNC
1   1          rack-1  🦔  ●       0s ago
2   1          rack-1  🐢  ●       0s ago
`),
			wantWideTable: pointer.Pointer(`
ID  PARTITION  RACK    OS           METALCORE  IP       MODE         LAST SYNC  SYNC DURATION  LAST ERROR
1   1          rack-1  SONiC (1)    1.2.3      1.2.3.4  operational  0s ago     1s             5m ago: error
2   1          rack-1  Cumulus (2)                      operational  0s ago     1s             5m ago: error
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 switch-1
2 switch-2
`),
			wantMarkdown: pointer.Pointer(`
| ID | PARTITION | RACK   | OS | STATUS | LAST SYNC |
|----|-----------|--------|----|--------|-----------|
| 1  | 1         | rack-1 | 🦔 | ●      | 0s ago    |
| 2  | 1         | rack-1 | 🐢 | ●      | 0s ago    |
`),
		},
		{
			name: "list with filters",
			cmd: func(want []*models.V1SwitchResponse) []string {
				args := []string{"switch", "list", "--id", *want[0].ID, "--name", want[0].Name, "--os-vendor", want[0].Os.Vendor, "--os-version", want[0].Os.Version, "--partition", *want[0].Partition.ID, "--rack", *want[0].RackID}
				assertExhaustiveArgs(t, args, "sort-by")
				return args
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("FindSwitches", testcommon.MatchIgnoreContext(t, switch_operations.NewFindSwitchesParams().WithBody(&models.V1SwitchFindRequest{
						ID:          *switch1.ID,
						Name:        switch1.Name,
						Osvendor:    switch1.Os.Vendor,
						Osversion:   switch1.Os.Version,
						Partitionid: *switch1.Partition.ID,
						Rackid:      *switch1.RackID,
					})), nil).Return(&switch_operations.FindSwitchesOK{
						Payload: []*models.V1SwitchResponse{
							switch1,
						},
					}, nil)
				},
			},
			want: []*models.V1SwitchResponse{
				switch1,
			},
			wantTable: pointer.Pointer(`
ID  PARTITION  RACK    OS  STATUS  LAST SYNC
1   1          rack-1  🦔  ●       0s ago
		`),
			wantWideTable: pointer.Pointer(`
ID  PARTITION  RACK    OS         METALCORE  IP       MODE         LAST SYNC  SYNC DURATION  LAST ERROR
1   1          rack-1  SONiC (1)  1.2.3      1.2.3.4  operational  0s ago     1s             5m ago: error
		`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 switch-1
		`),
			wantMarkdown: pointer.Pointer(`
| ID | PARTITION | RACK   | OS | STATUS | LAST SYNC |
|----|-----------|--------|----|--------|-----------|
| 1  | 1         | rack-1 | 🦔 | ●      | 0s ago    |
		`),
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1SwitchResponse) []string {
				return appendFromFileCommonArgs("switch", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SwitchResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("UpdateSwitch", testcommon.MatchIgnoreContext(t, switch_operations.NewUpdateSwitchParams().WithBody(switchResponseToUpdate(switch1))), nil).Return(&switch_operations.UpdateSwitchOK{
						Payload: switch1,
					}, nil)
				},
			},
			want: []*models.V1SwitchResponse{
				switch1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1SwitchResponse) []string {
				return appendFromFileCommonArgs("switch", "delete")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1SwitchResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("DeleteSwitch", testcommon.MatchIgnoreContext(t, switch_operations.NewDeleteSwitchParams().WithID(*switch1.ID).WithForce(pointer.Pointer(false))), nil).Return(&switch_operations.DeleteSwitchOK{
						Payload: switch1,
					}, nil)
				},
			},
			want: []*models.V1SwitchResponse{
				switch1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_SwitchCmd_ConnectedMachinesResult(t *testing.T) {
	tests := []*test[*tableprinters.SwitchesWithMachines]{
		{
			name: "connected-machines",
			cmd: func(want *tableprinters.SwitchesWithMachines) []string {
				return []string{"switch", "connected-machines"}
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("FindIPMIMachines", testcommon.MatchIgnoreContext(t, machine.NewFindIPMIMachinesParams().WithBody(&models.V1MachineFindRequest{})), nil).Return(&machine.FindIPMIMachinesOK{
						Payload: []*models.V1MachineIPMIResponse{
							{
								ID:     pointer.Pointer("machine-1"),
								Rackid: "rack-1",
								Allocation: &models.V1MachineAllocation{
									Hostname: pointer.Pointer("alloc-1"),
								},
								Partition: &models.V1PartitionResponse{
									ID: pointer.Pointer("1"),
								},
								Size: &models.V1SizeResponse{
									ID: pointer.Pointer("n1-medium-x86"),
								},
								Ipmi: &models.V1MachineIPMI{
									Fru: &models.V1MachineFru{
										ProductSerial:     "123",
										ChassisPartSerial: "456",
									},
								},
							},
						},
					}, nil)
				},
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("FindSwitches", testcommon.MatchIgnoreContext(t, switch_operations.NewFindSwitchesParams().WithBody(&models.V1SwitchFindRequest{})), nil).Return(&switch_operations.FindSwitchesOK{
						Payload: []*models.V1SwitchResponse{
							switch2,
							switch1,
						},
					}, nil)
				},
			},
			want: &tableprinters.SwitchesWithMachines{
				SS: []*models.V1SwitchResponse{
					switch1,
					switch2,
				},
				MS: map[string]*models.V1MachineIPMIResponse{
					"machine-1": {
						ID:     pointer.Pointer("machine-1"),
						Rackid: "rack-1",
						Allocation: &models.V1MachineAllocation{
							Hostname: pointer.Pointer("alloc-1"),
						},
						Partition: &models.V1PartitionResponse{
							ID: pointer.Pointer("1"),
						},
						Size: &models.V1SizeResponse{
							ID: pointer.Pointer("n1-medium-x86"),
						},
						Ipmi: &models.V1MachineIPMI{
							Fru: &models.V1MachineFru{
								ProductSerial:     "123",
								ChassisPartSerial: "456",
							},
						},
					},
				},
			},
			wantTable: pointer.Pointer(`
ID            NIC NAME       IDENTIFIER  PARTITION  RACK    SIZE           PRODUCT SERIAL  CHASSIS SERIAL
1                                        1          rack-1
└─╴machine-1  a-name         a-mac       1          rack-1  n1-medium-x86  123             456
2                                        1          rack-1
└─╴machine-1  a-name (DOWN)  a-mac       1          rack-1  n1-medium-x86  123             456
`),
			wantWideTable: pointer.Pointer(`
ID                NIC NAME       IDENTIFIER  PARTITION  RACK    SIZE           HOSTNAME  PRODUCT SERIAL  CHASSIS SERIAL
1                                            1          rack-1
└─╴machine-1  ❓  a-name         a-mac       1          rack-1  n1-medium-x86  alloc-1   123             456
2                                            1          rack-1
└─╴machine-1  ❓  a-name (DOWN)  a-mac       1          rack-1  n1-medium-x86  alloc-1   123             456
`),
			template: pointer.Pointer(`{{ $machines := .machines }}{{ range .switches }}{{ $switch := . }}{{ range .connections }}{{ $switch.id }},{{ $switch.rack_id }},{{ .nic.name }},{{ .machine_id }},{{ (index $machines .machine_id).ipmi.fru.product_serial }}{{ printf "\n" }}{{ end }}{{ end }}`),
			wantTemplate: pointer.Pointer(`
1,rack-1,a-name,machine-1,123
2,rack-1,a-name,machine-1,123
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_SwitchCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1SwitchResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1SwitchResponse) []string {
				return []string{"switch", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("FindSwitch", testcommon.MatchIgnoreContext(t, switch_operations.NewFindSwitchParams().WithID(*switch1.ID)), nil).Return(&switch_operations.FindSwitchOK{
						Payload: switch1,
					}, nil)
				},
			},
			want: switch1,
			wantTable: pointer.Pointer(`
ID  PARTITION  RACK    OS  STATUS  LAST SYNC
1   1          rack-1  🦔  ●       0s ago
		`),
			wantWideTable: pointer.Pointer(`
ID  PARTITION  RACK    OS         METALCORE  IP       MODE         LAST SYNC  SYNC DURATION  LAST ERROR
1   1          rack-1  SONiC (1)  1.2.3      1.2.3.4  operational  0s ago     1s             5m ago: error
					`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 switch-1
		`),
			wantMarkdown: pointer.Pointer(`
| ID | PARTITION | RACK   | OS | STATUS | LAST SYNC |
|----|-----------|--------|----|--------|-----------|
| 1  | 1         | rack-1 | 🦔 | ●      | 0s ago    |
		`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1SwitchResponse) []string {
				return []string{"switch", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("DeleteSwitch", testcommon.MatchIgnoreContext(t, switch_operations.NewDeleteSwitchParams().WithID(*switch1.ID).WithForce(pointer.Pointer(false))), nil).Return(&switch_operations.DeleteSwitchOK{
						Payload: switch1,
					}, nil)
				},
			},
			want: switch1,
		},
		{
			name: "update from file",
			cmd: func(want *models.V1SwitchResponse) []string {
				return []string{"switch", "update", "-f", "/file.yaml"}
			},
			fsMocks: func(fs afero.Fs, want *models.V1SwitchResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshal(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("UpdateSwitch", testcommon.MatchIgnoreContext(t, switch_operations.NewUpdateSwitchParams().WithBody(switchResponseToUpdate(switch1))), nil).Return(&switch_operations.UpdateSwitchOK{
						Payload: switch1,
					}, nil)
				},
			},
			want: switch1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_SwitchCmd_ToggleResult(t *testing.T) {
	// first create a full copy of the test data so we can change some data
	data, err := json.Marshal(switch1)
	require.NoError(t, err)
	var sw1Down models.V1SwitchResponse
	err = json.Unmarshal(data, &sw1Down)
	require.NoError(t, err)
	sw1Down.Nics[0].Actual = pointer.Pointer("DOWN")

	tests := []*test[currentSwitchPortStateDump]{
		{
			name: "query state",
			cmd: func(want currentSwitchPortStateDump) []string {
				return []string{"switch", "port", "describe", *switch1.ID, "--port", *switch1.Nics[0].Name}
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("FindSwitch", testcommon.MatchIgnoreContext(t, switch_operations.NewFindSwitchParams().WithID(*switch1.ID)), nil).Return(&switch_operations.FindSwitchOK{
						Payload: switch1,
					}, nil)
				},
			},
			want: currentSwitchPortStateDump{
				Actual:  *switch1.Connections[0],
				Desired: *switch1.Nics[0],
			},
		},
		{
			name: "toggle down",
			cmd: func(want currentSwitchPortStateDump) []string {
				return []string{"switch", "port", "down", *switch1.ID, "--port", *switch1.Nics[0].Name}
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {

					mock.On("ToggleSwitchPort",
						testcommon.MatchIgnoreContext(t,
							switch_operations.NewToggleSwitchPortParams().
								WithID(*switch1.ID).
								WithBody(&models.V1SwitchPortToggleRequest{
									Nic:    switch1.Nics[0].Name,
									Status: pointer.Pointer("DOWN"),
								})), nil).Return(&switch_operations.ToggleSwitchPortOK{
						Payload: &sw1Down,
					}, nil)
				},
			},
			want: currentSwitchPortStateDump{
				Actual:  *sw1Down.Connections[0],
				Desired: *sw1Down.Nics[0],
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
