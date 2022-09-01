package api

import (
	"fmt"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/stretchr/testify/require"
)

var (
	testTime    = time.Date(2022, time.May, 19, 1, 2, 3, 4, time.UTC)
	goodMachine = &models.V1MachineIPMIResponse{
		ID:         pointer.Pointer("0"),
		Liveliness: pointer.Pointer("Alive"),
		Partition:  &models.V1PartitionResponse{ID: pointer.Pointer("a")},
		Ipmi: &models.V1MachineIPMI{
			Address: pointer.Pointer("1.2.3.4"),
			Mac:     pointer.Pointer("aa:bb:..."),
		},
	}
	noPartitionMachine    = &models.V1MachineIPMIResponse{ID: pointer.Pointer("1"), Partition: nil}
	deadMachine           = &models.V1MachineIPMIResponse{ID: pointer.Pointer("2"), Liveliness: pointer.Pointer("Dead"), Partition: &models.V1PartitionResponse{ID: pointer.Pointer("a")}}
	unknownMachine        = &models.V1MachineIPMIResponse{ID: pointer.Pointer("3"), Liveliness: pointer.Pointer("Unknown"), Partition: &models.V1PartitionResponse{ID: pointer.Pointer("a")}}
	notAvailableMachine1  = &models.V1MachineIPMIResponse{ID: pointer.Pointer("4"), Partition: &models.V1PartitionResponse{ID: pointer.Pointer("a")}}
	notAvailableMachine2  = &models.V1MachineIPMIResponse{ID: pointer.Pointer("5"), Liveliness: pointer.Pointer(""), Partition: &models.V1PartitionResponse{ID: pointer.Pointer("a")}}
	failedReclaimMachine  = &models.V1MachineIPMIResponse{ID: pointer.Pointer("6"), Events: &models.V1MachineRecentProvisioningEvents{FailedMachineReclaim: pointer.Pointer(true)}}
	crashingMachine       = &models.V1MachineIPMIResponse{ID: pointer.Pointer("7"), Events: &models.V1MachineRecentProvisioningEvents{CrashLoop: pointer.Pointer(true)}}
	lastEventErrorMachine = &models.V1MachineIPMIResponse{
		ID: pointer.Pointer("8"),
		Events: &models.V1MachineRecentProvisioningEvents{
			LastErrorEvent: &models.V1MachineProvisioningEvent{
				Time: strfmt.DateTime(testTime.Add(-5 * time.Minute)),
			},
		},
	}
	bmcWithoutMacMachine = &models.V1MachineIPMIResponse{ID: pointer.Pointer("9"), Ipmi: &models.V1MachineIPMI{}}
	bmcWithoutIPMachine  = &models.V1MachineIPMIResponse{ID: pointer.Pointer("10"), Ipmi: &models.V1MachineIPMI{}}
	asnSharedMachine1    = &models.V1MachineIPMIResponse{
		ID: pointer.Pointer("11"),
		Allocation: &models.V1MachineAllocation{
			Role: pointer.Pointer(models.V1MachineAllocationRoleFirewall),
			Networks: []*models.V1MachineNetwork{
				{
					Asn: pointer.Pointer(int64(0)),
				},
				{
					Asn: pointer.Pointer(int64(100)),
				},
				{
					Asn: pointer.Pointer(int64(200)),
				},
			},
		},
	}
	asnSharedMachine2 = &models.V1MachineIPMIResponse{
		ID: pointer.Pointer("12"),
		Allocation: &models.V1MachineAllocation{
			Role: pointer.Pointer(models.V1MachineAllocationRoleFirewall),
			Networks: []*models.V1MachineNetwork{
				{
					Asn: pointer.Pointer(int64(1)),
				},
				{
					Asn: pointer.Pointer(int64(100)),
				},
				{
					Asn: pointer.Pointer(int64(200)),
				},
			},
		},
	}
	nonDistinctBMCMachine1 = &models.V1MachineIPMIResponse{
		ID: pointer.Pointer("13"),
		Ipmi: &models.V1MachineIPMI{
			Address: pointer.Pointer("127.0.0.1"),
		},
	}
	nonDistinctBMCMachine2 = &models.V1MachineIPMIResponse{
		ID: pointer.Pointer("14"),
		Ipmi: &models.V1MachineIPMI{
			Address: pointer.Pointer("127.0.0.1"),
		},
	}
)

func init() {
	_ = monkey.Patch(time.Now, func() time.Time { return testTime })
}

