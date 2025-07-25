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
			Description:      "firewall allocation 1",
			Filesystemlayout: fsl1,
			Hostname:         pointer.Pointer("firewall-hostname-1"),
			Image:            image1,
			Name:             pointer.Pointer("firewall-1"),
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
			Role:       pointer.Pointer(models.V1MachineAllocationRoleFirewall),
			SSHPubKeys: []string{"sshpubkey"},
			Succeeded:  pointer.Pointer(true),
			UserData:   "---userdata---",
			DNSServers: []*models.V1DNSServer{{IP: pointer.Pointer("8.8.8.8")}},
			NtpServers: []*models.V1NTPServer{{Address: pointer.Pointer("1.pool.ntp.org")}},
		},
		Bios: &models.V1MachineBIOS{
			Date:    pointer.Pointer("biosdata"),
			Vendor:  pointer.Pointer("biosvendor"),
			Version: pointer.Pointer("biosversion"),
		},
		Description: "firewall 1",
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
		Name:       "firewall-1",
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
	firewall2 = &models.V1FirewallResponse{
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
			Description:      "firewall allocation 2",
			Filesystemlayout: fsl1,
			Hostname:         pointer.Pointer("firewall-hostname-2"),
			Image:            image1,
			Name:             pointer.Pointer("firewall-2"),
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
			Role:       pointer.Pointer(models.V1MachineAllocationRoleFirewall),
			SSHPubKeys: []string{"sshpubkey"},
			Succeeded:  pointer.Pointer(true),
			UserData:   "---userdata---",
		},
		Bios: &models.V1MachineBIOS{
			Date:    pointer.Pointer("biosdata"),
			Vendor:  pointer.Pointer("biosvendor"),
			Version: pointer.Pointer("biosversion"),
		},
		Description: "firewall 2",
		Events: &models.V1MachineRecentProvisioningEvents{
			CrashLoop:            pointer.Pointer(false),
			FailedMachineReclaim: pointer.Pointer(false),
			LastErrorEvent:       &models.V1MachineProvisioningEvent{},
			LastEventTime:        strfmt.DateTime(testTime.Add(-1 * time.Minute)),
			Log: []*models.V1MachineProvisioningEvent{
				{
					Event:   pointer.Pointer("Phoned Home"),
					Message: "phoning home",
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
		Name:       "firewall-2",
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
			wantTable: pointer.Pointer(`
ID  AGE  HOSTNAME             PROJECT    NETWORKS  IPS      PARTITION
1   14d  firewall-hostname-1  project-1  private   1.1.1.1  1
2   14d  firewall-hostname-2  project-1  private   1.1.1.1  1
`),
			wantWideTable: pointer.Pointer(`
ID  AGE  HOSTNAME             PROJECT    NETWORKS  IPS      PARTITION
1   14d  firewall-hostname-1  project-1  private   1.1.1.1  1
2   14d  firewall-hostname-2  project-1  private   1.1.1.1  1
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 firewall-1
2 firewall-2
`),
			wantMarkdown: pointer.Pointer(`
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
			wantTable: pointer.Pointer(`
ID  AGE  HOSTNAME             PROJECT    NETWORKS  IPS      PARTITION
1   14d  firewall-hostname-1  project-1  private   1.1.1.1  1
`),
			wantWideTable: pointer.Pointer(`
ID  AGE  HOSTNAME             PROJECT    NETWORKS  IPS      PARTITION
1   14d  firewall-hostname-1  project-1  private   1.1.1.1  1
`),
			template: pointer.Pointer("{{ .id }} {{ .name }}"),
			wantTemplate: pointer.Pointer(`
1 firewall-1
`),
			wantMarkdown: pointer.Pointer(`
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
