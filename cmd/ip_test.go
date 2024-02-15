package cmd

import (
	"strings"
	"testing"

	"github.com/metal-stack/metal-go/api/client/ip"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	ip1 = &models.V1IPResponse{
		Allocationuuid: pointer.Pointer("1"),
		Description:    "ip 1",
		Ipaddress:      pointer.Pointer("1.1.1.1"),
		Name:           "ip-1",
		Networkid:      pointer.Pointer("internet"),
		Projectid:      pointer.Pointer("project-1"),
		Tags:           []string{"a"},
		Type:           pointer.Pointer(models.V1IPAllocateRequestTypeEphemeral),
	}
	ip2 = &models.V1IPResponse{
		Allocationuuid: pointer.Pointer("2"),
		Description:    "ip 2",
		Ipaddress:      pointer.Pointer("2.2.2.2"),
		Name:           "ip-2",
		Networkid:      pointer.Pointer("internet"),
		Projectid:      pointer.Pointer("project-2"),
		Tags:           []string{"b"},
		Type:           pointer.Pointer(models.V1IPAllocateRequestTypeStatic),
	}
)

func Test_IPCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1IPResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1IPResponse) []string {
				return []string{"network", "ip", "list"}
			},
			mocks: &client.MetalMockFns{
				IP: func(mock *mock.Mock) {
					mock.On("FindIPs", testcommon.MatchIgnoreContext(t, ip.NewFindIPsParams().WithBody(&models.V1IPFindRequest{
						Tags: []string{},
					})), nil).Return(&ip.FindIPsOK{
						Payload: []*models.V1IPResponse{
							ip2,
							ip1,
						},
					}, nil)
				},
			},
			want: []*models.V1IPResponse{
				ip1,
				ip2,
			},
			wantTable: pointer.Pointer(`
IP        DESCRIPTION   NAME   NETWORK    PROJECT     TYPE        TAGS
1.1.1.1   ip 1          ip-1   internet   project-1   ephemeral   a
2.2.2.2   ip 2          ip-2   internet   project-2   static      b
`),
			wantWideTable: pointer.Pointer(`
IP        ALLOCATION UUID   DESCRIPTION   NAME   NETWORK    PROJECT     TYPE        TAGS
1.1.1.1   1                 ip 1          ip-1   internet   project-1   ephemeral   a
2.2.2.2   2                 ip 2          ip-2   internet   project-2   static      b
`),
			template: pointer.Pointer("{{ .ipaddress }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1.1.1.1 ip-1
2.2.2.2 ip-2
`),
			wantMarkdown: pointer.Pointer(`
|   IP    | DESCRIPTION | NAME | NETWORK  |  PROJECT  |   TYPE    | TAGS |
|---------|-------------|------|----------|-----------|-----------|------|
| 1.1.1.1 | ip 1        | ip-1 | internet | project-1 | ephemeral | a    |
| 2.2.2.2 | ip 2        | ip-2 | internet | project-2 | static    | b    |
`),
		},
		{
			name: "apply",
			cmd: func(want []*models.V1IPResponse) []string {
				return appendFromFileCommonArgs("network", "ip", "apply")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1IPResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				IP: func(mock *mock.Mock) {
					mock.On("AllocateSpecificIP", testcommon.MatchIgnoreContext(t, ip.NewAllocateSpecificIPParams().WithBody(ipResponseToCreate(ip1).V1IPAllocateRequest).WithIP(*ip1.Ipaddress)), nil).Return(nil, ip.NewAllocateSpecificIPConflict()).Once()
					mock.On("UpdateIP", testcommon.MatchIgnoreContext(t, ip.NewUpdateIPParams().WithBody(ipResponseToUpdate(ip1))), nil).Return(&ip.UpdateIPOK{
						Payload: ip1,
					}, nil)
					mock.On("AllocateSpecificIP", testcommon.MatchIgnoreContext(t, ip.NewAllocateSpecificIPParams().WithBody(ipResponseToCreate(ip2).V1IPAllocateRequest).WithIP(*ip2.Ipaddress)), nil).Return(&ip.AllocateSpecificIPCreated{
						Payload: ip2,
					}, nil)
				},
			},
			want: []*models.V1IPResponse{
				ip1,
				ip2,
			},
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1IPResponse) []string {
				return appendFromFileCommonArgs("network", "ip", "create")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1IPResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				IP: func(mock *mock.Mock) {
					mock.On("AllocateSpecificIP", testcommon.MatchIgnoreContext(t, ip.NewAllocateSpecificIPParams().WithBody(ipResponseToCreate(ip1).V1IPAllocateRequest).WithIP(*ip1.Ipaddress)), nil).Return(&ip.AllocateSpecificIPCreated{
						Payload: ip1,
					}, nil)
				},
			},
			want: []*models.V1IPResponse{
				ip1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1IPResponse) []string {
				return appendFromFileCommonArgs("network", "ip", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1IPResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				IP: func(mock *mock.Mock) {
					mock.On("UpdateIP", testcommon.MatchIgnoreContext(t, ip.NewUpdateIPParams().WithBody(ipResponseToUpdate(ip1))), nil).Return(&ip.UpdateIPOK{
						Payload: ip1,
					}, nil)
				},
			},
			want: []*models.V1IPResponse{
				ip1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1IPResponse) []string {
				return appendFromFileCommonArgs("network", "ip", "delete")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1IPResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				IP: func(mock *mock.Mock) {
					mock.On("FreeIP", testcommon.MatchIgnoreContext(t, ip.NewFreeIPParams().WithID(*ip1.Ipaddress)), nil).Return(&ip.FreeIPOK{
						Payload: ip1,
					}, nil)
				},
			},
			want: []*models.V1IPResponse{
				ip1,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_IPCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1IPResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1IPResponse) []string {
				return []string{"network", "ip", "describe", *want.Ipaddress}
			},
			mocks: &client.MetalMockFns{
				IP: func(mock *mock.Mock) {
					mock.On("FindIP", testcommon.MatchIgnoreContext(t, ip.NewFindIPParams().WithID(*ip1.Ipaddress)), nil).Return(&ip.FindIPOK{
						Payload: ip1,
					}, nil)
				},
			},
			want: ip1,
			wantTable: pointer.Pointer(`
IP        DESCRIPTION   NAME   NETWORK    PROJECT     TYPE        TAGS
1.1.1.1   ip 1          ip-1   internet   project-1   ephemeral   a
		`),
			wantWideTable: pointer.Pointer(`
IP        ALLOCATION UUID   DESCRIPTION   NAME   NETWORK    PROJECT     TYPE        TAGS
1.1.1.1   1                 ip 1          ip-1   internet   project-1   ephemeral   a
		`),
			template: pointer.Pointer("{{ .ipaddress }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1.1.1.1 ip-1
		`),
			wantMarkdown: pointer.Pointer(`
|   IP    | DESCRIPTION | NAME | NETWORK  |  PROJECT  |   TYPE    | TAGS |
|---------|-------------|------|----------|-----------|-----------|------|
| 1.1.1.1 | ip 1        | ip-1 | internet | project-1 | ephemeral | a    |
		`),
		},
		{
			name: "delete",
			cmd: func(want *models.V1IPResponse) []string {
				return []string{"network", "ip", "rm", *want.Ipaddress}
			},
			mocks: &client.MetalMockFns{
				IP: func(mock *mock.Mock) {
					mock.On("FreeIP", testcommon.MatchIgnoreContext(t, ip.NewFreeIPParams().WithID(*ip1.Ipaddress)), nil).Return(&ip.FreeIPOK{
						Payload: ip1,
					}, nil)
				},
			},
			want: ip1,
		},
		{
			name: "create",
			cmd: func(want *models.V1IPResponse) []string {
				args := []string{"network", "ip", "create",
					"--ipaddress", *want.Ipaddress,
					"--name", want.Name,
					"--description", want.Description,
					"--network", *want.Networkid,
					"--project", *want.Projectid,
					"--type", *want.Type,
					"--tags", strings.Join(want.Tags, ","),
				}
				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
				return args
			},
			mocks: &client.MetalMockFns{
				IP: func(mock *mock.Mock) {
					mock.On("AllocateSpecificIP", testcommon.MatchIgnoreContext(t, ip.NewAllocateSpecificIPParams().WithBody(ipResponseToCreate(ip1).V1IPAllocateRequest).WithIP(*ip1.Ipaddress)), nil).Return(&ip.AllocateSpecificIPCreated{
						Payload: ip1,
					}, nil)
				},
			},
			want: ip1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
