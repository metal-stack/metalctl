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
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	machine1 = &models.V1MachineResponse{
		Allocation: &models.V1MachineAllocation{
			BootInfo: &models.V1BootInfo{
				Bootloaderid: new("bootloaderid"),
				Cmdline:      new("cmdline"),
				ImageID:      new("imageid"),
				Initrd:       new("initrd"),
				Kernel:       new("kernel"),
				OsPartition:  new("ospartition"),
				PrimaryDisk:  new("primarydisk"),
			},
			Created:          new(strfmt.DateTime(testTime.Add(-14 * 24 * time.Hour))),
			Creator:          new("creator"),
			Description:      "machine allocation 1",
			Filesystemlayout: fsl1,
			Hostname:         new("machine-hostname-1"),
			Image:            image1,
			Name:             new("machine-1"),
			Networks: []*models.V1MachineNetwork{
				{
					Asn:                 new(int64(200)),
					Destinationprefixes: []string{"2.2.2.2"},
					Ips:                 []string{"1.1.1.1"},
					Nat:                 new(false),
					Networkid:           new("private"),
					Networktype:         pointer.Pointer(net.PrivatePrimaryUnshared),
					Prefixes:            []string{"prefixes"},
					Private:             new(true),
					Underlay:            new(false),
					Vrf:                 new(int64(100)),
				},
			},
			Project:    new("project-1"),
			Reinstall:  new(false),
			Role:       pointer.Pointer(models.V1MachineAllocationRoleMachine),
			SSHPubKeys: []string{"sshpubkey"},
			Succeeded:  new(true),
			UserData:   "---userdata---",
			DNSServers: []*models.V1DNSServer{{IP: new("8.8.8.8")}},
			NtpServers: []*models.V1NTPServer{{Address: new("1.pool.ntp.org")}},
		},
		Bios: &models.V1MachineBIOS{
			Date:    new("biosdata"),
			Vendor:  new("biosvendor"),
			Version: new("biosversion"),
		},
		Description: "machine 1",
		Events: &models.V1MachineRecentProvisioningEvents{
			CrashLoop:            new(false),
			FailedMachineReclaim: new(false),
			LastErrorEvent: &models.V1MachineProvisioningEvent{
				Event:   new("Crashed"),
				Message: "crash",
				Time:    strfmt.DateTime(testTime.Add(-10 * 24 * time.Hour)),
			},
			LastEventTime: strfmt.DateTime(testTime.Add(-7 * 24 * time.Hour)),
			Log: []*models.V1MachineProvisioningEvent{
				{
					Event:   new("Phoned Home"),
					Message: "phoning home",
					Time:    strfmt.DateTime(testTime.Add(-7 * 24 * time.Hour)),
				},
			},
		},
		Hardware: &models.V1MachineHardware{
			CPUCores: new(int32(16)),
			Disks:    []*models.V1MachineBlockDevice{},
			Memory:   new(int64(32)),
			Nics:     []*models.V1MachineNic{},
		},
		ID: new("1"),
		Ledstate: &models.V1ChassisIdentifyLEDState{
			Description: new(""),
			Value:       new(""),
		},
		Liveliness: new("Alive"),
		Name:       "machine-1",
		Partition:  partition1,
		Rackid:     "rack-1",
		Size:       size1,
		State: &models.V1MachineState{
			Description:        new("state"),
			Issuer:             "issuer",
			MetalHammerVersion: new("version"),
			Value:              new(""),
		},
		Tags: []string{"a"},
	}
	machine2 = &models.V1MachineResponse{
		Bios: &models.V1MachineBIOS{
			Date:    new("biosdata"),
			Vendor:  new("biosvendor"),
			Version: new("biosversion"),
		},
		Description: "machine 2",
		Events: &models.V1MachineRecentProvisioningEvents{
			CrashLoop:            new(false),
			FailedMachineReclaim: new(false),
			LastErrorEvent:       &models.V1MachineProvisioningEvent{},
			LastEventTime:        strfmt.DateTime(testTime.Add(-1 * time.Minute)),
			Log: []*models.V1MachineProvisioningEvent{
				{
					Event:   new("Waiting"),
					Message: "waiting",
					Time:    strfmt.DateTime{},
				},
			},
		},
		Hardware: &models.V1MachineHardware{
			CPUCores: new(int32(16)),
			Disks:    []*models.V1MachineBlockDevice{},
			Memory:   new(int64(32)),
			Nics:     []*models.V1MachineNic{},
		},
		ID: new("2"),
		Ledstate: &models.V1ChassisIdentifyLEDState{
			Description: new(""),
			Value:       new(""),
		},
		Liveliness: new("Alive"),
		Name:       "machine-2",
		Partition:  partition1,
		Rackid:     "rack-1",
		Size:       size1,
		State: &models.V1MachineState{
			Description:        new("state"),
			Issuer:             "issuer",
			MetalHammerVersion: new("version"),
			Value:              new(""),
		},
		Tags: []string{"b"},
	}
	ipmiMachine1 = &models.V1MachineIPMIResponse{
		Allocation: machine1.Allocation,
		Bios: &models.V1MachineBIOS{
			Version: new("2.0"),
		},
		Changed:     machine1.Changed,
		Created:     machine1.Created,
		Description: machine1.Description,
		Events:      machine1.Events,
		Hardware:    machine1.Hardware,
		ID:          machine1.ID,
		Ipmi: &models.V1MachineIPMI{
			Address:    new("1.2.3.4"),
			Bmcversion: new("1.1"),
			Fru: &models.V1MachineFru{
				BoardPartNumber:   "part123",
				ChassisPartSerial: "chassis123",
				ProductSerial:     "product123",
			},
			LastUpdated: new(strfmt.DateTime(testTime.Add(-5 * time.Second))),
			Mac:         new("1.2.3.4"),
			Powermetric: &models.V1PowerMetric{
				Averageconsumedwatts: new(float32(16.0)),
			},
			Powerstate: new("ON"),
		},
		Ledstate:   &models.V1ChassisIdentifyLEDState{},
		Liveliness: machine1.Liveliness,
		Name:       machine1.Name,
		Partition:  machine1.Partition,
		Rackid:     machine1.Rackid,
		Size:       machine1.Size,
		State:      machine1.State,
		Tags:       machine1.Tags,
	}
	ipmiMachine2 = &models.V1MachineIPMIResponse{
		Allocation: machine1.Allocation,
		Bios: &models.V1MachineBIOS{
			Version: new("2.0"),
		},
		Changed:     machine1.Changed,
		Created:     machine1.Created,
		Description: machine1.Description,
		Events:      machine1.Events,
		Hardware:    machine1.Hardware,
		ID:          machine1.ID,
		Ipmi: &models.V1MachineIPMI{
			Address:    new("1.2.3.4"),
			Bmcversion: new("1.1"),
			Fru: &models.V1MachineFru{
				BoardPartNumber:   "part123",
				ChassisPartSerial: "chassis123",
				ProductSerial:     "product123",
			},
			LastUpdated: new(strfmt.DateTime(testTime.Add(-5 * time.Second))),
			Mac:         new("1.2.3.4"),
			Powermetric: &models.V1PowerMetric{
				Averageconsumedwatts: new(float32(16.0)),
			},
			Powerstate: new("ON"),
			Powersupplies: []*models.V1PowerSupply{
				{Status: &models.V1PowerSupplyStatus{Health: new("OK")}},
				{Status: &models.V1PowerSupplyStatus{Health: new("NOT-OK")}},
			},
		},
		Ledstate:   &models.V1ChassisIdentifyLEDState{},
		Liveliness: machine1.Liveliness,
		Name:       machine1.Name,
		Partition:  machine1.Partition,
		Rackid:     machine1.Rackid,
		Size:       machine1.Size,
		State:      machine1.State,
		Tags:       machine1.Tags,
	}

	machineIssue1 = &models.V1MachineIssue{
		Description: new("this is a test issue 1"),
		Details:     new("more details 1"),
		ID:          new("issue-1-id"),
		RefURL:      new("https://url-1"),
		Severity:    new("minor"),
	}
	machineIssue2 = &models.V1MachineIssue{
		Description: new("this is a test issue 2"),
		Details:     new("more details 2"),
		ID:          new("issue-2-id"),
		RefURL:      new("https://url-2"),
		Severity:    new("major"),
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
						NicsMacAddresses:           nil,
						NetworkDestinationPrefixes: []string{},
						NetworkIps:                 []string{},
						NetworkIds:                 []string{},
						Tags:                       []string{},
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
			wantTable: new(`
ID    LAST EVENT   WHEN  AGE  HOSTNAME            PROJECT    SIZE  IMAGE        PARTITION  RACK
2     Waiting      1m                                        1                  1          rack-1
1     Phoned Home  7d    14d  machine-hostname-1  project-1  1     debian-name  1          rack-1
`),
			wantWideTable: new(`
ID  LAST EVENT   WHEN  AGE  DESCRIPTION           NAME       HOSTNAME            PROJECT    IPS      SIZE  IMAGE        PARTITION  RACK    STARTED               TAGS  LOCK / RESERVE  
2   Waiting      1m                                                                                  1                  1          rack-1                        b                   
1   Phoned Home  7d    14d  machine allocation 1  machine-1  machine-hostname-1  project-1  1.1.1.1  1     debian-name  1          rack-1  2022-05-05T01:02:03Z  a
`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
2 machine-2
1 machine-1
`),
			wantMarkdown: new(`
| ID |  | LAST EVENT  | WHEN | AGE | HOSTNAME           | PROJECT   | SIZE | IMAGE       | PARTITION | RACK   |
|----|--|-------------|------|-----|--------------------|-----------|------|-------------|-----------|--------|
| 2  |  | Waiting     | 1m   |     |                    |           | 1    |             | 1         | rack-1 |
| 1  |  | Phoned Home | 7d   | 14d | machine-hostname-1 | project-1 | 1    | debian-name | 1         | rack-1 |
`),
		},
		{
			name: "create from file",
			cmd: func(want []*models.V1MachineResponse) []string {
				return appendFromFileCommonArgs("machine", "create")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1MachineResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("AllocateMachine", testcommon.MatchIgnoreContext(t, machine.NewAllocateMachineParams().WithBody(machineResponseToCreate(machine1))), nil).Return(&machine.AllocateMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: []*models.V1MachineResponse{
				machine1,
			},
		},
		{
			name: "update from file",
			cmd: func(want []*models.V1MachineResponse) []string {
				return appendFromFileCommonArgs("machine", "update")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1MachineResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("UpdateMachine", testcommon.MatchIgnoreContext(t, machine.NewUpdateMachineParams().WithBody(machineResponseToUpdate(machine1))), nil).Return(&machine.UpdateMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: []*models.V1MachineResponse{
				machine1,
			},
		},
		{
			name: "delete from file",
			cmd: func(want []*models.V1MachineResponse) []string {
				return appendFromFileCommonArgs("machine", "delete")
			},
			fsMocks: func(fs afero.Fs, want []*models.V1MachineResponse) {
				require.NoError(t, afero.WriteFile(fs, "/file.yaml", mustMarshalToMultiYAML(t, want), 0755))
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("FreeMachine", testcommon.MatchIgnoreContext(t, machine.NewFreeMachineParams().WithID(*machine1.ID)), nil).Return(&machine.FreeMachineOK{
						Payload: machine1,
					}, nil)
				},
			},
			want: []*models.V1MachineResponse{
				machine1,
			},
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
			wantTable: new(`
		ID    LAST EVENT   WHEN  AGE  HOSTNAME            PROJECT    SIZE  IMAGE        PARTITION  RACK
		1     Phoned Home  7d    14d  machine-hostname-1  project-1  1     debian-name  1          rack-1
		`),
			wantWideTable: new(`
		ID  LAST EVENT   WHEN  AGE  DESCRIPTION           NAME       HOSTNAME            PROJECT    IPS      SIZE  IMAGE        PARTITION  RACK    STARTED               TAGS  LOCK / RESERVE
		1   Phoned Home  7d    14d  machine allocation 1  machine-1  machine-hostname-1  project-1  1.1.1.1  1     debian-name  1          rack-1  2022-05-05T01:02:03Z  a
		`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
		1 machine-1
		`),
			wantMarkdown: new(`
		| ID |  | LAST EVENT  | WHEN | AGE | HOSTNAME           | PROJECT   | SIZE | IMAGE       | PARTITION | RACK   |
		|----|--|-------------|------|-----|--------------------|-----------|------|-------------|-----------|--------|
		| 1  |  | Phoned Home | 7d   | 14d | machine-hostname-1 | project-1 | 1    | debian-name | 1         | rack-1 |
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
					ips        []string
					networks   []string
					dnsServers []string
					ntpservers []string
				)
				for _, s := range want.Allocation.Networks {
					ips = append(ips, s.Ips...)
					networks = append(networks, *s.Networkid+":noauto")
				}
				for _, dns := range want.Allocation.DNSServers {
					dnsServers = append(dnsServers, *dns.IP)
				}
				for _, ntp := range want.Allocation.NtpServers {
					ntpservers = append(ntpservers, *ntp.Address)
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
					"--dnsservers", strings.Join(dnsServers, ","),
					"--ntpservers", strings.Join(ntpservers, ","),
				}
				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
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
				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
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
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_MachineIPMICmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1MachineIPMIResponse]{
		{
			name: "machine ipmi",
			cmd: func(want []*models.V1MachineIPMIResponse) []string {
				return []string{"machine", "ipmi"}
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("FindIPMIMachines", testcommon.MatchIgnoreContext(t, machine.NewFindIPMIMachinesParams().WithBody(&models.V1MachineFindRequest{
						NicsMacAddresses:           nil,
						NetworkDestinationPrefixes: []string{},
						NetworkIps:                 []string{},
						NetworkIds:                 []string{},
						Tags:                       []string{},
					})), nil).Return(&machine.FindIPMIMachinesOK{
						Payload: []*models.V1MachineIPMIResponse{
							ipmiMachine1,
						},
					}, nil)
				},
			},
			want: []*models.V1MachineIPMIResponse{
				ipmiMachine1,
			},
			wantTable: new(`
ID    POWER  IP       MAC      BOARD PART NUMBER  BIOS  BMC  SIZE  PARTITION  RACK    UPDATED
1     ⏻ 16W  1.2.3.4  1.2.3.4  part123            2.0   1.1  1     1          rack-1  5s ago
`),
			wantWideTable: new(`
ID  LAST EVENT   STATUS  POWER   IP       MAC      BOARD PART NUMBER  CHASSIS SERIAL  PRODUCT SERIAL  BIOS VERSION  BMC VERSION  SIZE  PARTITION  RACK    UPDATED
1   Phoned Home          ON 16W  1.2.3.4  1.2.3.4  part123            chassis123      product123      2.0           1.1          1     1          rack-1  5s ago
`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
1 machine-1
`),
			wantMarkdown: new(`
| ID |  | POWER | IP      | MAC     | BOARD PART NUMBER | BIOS | BMC | SIZE | PARTITION | RACK   | UPDATED |
|----|--|-------|---------|---------|-------------------|------|-----|------|-----------|--------|---------|
| 1  |  | ⏻ 16W | 1.2.3.4 | 1.2.3.4 | part123           | 2.0  | 1.1 | 1    | 1         | rack-1 | 5s ago  |
`),
		},
		{
			name: "machine ipmi with broken powersupply",
			cmd: func(want []*models.V1MachineIPMIResponse) []string {
				return []string{"machine", "ipmi"}
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("FindIPMIMachines", testcommon.MatchIgnoreContext(t, machine.NewFindIPMIMachinesParams().WithBody(&models.V1MachineFindRequest{
						NicsMacAddresses:           nil,
						NetworkDestinationPrefixes: []string{},
						NetworkIps:                 []string{},
						NetworkIds:                 []string{},
						Tags:                       []string{},
					})), nil).Return(&machine.FindIPMIMachinesOK{
						Payload: []*models.V1MachineIPMIResponse{
							ipmiMachine2,
						},
					}, nil)
				},
			},
			want: []*models.V1MachineIPMIResponse{
				ipmiMachine2,
			},
			wantTable: new(`
ID    POWER  IP       MAC      BOARD PART NUMBER  BIOS  BMC  SIZE  PARTITION  RACK    UPDATED
1     ⏻ 16W  1.2.3.4  1.2.3.4  part123            2.0   1.1  1     1          rack-1  5s ago
`),
			wantWideTable: new(`
ID  LAST EVENT   STATUS  POWER                       IP       MAC      BOARD PART NUMBER  CHASSIS SERIAL  PRODUCT SERIAL  BIOS VERSION  BMC VERSION  SIZE  PARTITION  RACK    UPDATED
1   Phoned Home          ON Power Supply NOT-OK 16W  1.2.3.4  1.2.3.4  part123            chassis123      product123      2.0           1.1          1     1          rack-1  5s ago
`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
1 machine-1
`),
			wantMarkdown: new(`
| ID |  | POWER | IP      | MAC     | BOARD PART NUMBER | BIOS | BMC | SIZE | PARTITION | RACK   | UPDATED |
|----|--|-------|---------|---------|-------------------|------|-----|------|-----------|--------|---------|
| 1  |  | ⏻ 16W | 1.2.3.4 | 1.2.3.4 | part123           | 2.0  | 1.1 | 1    | 1         | rack-1 | 5s ago  |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_MachineIssuesListCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1MachineIssue]{
		{
			name: "issues list",
			cmd: func(want []*models.V1MachineIssue) []string {
				return []string{"machine", "issues", "list"}
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("ListIssues", testcommon.MatchIgnoreContext(t, machine.NewListIssuesParams()), nil).Return(&machine.ListIssuesOK{
						Payload: []*models.V1MachineIssue{
							machineIssue1,
							machineIssue2,
						},
					}, nil)
				},
			},
			want: []*models.V1MachineIssue{
				machineIssue2,
				machineIssue1,
			},
			wantTable: new(`
ID          SEVERITY  DESCRIPTION             REFERENCE URL
issue-2-id  major     this is a test issue 2  https://url-2
issue-1-id  minor     this is a test issue 1  https://url-1
`),
			wantWideTable: new(`
ID          SEVERITY  DESCRIPTION             REFERENCE URL
issue-2-id  major     this is a test issue 2  https://url-2
issue-1-id  minor     this is a test issue 1  https://url-1
`),
			template: new("{{ .id }}"),
			wantTemplate: new(`
			issue-2-id
issue-1-id
`),
			wantMarkdown: new(`
| ID         | SEVERITY | DESCRIPTION            | REFERENCE URL |
|------------|----------|------------------------|---------------|
| issue-2-id | major    | this is a test issue 2 | https://url-2 |
| issue-1-id | minor    | this is a test issue 1 | https://url-1 |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_MachineIssuesCmd(t *testing.T) {
	machineWithIssues := &tableprinters.MachinesAndIssues{
		EvaluationResult: []*models.V1MachineIssueResponse{
			{
				Machineid: machine1.ID,
				Issues: []string{
					pointer.SafeDeref(machineIssue1.ID),
					pointer.SafeDeref(machineIssue2.ID),
				},
			},
		},
		Issues: []*models.V1MachineIssue{
			machineIssue1,
			machineIssue2,
		},
		Machines: []*models.V1MachineIPMIResponse{
			ipmiMachine1,
		},
	}

	tests := []*test[*tableprinters.MachinesAndIssues]{
		{
			name: "issues",
			cmd: func(want *tableprinters.MachinesAndIssues) []string {
				return []string{"machine", "issues"}
			},
			mocks: &client.MetalMockFns{
				Machine: func(mock *mock.Mock) {
					mock.On("Issues", testcommon.MatchIgnoreContext(t, machine.NewIssuesParams().WithBody(&models.V1MachineIssuesRequest{
						Omit: []string{},
						Only: []string{},

						NicsMacAddresses:           nil,
						NetworkDestinationPrefixes: []string{},
						NetworkIps:                 []string{},
						NetworkIds:                 []string{},
						Tags:                       []string{},
					})), nil).Return(&machine.IssuesOK{
						Payload: machineWithIssues.EvaluationResult,
					}, nil)
					mock.On("ListIssues", testcommon.MatchIgnoreContext(t, machine.NewListIssuesParams()), nil).Return(&machine.ListIssuesOK{
						Payload: machineWithIssues.Issues,
					}, nil)
					mock.On("FindIPMIMachines", testcommon.MatchIgnoreContext(t, machine.NewFindIPMIMachinesParams().WithBody(&models.V1MachineFindRequest{
						NicsMacAddresses:           nil,
						NetworkDestinationPrefixes: []string{},
						NetworkIps:                 []string{},
						NetworkIds:                 []string{},
						Tags:                       []string{},
					})), nil).Return(&machine.FindIPMIMachinesOK{
						Payload: machineWithIssues.Machines,
					}, nil)
				},
			},
			want: machineWithIssues,
			wantTable: new(`
ID  POWER  ALLOCATED    LOCK REASON  LAST EVENT   WHEN  ISSUES
1   ⏻ 16W  yes          state        Phoned Home  7d    this is a test issue 1 (issue-1-id)
							this is a test issue 2 (issue-2-id)
`),
			wantWideTable: new(`
ID  NAME       PARTITION  PROJECT    POWER   STATE  LOCK REASON  LAST EVENT   WHEN  ISSUES                               REF URL        DETAILS
1   machine-1  1          project-1  ON 16W         state        Phoned Home  7d    this is a test issue 1 (issue-1-id)  https://url-1  more details 1
																					this is a test issue 2 (issue-2-id)  https://url-2  more details 2

`),
			wantMarkdown: new(`
| ID | POWER | ALLOCATED |  | LOCK REASON | LAST EVENT  | WHEN | ISSUES                              |
|----|-------|-----------|--|-------------|-------------|------|-------------------------------------|
| 1  | ⏻ 16W | yes       |  | state       | Phoned Home | 7d   | this is a test issue 1 (issue-1-id) |
|    |       |           |  |             |             |      | this is a test issue 2 (issue-2-id) |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
