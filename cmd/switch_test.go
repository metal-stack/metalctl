package cmd

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/metal-stack/metal-go/api/client/switch_operations"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
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
					Mac:  pointer.Pointer("a-mac"),
					Name: pointer.Pointer("a-name"),
					Vrf:  "100",
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
				Mac:  pointer.Pointer("a-mac"),
				Name: pointer.Pointer("a-name"),
				Vrf:  "100",
			},
		},
		Partition: partition1,
		RackID:    pointer.Pointer("rack-1"),
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
					Mac:  pointer.Pointer("a-mac"),
					Name: pointer.Pointer("a-name"),
					Vrf:  "100",
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
				Mac:  pointer.Pointer("a-mac"),
				Name: pointer.Pointer("a-name"),
				Vrf:  "100",
			},
		},
		Partition: partition1,
		RackID:    pointer.Pointer("rack-1"),
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
					mock.On("ListSwitches", testcommon.MatchIgnoreContext(t, switch_operations.NewListSwitchesParams()), nil).Return(&switch_operations.ListSwitchesOK{
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
ID   PARTITION   RACK     STATUS
1    1           rack-1    ●
2    1           rack-1    ●
`),
			wantWideTable: pointer.Pointer(`
ID   PARTITION   RACK     MODE          LAST SYNC   SYNC DURATION   LAST SYNC ERROR
1    1           rack-1   operational   0s          1s              5m ago: error
2    1           rack-1   operational   0s          1s              5m ago: error
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 switch-1
2 switch-2
`),
			wantMarkdown: pointer.Pointer(`
| ID | PARTITION |  RACK  | STATUS |
|----|-----------|--------|--------|
|  1 |         1 | rack-1 |  ●     |
|  2 |         1 | rack-1 |  ●     |
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
					mock.On("DeleteSwitch", testcommon.MatchIgnoreContext(t, switch_operations.NewDeleteSwitchParams().WithID(*switch1.ID)), nil).Return(&switch_operations.DeleteSwitchOK{
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
ID   PARTITION   RACK     STATUS
1    1           rack-1    ●
		`),
			wantWideTable: pointer.Pointer(`
ID   PARTITION   RACK     MODE          LAST SYNC   SYNC DURATION   LAST SYNC ERROR
1    1           rack-1   operational   0s          1s              5m ago: error
					`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 switch-1
		`),
			wantMarkdown: pointer.Pointer(`
| ID | PARTITION |  RACK  | STATUS |
|----|-----------|--------|--------|
|  1 |         1 | rack-1 |  ●     |
		`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1SwitchResponse) []string {
				return []string{"switch", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				SwitchOperations: func(mock *mock.Mock) {
					mock.On("DeleteSwitch", testcommon.MatchIgnoreContext(t, switch_operations.NewDeleteSwitchParams().WithID(*switch1.ID)), nil).Return(&switch_operations.DeleteSwitchOK{
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
