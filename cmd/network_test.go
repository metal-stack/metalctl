package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/metal-stack/metal-go/api/client/network"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	network1 = &models.V1NetworkResponse{
		Description:         "network 1",
		Destinationprefixes: []string{"dest"},
		ID:                  pointer.Pointer("nw1"),
		Labels:              map[string]string{"a": "b"},
		Name:                "network-1",
		Nat:                 pointer.Pointer(true),
		Parentnetworkid:     "",
		Partitionid:         "partition-1",
		Prefixes:            []string{"prefix"},
		Privatesuper:        pointer.Pointer(true),
		Projectid:           "",
		Shared:              false,
		Underlay:            pointer.Pointer(true),
		Usage: &models.V1NetworkUsage{
			AvailableIps:      pointer.Pointer(int64(100)),
			AvailablePrefixes: pointer.Pointer(int64(200)),
			UsedIps:           pointer.Pointer(int64(300)),
			UsedPrefixes:      pointer.Pointer(int64(400)),
		},
		Vrf:       50,
		Vrfshared: true,
	}
	network1child = &models.V1NetworkResponse{
		Description:         "child 1",
		Destinationprefixes: []string{"dest"},
		ID:                  pointer.Pointer("child1"),
		Labels:              map[string]string{"e": "f"},
		Name:                "network-1",
		Nat:                 pointer.Pointer(true),
		Parentnetworkid:     "nw1",
		Partitionid:         "partition-1",
		Prefixes:            []string{"prefix"},
		Privatesuper:        pointer.Pointer(false),
		Projectid:           "project-1",
		Shared:              false,
		Underlay:            pointer.Pointer(false),
		Usage: &models.V1NetworkUsage{
			AvailableIps:      pointer.Pointer(int64(100)),
			AvailablePrefixes: pointer.Pointer(int64(200)),
			UsedIps:           pointer.Pointer(int64(300)),
			UsedPrefixes:      pointer.Pointer(int64(400)),
		},
		Vrf:       50,
		Vrfshared: true,
	}
	network2 = &models.V1NetworkResponse{
		Description:         "network 2",
		Destinationprefixes: []string{"dest"},
		ID:                  pointer.Pointer("nw2"),
		Labels:              map[string]string{"c": "d"},
		Name:                "network-2",
		Nat:                 pointer.Pointer(false),
		Parentnetworkid:     "internet",
		Partitionid:         "partition-1",
		Prefixes:            []string{"prefix"},
		Privatesuper:        pointer.Pointer(false),
		Projectid:           "project-1",
		Shared:              false,
		Underlay:            pointer.Pointer(false),
		Usage: &models.V1NetworkUsage{
			AvailableIps:      pointer.Pointer(int64(400)),
			AvailablePrefixes: pointer.Pointer(int64(300)),
			UsedIps:           pointer.Pointer(int64(200)),
			UsedPrefixes:      pointer.Pointer(int64(100)),
		},
		Vrf:       60,
		Vrfshared: true,
	}
)

