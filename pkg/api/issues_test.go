package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/stretchr/testify/require"
	"github.com/undefinedlabs/go-mpatch"
)

var (
	testTime = time.Date(2022, time.May, 19, 1, 2, 3, 4, time.UTC)
)

func init() {
	_, err := mpatch.PatchMethod(time.Now, func() time.Time { return testTime })
	if err != nil {
		panic(err)
	}
}

func TestFindIssues(t *testing.T) {
	goodMachine := func(id string) *models.V1MachineIPMIResponse {
		return &models.V1MachineIPMIResponse{
			ID:         pointer.Pointer(id),
			Liveliness: pointer.Pointer("Alive"),
			Partition:  &models.V1PartitionResponse{ID: pointer.Pointer("a")},
			Ipmi: &models.V1MachineIPMI{
				Address:     pointer.Pointer("1.2.3.4"),
				Mac:         pointer.Pointer("aa:bb:00"),
				LastUpdated: pointer.Pointer(strfmt.DateTime(testTime.Add(-1 * time.Minute))),
			},
		}
	}

	tests := []struct {
		name     string
		only     []IssueType
		machines func() []*models.V1MachineIPMIResponse
		want     func(machines []*models.V1MachineIPMIResponse) MachineIssues
	}{
		{
			name: "good machine has no issues",
			machines: func() []*models.V1MachineIPMIResponse {
				return []*models.V1MachineIPMIResponse{
					goodMachine("0"),
				}
			},
			want: nil,
		},
		{
			name: "no partition",
			only: []IssueType{IssueTypeNoPartition},
			machines: func() []*models.V1MachineIPMIResponse {
				noPartitionMachine := goodMachine("no-partition")
				noPartitionMachine.ID = pointer.Pointer("no-partition")
				noPartitionMachine.Partition = nil

				return []*models.V1MachineIPMIResponse{
					noPartitionMachine,
					goodMachine("0"),
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[0],
						Issues: Issues{
							toIssue(&IssueNoPartition{}),
						},
					},
				}
			},
		},
		{
			name: "liveliness dead",
			only: []IssueType{IssueTypeLivelinessDead},
			machines: func() []*models.V1MachineIPMIResponse {
				deadMachine := goodMachine("dead")
				deadMachine.Liveliness = pointer.Pointer("Dead")

				return []*models.V1MachineIPMIResponse{
					deadMachine,
					goodMachine("0"),
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[0],
						Issues: Issues{
							toIssue(&IssueLivelinessDead{}),
						},
					},
				}
			},
		},
		{
			name: "liveliness unknown",
			only: []IssueType{IssueTypeLivelinessUnknown},
			machines: func() []*models.V1MachineIPMIResponse {
				unknownMachine := goodMachine("unknown")
				unknownMachine.Liveliness = pointer.Pointer("Unknown")

				return []*models.V1MachineIPMIResponse{
					unknownMachine,
					goodMachine("0"),
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[0],
						Issues: Issues{
							toIssue(&IssueLivelinessUnknown{}),
						},
					},
				}
			},
		},
		{
			name: "liveliness not available",
			only: []IssueType{IssueTypeLivelinessNotAvailable},
			machines: func() []*models.V1MachineIPMIResponse {
				notAvailableMachine1 := goodMachine("na1")
				notAvailableMachine1.Liveliness = nil

				notAvailableMachine2 := goodMachine("na2")
				notAvailableMachine2.Liveliness = pointer.Pointer("")

				return []*models.V1MachineIPMIResponse{
					notAvailableMachine1,
					notAvailableMachine2,
					goodMachine("0"),
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[0],
						Issues: Issues{
							toIssue(&IssueLivelinessNotAvailable{}),
						},
					},
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueLivelinessNotAvailable{}),
						},
					},
				}
			},
		},
		{
			name: "failed machine reclaim, flag is set",
			only: []IssueType{IssueTypeFailedMachineReclaim},
			machines: func() []*models.V1MachineIPMIResponse {
				failedReclaimMachine := goodMachine("failed")
				failedReclaimMachine.Events = &models.V1MachineRecentProvisioningEvents{FailedMachineReclaim: pointer.Pointer(true)}

				return []*models.V1MachineIPMIResponse{
					goodMachine("0"),
					failedReclaimMachine,
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueFailedMachineReclaim{}),
						},
					},
				}
			},
		},
		{
			name: "failed machine reclaim, phoned home and not allocated (old behavior)",
			only: []IssueType{IssueTypeFailedMachineReclaim},
			machines: func() []*models.V1MachineIPMIResponse {
				failedReclaimMachine := goodMachine("failed")
				failedReclaimMachine.Allocation = nil
				failedReclaimMachine.Events = &models.V1MachineRecentProvisioningEvents{
					Log: []*models.V1MachineProvisioningEvent{
						{
							Event: pointer.Pointer("Phoned Home"),
						},
					},
				}

				return []*models.V1MachineIPMIResponse{
					goodMachine("0"),
					failedReclaimMachine,
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueFailedMachineReclaim{}),
						},
					},
				}
			},
		},
		{
			name: "crashloop",
			only: []IssueType{IssueTypeCrashLoop},
			machines: func() []*models.V1MachineIPMIResponse {
				crashingMachine := goodMachine("crash")
				crashingMachine.Events = &models.V1MachineRecentProvisioningEvents{
					CrashLoop: pointer.Pointer(true),
				}

				return []*models.V1MachineIPMIResponse{
					goodMachine("0"),
					crashingMachine,
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueCrashLoop{}),
						},
					},
				}
			},
		},
		{
			name: "last event error",
			only: []IssueType{IssueTypeLastEventError},
			machines: func() []*models.V1MachineIPMIResponse {
				lastEventErrorMachine := goodMachine("last")
				lastEventErrorMachine.Events = &models.V1MachineRecentProvisioningEvents{
					LastErrorEvent: &models.V1MachineProvisioningEvent{
						Time: strfmt.DateTime(testTime.Add(-5 * time.Minute)),
					},
				}

				return []*models.V1MachineIPMIResponse{
					goodMachine("0"),
					lastEventErrorMachine,
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueLastEventError{details: "occurred 5m0s ago"}),
						},
					},
				}
			},
		},
		{
			name: "bmc without mac",
			only: []IssueType{IssueTypeBMCWithoutMAC},
			machines: func() []*models.V1MachineIPMIResponse {
				bmcWithoutMacMachine := goodMachine("no-mac")
				bmcWithoutMacMachine.Ipmi.Mac = nil

				return []*models.V1MachineIPMIResponse{
					goodMachine("0"),
					bmcWithoutMacMachine,
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueBMCWithoutMAC{}),
						},
					},
				}
			},
		},
		{
			name: "bmc without ip",
			only: []IssueType{IssueTypeBMCWithoutIP},
			machines: func() []*models.V1MachineIPMIResponse {
				bmcWithoutMacMachine := goodMachine("no-ip")
				bmcWithoutMacMachine.Ipmi.Address = nil

				return []*models.V1MachineIPMIResponse{
					goodMachine("0"),
					bmcWithoutMacMachine,
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueBMCWithoutIP{}),
						},
					},
				}
			},
		},
		{
			name: "bmc info outdated",
			only: []IssueType{IssueTypeBMCInfoOutdated},
			machines: func() []*models.V1MachineIPMIResponse {
				bmcOutdatedMachine := goodMachine("outdated")
				bmcOutdatedMachine.Ipmi.LastUpdated = pointer.Pointer(strfmt.DateTime(testTime.Add(-3 * 60 * time.Minute)))

				return []*models.V1MachineIPMIResponse{
					goodMachine("0"),
					bmcOutdatedMachine,
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueBMCInfoOutdated{
								details: "last updated 3h0m0s ago",
							}),
						},
					},
				}
			},
		},
		{
			name: "asn shared",
			only: []IssueType{IssueTypeASNUniqueness},
			machines: func() []*models.V1MachineIPMIResponse {
				asnSharedMachine1 := goodMachine("shared1")
				asnSharedMachine1.Allocation = &models.V1MachineAllocation{
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
				}

				asnSharedMachine2 := goodMachine("shared2")
				asnSharedMachine2.Allocation = &models.V1MachineAllocation{
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
				}

				return []*models.V1MachineIPMIResponse{
					asnSharedMachine1,
					asnSharedMachine2,
					goodMachine("0"),
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[0],
						Issues: Issues{
							toIssue(&IssueASNUniqueness{
								details: fmt.Sprintf("- ASN (100) not unique, shared with [%[1]s]\n- ASN (200) not unique, shared with [%[1]s]", *machines[1].ID),
							}),
						},
					},
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueASNUniqueness{
								details: fmt.Sprintf("- ASN (100) not unique, shared with [%[1]s]\n- ASN (200) not unique, shared with [%[1]s]", *machines[0].ID),
							}),
						},
					},
				}
			},
		},
		{
			name: "non distinct bmc ip",
			only: []IssueType{IssueTypeNonDistinctBMCIP},
			machines: func() []*models.V1MachineIPMIResponse {
				nonDistinctBMCMachine1 := goodMachine("bmc1")
				nonDistinctBMCMachine1.Ipmi.Address = pointer.Pointer("127.0.0.1")

				nonDistinctBMCMachine2 := goodMachine("bmc2")
				nonDistinctBMCMachine2.Ipmi.Address = pointer.Pointer("127.0.0.1")

				return []*models.V1MachineIPMIResponse{
					nonDistinctBMCMachine1,
					nonDistinctBMCMachine2,
					goodMachine("0"),
				}
			},
			want: func(machines []*models.V1MachineIPMIResponse) MachineIssues {
				return MachineIssues{
					{
						Machine: machines[0],
						Issues: Issues{
							toIssue(&IssueNonDistinctBMCIP{
								details: fmt.Sprintf("BMC IP (127.0.0.1) not unique, shared with [%[1]s]", *machines[1].ID),
							}),
						},
					},
					{
						Machine: machines[1],
						Issues: Issues{
							toIssue(&IssueNonDistinctBMCIP{
								details: fmt.Sprintf("BMC IP (127.0.0.1) not unique, shared with [%[1]s]", *machines[0].ID),
							}),
						},
					},
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ms := tt.machines()

			got, err := FindIssues(&IssueConfig{
				Machines:           ms,
				Only:               tt.only,
				LastErrorThreshold: DefaultLastErrorThreshold(),
			})
			require.NoError(t, err)

			var want MachineIssues
			if tt.want != nil {
				want = tt.want(ms)
			}

			if diff := cmp.Diff(want, got, cmp.AllowUnexported(IssueLastEventError{}, IssueASNUniqueness{}, IssueNonDistinctBMCIP{})); diff != "" {
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
