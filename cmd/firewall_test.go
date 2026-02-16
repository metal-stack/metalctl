package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/metal-stack/metal-go/api/client/firewall"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/net"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/stretchr/testify/mock"
)

var (
	firewall1 = &models.V1FirewallResponse{
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
			Description:      "firewall allocation 1",
			Filesystemlayout: fsl1,
			Hostname:         new("firewall-hostname-1"),
			Image:            image1,
			Name:             new("firewall-1"),
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
			Role:       pointer.Pointer(models.V1MachineAllocationRoleFirewall),
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
		Description: "firewall 1",
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
		Name:       "firewall-1",
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
	firewall2 = &models.V1FirewallResponse{
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
			Description:      "firewall allocation 2",
			Filesystemlayout: fsl1,
			Hostname:         new("firewall-hostname-2"),
			Image:            image1,
			Name:             new("firewall-2"),
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
			Role:       pointer.Pointer(models.V1MachineAllocationRoleFirewall),
			SSHPubKeys: []string{"sshpubkey"},
			Succeeded:  new(true),
			UserData:   "---userdata---",
		},
		Bios: &models.V1MachineBIOS{
			Date:    new("biosdata"),
			Vendor:  new("biosvendor"),
			Version: new("biosversion"),
		},
		Description: "firewall 2",
		Events: &models.V1MachineRecentProvisioningEvents{
			CrashLoop:            new(false),
			FailedMachineReclaim: new(false),
			LastErrorEvent:       &models.V1MachineProvisioningEvent{},
			LastEventTime:        strfmt.DateTime(testTime.Add(-1 * time.Minute)),
			Log: []*models.V1MachineProvisioningEvent{
				{
					Event:   new("Phoned Home"),
					Message: "phoning home",
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
		Name:       "firewall-2",
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
)

func Test_FirewallCmd_MultiResult(t *testing.T) {
	tests := []*test[[]*models.V1FirewallResponse]{
		{
			name: "list",
			cmd: func(want []*models.V1FirewallResponse) []string {
				return []string{"firewall", "list"}
			},
			mocks: &client.MetalMockFns{
				Firewall: func(mock *mock.Mock) {
					mock.On("FindFirewalls", testcommon.MatchIgnoreContext(t, firewall.NewFindFirewallsParams().WithBody(&models.V1FirewallFindRequest{
						NicsMacAddresses: nil,
						Tags:             []string{},
					})), nil).Return(&firewall.FindFirewallsOK{
						Payload: []*models.V1FirewallResponse{
							firewall2,
							firewall1,
						},
					}, nil)
				},
			},
			want: []*models.V1FirewallResponse{
				firewall1,
				firewall2,
			},
			wantTable: new(`
ID  AGE  HOSTNAME             PROJECT    NETWORKS  IPS      PARTITION
1   14d  firewall-hostname-1  project-1  private   1.1.1.1  1
2   14d  firewall-hostname-2  project-1  private   1.1.1.1  1
`),
			wantWideTable: new(`
ID  AGE  HOSTNAME             PROJECT    NETWORKS  IPS      PARTITION
1   14d  firewall-hostname-1  project-1  private   1.1.1.1  1
2   14d  firewall-hostname-2  project-1  private   1.1.1.1  1
`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
1 firewall-1
2 firewall-2
`),
			wantMarkdown: new(`
| ID | AGE | HOSTNAME            | PROJECT   | NETWORKS | IPS     | PARTITION |
|----|-----|---------------------|-----------|----------|---------|-----------|
| 1  | 14d | firewall-hostname-1 | project-1 | private  | 1.1.1.1 | 1         |
| 2  | 14d | firewall-hostname-2 | project-1 | private  | 1.1.1.1 | 1         |
`),
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}

func Test_FirewallCmd_SingleResult(t *testing.T) {
	tests := []*test[*models.V1FirewallResponse]{
		{
			name: "describe",
			cmd: func(want *models.V1FirewallResponse) []string {
				return []string{"firewall", "describe", *want.ID}
			},
			mocks: &client.MetalMockFns{
				Firewall: func(mock *mock.Mock) {
					mock.On("FindFirewall", testcommon.MatchIgnoreContext(t, firewall.NewFindFirewallParams().WithID(*firewall1.ID)), nil).Return(&firewall.FindFirewallOK{
						Payload: firewall1,
					}, nil)
				},
			},
			want: firewall1,
			wantTable: new(`
ID  AGE  HOSTNAME             PROJECT    NETWORKS  IPS      PARTITION
1   14d  firewall-hostname-1  project-1  private   1.1.1.1  1
`),
			wantWideTable: new(`
ID  AGE  HOSTNAME             PROJECT    NETWORKS  IPS      PARTITION
1   14d  firewall-hostname-1  project-1  private   1.1.1.1  1
`),
			template: new("{{ .id }} {{ .name }}"),
			wantTemplate: new(`
1 firewall-1
`),
			wantMarkdown: new(`
| ID | AGE | HOSTNAME            | PROJECT   | NETWORKS | IPS     | PARTITION |
|----|-----|---------------------|-----------|----------|---------|-----------|
| 1  | 14d | firewall-hostname-1 | project-1 | private  | 1.1.1.1 | 1         |
`),
		},
		{
			name: "create",
			cmd: func(want *models.V1FirewallResponse) []string {
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

				args := []string{"firewall", "create",
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
					"--firewall-rules-file", "",
					"--dnsservers", strings.Join(dnsServers, ","),
					"--ntpservers", strings.Join(ntpservers, ","),
				}
				assertExhaustiveArgs(t, args, commonExcludedFileArgs()...)
				return args
			},
			mocks: &client.MetalMockFns{
				Firewall: func(mock *mock.Mock) {
					mock.On("AllocateFirewall", testcommon.MatchIgnoreContext(t, firewall.NewAllocateFirewallParams().WithBody(firewallResponseToCreate(firewall1))), nil).Return(&firewall.AllocateFirewallOK{
						Payload: firewall1,
					}, nil)
				},
			},
			want: firewall1,
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