func Test_NetworkCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1NetworkResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1NetworkResponse) []string {
				return []string{"network", "list"}
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("FindNetworks", testcommon.MatchIgnoreContext(t, network.NewFindNetworksParams().WithBody(&models.V1NetworkFindRequest{
						Destinationprefixes: []string{},
						Prefixes:            []string{},
					})), nil).Return(&network.FindNetworksOK{
						Payload: []*models.V1NetworkResponse{
							network2,
							network1,
							network1child,
						},
					}, nil)
				},
			},
			want: []*models.V1NetworkResponse{
				network1child,
				network1,
				network2,
			},
			wantTable: pointer.Pointer(`
ID          NAME        PROJECT     PARTITION     NAT     SHARED   PREFIXES       IPS
nw1         network-1               partition-1   true    false    prefix     ●    ●
└─╴child1   network-1   project-1   partition-1   true    false    prefix     ●    ●
nw2         network-2   project-1   partition-1   false   false    prefix          ●
`),
			wantWideTable: pointer.Pointer(`
ID          DESCRIPTION   NAME        PROJECT     PARTITION     NAT     SHARED   PREFIXES   USAGE              PRIVATESUPER   ANNOTATIONS
nw1         network 1     network-1               partition-1   true    false    prefix     IPs:     300/100   true           a=b
																							Prefixes:400/200
└─╴child1   child 1       network-1   project-1   partition-1   true    false    prefix     IPs:     300/100   false          e=f
																							Prefixes:400/200
nw2         network 2     network-2   project-1   partition-1   false   false    prefix     IPs:     200/400   false          c=d
																							Prefixes:100/300
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
child1 network-1
nw1 network-1
nw2 network-2
`),
			wantMarkdown: pointer.Pointer(`
|    ID     |   NAME    |  PROJECT  |  PARTITION  |  NAT  | SHARED | PREFIXES |   | IPS |
|-----------|-----------|-----------|-------------|-------|--------|----------|---|-----|
| nw1       | network-1 |           | partition-1 | true  | false  | prefix   | ● |  ●  |
| └─╴child1 | network-1 | project-1 | partition-1 | true  | false  | prefix   | ● |  ●  |
| nw2       | network-2 | project-1 | partition-1 | false | false  | prefix   |   |  ●  |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1NetworkResponse) []string {
				return appendFromFileCommonArgs("network", "apply")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1NetworkResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("CreateNetwork", testcommon.MatchIgnoreContext(t, network.NewCreateNetworkParams().WithBody(networkResponseToCreate(network1))), nil).Return(nil, &network.CreateNetworkConflict{}).Once()
					mock.On("UpdateNetwork", testcommon.MatchIgnoreContext(t, network.NewUpdateNetworkParams().WithBody(networkResponseToUpdate(network1))), nil).Return(&network.UpdateNetworkOK{
						Payload: network1,
					}, nil)
					mock.On("CreateNetwork", testcommon.MatchIgnoreContext(t, network.NewCreateNetworkParams().WithBody(networkResponseToCreate(network2))), nil).Return(&network.CreateNetworkCreated{
						Payload: network2,
					}, nil)
				},
			},
			want: []*models.V1NetworkResponse{
				network1,
				network2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1NetworkResponse) []string {
				return appendFromFileCommonArgs("network", "create")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1NetworkResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("CreateNetwork", testcommon.MatchIgnoreContext(t, network.NewCreateNetworkParams().WithBody(networkResponseToCreate(network1))), nil).Return(&network.CreateNetworkCreated{
						Payload: network1,
					}, nil)
				},
			},
			want: []*models.V1NetworkResponse{
				network1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1NetworkResponse) []string {
				return appendFromFileCommonArgs("network", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1NetworkResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("UpdateNetwork", testcommon.MatchIgnoreContext(t, network.NewUpdateNetworkParams().WithBody(networkResponseToUpdate(network1))), nil).Return(&network.UpdateNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			want: []*models.V1NetworkResponse{
				network1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1NetworkResponse) []string {
				return appendFromFileCommonArgs("network", "delete")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1NetworkResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("DeleteNetwork", testcommon.MatchIgnoreContext(t, network.NewDeleteNetworkParams().WithID(*network1.ID)), nil).Return(&network.DeleteNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			want: []*models.V1NetworkResponse{
				network1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_NetworkCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1NetworkResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1NetworkResponse) []string {
				return []string{"network", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("FindNetwork", testcommon.MatchIgnoreContext(t, network.NewFindNetworkParams().WithID(*network1.ID)), nil).Return(&network.FindNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			want: network1,
			wantTable: pointer.Pointer(`
ID    NAME        PROJECT   PARTITION     NAT    SHARED   PREFIXES       IPS
nw1   network-1             partition-1   true   false    prefix     ●    ●
		`),
			wantWideTable: pointer.Pointer(`
ID    DESCRIPTION   NAME        PROJECT   PARTITION     NAT    SHARED   PREFIXES   USAGE              PRIVATESUPER   ANNOTATIONS
nw1   network 1     network-1             partition-1   true   false    prefix     IPs:     300/100   true           a=b
																					Prefixes:400/200
		`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
nw1 network-1
		`),
			wantMarkdown: pointer.Pointer(`
| ID  |   NAME    | PROJECT |  PARTITION  | NAT  | SHARED | PREFIXES |   | IPS |
|-----|-----------|---------|-------------|------|--------|----------|---|-----|
| nw1 | network-1 |         | partition-1 | true | false  | prefix   | ● |  ●  |
		`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1NetworkResponse) []string {
				return []string{"network", "rm", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("DeleteNetwork", testcommon.MatchIgnoreContext(t, network.NewDeleteNetworkParams().WithID(*network1.ID)), nil).Return(&network.DeleteNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			want: network1,
		},
		{
			name: "create",
			cmd: func(want *models.V1NetworkResponse) []string {
				args := []string{"network", "create",
					"--id", *want.ID,
					"--name", want.Name,
					"--description", want.Description,
					"--labels", strings.Join(genericcli.MapToLabels(want.Labels), ","),
					"--partition", want.Partitionid,
					"--project", want.Projectid,
					"--prefixes", strings.Join(want.Prefixes, ","),
					"--destination-prefixes", strings.Join(want.Destinationprefixes, ","),
					"--privatesuper", strconv.FormatBool(*want.Privatesuper),
					"--nat", strconv.FormatBool(*want.Nat),
					"--underlay", strconv.FormatBool(*want.Underlay),
					"--vrf", strconv.FormatInt(want.Vrf, 10),
					"--vrfshared", strconv.FormatBool(want.Vrfshared),
				}
				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
				return args
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("CreateNetwork", testcommon.MatchIgnoreContext(t, network.NewCreateNetworkParams().WithBody(networkResponseToCreate(network1))), nil).Return(&network.CreateNetworkCreated{
						Payload: network1,
					}, nil)
				},
			},
			want: network1,
		},
		{
			name: "update",
			cmd: func(want *models.V1NetworkResponse) []string {
				args := []string{"network", "update", *want.ID,
					"--description", want.Description,
					"--add-prefixes", "prefix",
					"--remove-prefixes", "a,b",
					"--add-destinationprefixes", "dest",
					"--remove-destinationprefixes", "y,z",
					fmt.Sprintf("--shared=%t", want.Shared),
					"--labels", "a=b",
					"--name", want.Name,
				}
				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
				return args
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					networkToUpdate := mustJsonDeepCopy(t, network1)
					networkToUpdate.Prefixes = []string{"a", "b"}
					networkToUpdate.Destinationprefixes = []string{"y", "z"}
					networkToUpdate.Description = "Old description"

					mock.On("FindNetwork", testcommon.MatchIgnoreContext(t, network.NewFindNetworkParams().WithID(*network1.ID)), nil).Return(&network.FindNetworkOK{
						Payload: networkToUpdate,
					}, nil)
					mock.On("UpdateNetwork", testcommon.MatchIgnoreContext(t, network.NewUpdateNetworkParams().WithBody(&models.V1NetworkUpdateRequest{
						ID:                  network1.ID,
						Name:                network1.Name,
						Description:         network1.Description,
						Destinationprefixes: network1.Destinationprefixes,
						Prefixes:            network1.Prefixes,
						Labels:              network1.Labels,
						Shared:              network1.Shared,
					})), nil).Return(&network.UpdateNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			want: network1,
		},
		{
			name: "cannot remove unpresent prefix",
			cmd: func(want *models.V1NetworkResponse) []string {
				args := []string{"network", "update", *network1.ID,
					"--remove-prefixes", "x,y,z",
				}
				return args
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("FindNetwork", testcommon.MatchIgnoreContext(t, network.NewFindNetworkParams().WithID(*network1.ID)), nil).Return(&network.FindNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			wantErr: fmt.Errorf("cannot remove prefixes because they are currently not present: [x y z]"),
		},
		{
			name: "cannot add prefix that is already present",
			cmd: func(want *models.V1NetworkResponse) []string {
				args := []string{"network", "update", *network1.ID,
					"--add-prefixes", strings.Join(network1.Prefixes, ","),
				}
				return args
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("FindNetwork", testcommon.MatchIgnoreContext(t, network.NewFindNetworkParams().WithID(*network1.ID)), nil).Return(&network.FindNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			wantErr: fmt.Errorf("cannot add prefixes because they are already present: %s", network1.Prefixes),
		},
		{
			name: "cannot remove unpresent destination prefix",
			cmd: func(want *models.V1NetworkResponse) []string {
				args := []string{"network", "update", *network1.ID,
					"--remove-destinationprefixes", "x,y,z",
				}
				return args
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("FindNetwork", testcommon.MatchIgnoreContext(t, network.NewFindNetworkParams().WithID(*network1.ID)), nil).Return(&network.FindNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			wantErr: fmt.Errorf("cannot remove destination prefixes because they are currently not present: [x y z]"),
		},
		{
			name: "cannot add destination prefix that is already present",
			cmd: func(want *models.V1NetworkResponse) []string {
				args := []string{"network", "update", *network1.ID,
					"--add-destinationprefixes", strings.Join(network1.Destinationprefixes, ","),
				}
				return args
			},
			mocks: &client.MetalMockFns{
				Network: func(mock *mock.Mock) {
					mock.On("FindNetwork", testcommon.MatchIgnoreContext(t, network.NewFindNetworkParams().WithID(*network1.ID)), nil).Return(&network.FindNetworkOK{
						Payload: network1,
					}, nil)
				},
			},
			wantErr: fmt.Errorf("cannot add destination prefixes because they are already present: %s", network1.Destinationprefixes),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