func TestFindIssues(t *testing.T) {
	tests := []struct {
		name string
		c    *IssueConfig
		want MachineIssues
	}{
		{
			name: "no partition",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeNoPartition},
				Machines: []*models.V1MachineIPMIResponse{noPartitionMachine, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: noPartitionMachine,
					Issues: Issues{
						toIssue(&IssueNoPartition{}),
					},
				},
			},
		},
		{
			name: "liveliness dead",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeLivelinessDead},
				Machines: []*models.V1MachineIPMIResponse{deadMachine, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: deadMachine,
					Issues: Issues{
						toIssue(&IssueLivelinessDead{}),
					},
				},
			},
		},
		{
			name: "liveliness unknown",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeLivelinessUnknown},
				Machines: []*models.V1MachineIPMIResponse{unknownMachine, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: unknownMachine,
					Issues: Issues{
						toIssue(&IssueLivelinessUnknown{}),
					},
				},
			},
		},
		{
			name: "liveliness not available",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeLivelinessNotAvailable},
				Machines: []*models.V1MachineIPMIResponse{notAvailableMachine1, notAvailableMachine2, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: notAvailableMachine1,
					Issues: Issues{
						toIssue(&IssueLivelinessNotAvailable{}),
					},
				},
				{
					Machine: notAvailableMachine2,
					Issues: Issues{
						toIssue(&IssueLivelinessNotAvailable{}),
					},
				},
			},
		},
		{
			name: "failed machine reclaim",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeFailedMachineReclaim},
				Machines: []*models.V1MachineIPMIResponse{failedReclaimMachine, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: failedReclaimMachine,
					Issues: Issues{
						toIssue(&IssueFailedMachineReclaim{}),
					},
				},
			},
		},
		{
			name: "crashloop",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeCrashLoop},
				Machines: []*models.V1MachineIPMIResponse{crashingMachine, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: crashingMachine,
					Issues: Issues{
						toIssue(&IssueCrashLoop{}),
					},
				},
			},
		},
		{
			name: "last event error",
			c: &IssueConfig{
				Only:               []IssueType{IssueTypeLastEventError},
				Machines:           []*models.V1MachineIPMIResponse{lastEventErrorMachine, goodMachine},
				LastErrorThreshold: 10 * time.Minute,
			},
			want: MachineIssues{
				{
					Machine: lastEventErrorMachine,
					Issues: Issues{
						toIssue(&IssueLastEventError{lastEventThreshold: 10 * time.Minute, details: "occurred 5m0s ago"}),
					},
				},
			},
		},
		{
			name: "bmc without mac",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeBMCWithoutMAC},
				Machines: []*models.V1MachineIPMIResponse{bmcWithoutMacMachine, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: bmcWithoutMacMachine,
					Issues: Issues{
						toIssue(&IssueBMCWithoutMAC{}),
					},
				},
			},
		},
		{
			name: "bmc without ip",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeBMCWithoutIP},
				Machines: []*models.V1MachineIPMIResponse{bmcWithoutIPMachine, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: bmcWithoutIPMachine,
					Issues: Issues{
						toIssue(&IssueBMCWithoutIP{}),
					},
				},
			},
		},
		{
			name: "asn shared",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeASNUniqueness},
				Machines: []*models.V1MachineIPMIResponse{asnSharedMachine1, asnSharedMachine2, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: asnSharedMachine1,
					Issues: Issues{
						toIssue(&IssueASNUniqueness{
							details: fmt.Sprintf("- ASN (100) not unique, shared with [%[1]s]\n- ASN (200) not unique, shared with [%[1]s]", *asnSharedMachine2.ID),
						}),
					},
				},
				{
					Machine: asnSharedMachine2,
					Issues: Issues{
						toIssue(&IssueASNUniqueness{
							details: fmt.Sprintf("- ASN (100) not unique, shared with [%[1]s]\n- ASN (200) not unique, shared with [%[1]s]", *asnSharedMachine1.ID),
						}),
					},
				},
			},
		},
		{
			name: "non distinct bmc ip",
			c: &IssueConfig{
				Only:     []IssueType{IssueTypeNonDistinctBMCIP},
				Machines: []*models.V1MachineIPMIResponse{nonDistinctBMCMachine1, nonDistinctBMCMachine2, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: nonDistinctBMCMachine1,
					Issues: Issues{
						toIssue(&IssueNonDistinctBMCIP{
							details: fmt.Sprintf("BMC IP (127.0.0.1) not unique, shared with [%[1]s]", *nonDistinctBMCMachine2.ID),
						}),
					},
				},
				{
					Machine: nonDistinctBMCMachine2,
					Issues: Issues{
						toIssue(&IssueNonDistinctBMCIP{
							details: fmt.Sprintf("BMC IP (127.0.0.1) not unique, shared with [%[1]s]", *nonDistinctBMCMachine1.ID),
						}),
					},
				},
			},
		},
		{
			name: "severity major",
			c: &IssueConfig{
				Severity: IssueSeverityMajor,
				Machines: []*models.V1MachineIPMIResponse{deadMachine, failedReclaimMachine, notAvailableMachine1, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: deadMachine,
					Issues: Issues{
						toIssue(&IssueLivelinessDead{}),
					},
				},
				{
					Machine: failedReclaimMachine,
					Issues: Issues{
						toIssue(&IssueNoPartition{}),
						toIssue(&IssueFailedMachineReclaim{}),
					},
				},
			},
		},
		{
			name: "severity critical",
			c: &IssueConfig{
				Severity: IssueSeverityCritical,
				Machines: []*models.V1MachineIPMIResponse{deadMachine, failedReclaimMachine, notAvailableMachine1, goodMachine},
			},
			want: MachineIssues{
				{
					Machine: failedReclaimMachine,
					Issues: Issues{
						toIssue(&IssueFailedMachineReclaim{}),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindIssues(tt.c)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(IssueLastEventError{}, IssueASNUniqueness{}, IssueNonDistinctBMCIP{})); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}
		})
	}
}

func TestAllIssues(t *testing.T) {
	issuesTypes := map[IssueType]bool{}
	for _, i := range AllIssues() {
		issuesTypes[i.Type] = true
	}

	for _, ty := range AllIssueTypes() {
		if _, ok := issuesTypes[ty]; !ok {
			t.Errorf("issue of type %s not contained in all issues", ty)
		}
	}
}
