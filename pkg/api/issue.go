package api

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/viper"
)

type (
	// MachineWithIssues summarizes a machine with issues
	MachineWithIssues struct {
		Machine models.V1MachineIPMIResponse
		Issues  Issues
	}
	// MachineIssues is map of a machine response to a list of machine issues
	MachineIssues map[string]MachineWithIssues

	// Issue formulates an issue of a machine
	Issue struct {
		ShortName   string
		Description string
		RefURL      string
	}
	// Issues is a list of machine issues
	Issues []Issue
)

var (
	circle = "↻"

	IssueNoPartition = Issue{
		ShortName:   "no-partition",
		Description: "machine with no partition",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#no-partition",
	}
	IssueLivelinessDead = Issue{
		ShortName:   "liveliness-dead",
		Description: "the machine is not sending events anymore",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#liveliness-dead",
	}
	IssueLivelinessUnknown = Issue{
		ShortName:   "liveliness-unknown",
		Description: "the machine is not sending LLDP alive messages anymore",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#liveliness-unknown",
	}
	IssueLivelinessNotAvailable = Issue{
		ShortName:   "liveliness-not-available",
		Description: "the machine liveliness is not available",
	}
	IssueFailedMachineReclaim = Issue{
		ShortName:   "failed-machine-reclaim",
		Description: "machine phones home but not allocated",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#failed-machine-reclaim",
	}
	IssueIncompleteCycles = Issue{
		ShortName:   "incomplete-cycles",
		Description: fmt.Sprintf("machine has an incomplete lifecycle (%s)", circle),
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#incomplete-cycles",
	}
	IssueASNUniqueness = Issue{
		ShortName:   "asn-not-unique",
		Description: "The ASN is not unique (only impact on firewalls)",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#asn-not-unique",
	}
	IssueBMCWithoutMAC = Issue{
		ShortName:   "bmc-without-mac",
		Description: "BMC has no mac address",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#bmc-without-mac",
	}
	IssueBMCWithoutIP = Issue{
		ShortName:   "bmc-without-ip",
		Description: "BMC has no ip address",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#bmc-without-ip",
	}
	IssueNonDistinctBMCIP = Issue{
		ShortName:   "bmc-no-distinct-ip",
		Description: "BMC IP address is not distinct",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#bmc-no-distinct-ip",
	}

	AllIssues = Issues{
		IssueNoPartition,
		IssueLivelinessDead,
		IssueLivelinessUnknown,
		IssueLivelinessNotAvailable,
		IssueFailedMachineReclaim,
		IssueIncompleteCycles,
		IssueASNUniqueness,
		IssueBMCWithoutMAC,
		IssueBMCWithoutIP,
		IssueNonDistinctBMCIP,
	}
)

func GetMachineIssues(machines []*models.V1MachineIPMIResponse) MachineIssues {
	only := viper.GetStringSlice("only")
	omit := viper.GetStringSlice("omit")

	var (
		res      = MachineIssues{}
		asnMap   = map[int64][]models.V1MachineIPMIResponse{}
		bmcIPMap = map[string][]models.V1MachineIPMIResponse{}

		conditionalAppend = func(issues Issues, issue Issue) Issues {
			for _, o := range omit {
				if issue.ShortName == o {
					return issues
				}
			}

			if len(only) > 0 {
				for _, o := range only {
					if issue.ShortName == o {
						return append(issues, issue)
					}
				}
				return issues
			}

			return append(issues, issue)
		}
	)

	for _, m := range machines {
		var issues Issues

		if m.Partition == nil {
			issues = conditionalAppend(issues, IssueNoPartition)
		}

		if m.Liveliness != nil {
			switch *m.Liveliness {
			case "Alive":
			case "Dead":
				issues = conditionalAppend(issues, IssueLivelinessDead)
			case "Unknown":
				issues = conditionalAppend(issues, IssueLivelinessUnknown)
			default:
				issues = conditionalAppend(issues, IssueLivelinessNotAvailable)
			}
		} else {
			issues = conditionalAppend(issues, IssueLivelinessNotAvailable)
		}

		if m.Allocation == nil && len(m.Events.Log) > 0 && *m.Events.Log[0].Event == "Phoned Home" {
			issues = conditionalAppend(issues, IssueFailedMachineReclaim)
		}

		if m.Events.IncompleteProvisioningCycles != nil &&
			*m.Events.IncompleteProvisioningCycles != "" &&
			*m.Events.IncompleteProvisioningCycles != "0" {
			if m.Events != nil && len(m.Events.Log) > 0 && *m.Events.Log[0].Event == "Waiting" {
				// Machine which are waiting are not considered to have issues
			} else {
				issues = conditionalAppend(issues, IssueIncompleteCycles)
			}
		}

		if m.Ipmi != nil {
			if m.Ipmi.Mac == nil || *m.Ipmi.Mac == "" {
				issues = conditionalAppend(issues, IssueBMCWithoutMAC)
			}

			if m.Ipmi.Address == nil || *m.Ipmi.Address == "" {
				issues = conditionalAppend(issues, IssueBMCWithoutIP)
			} else {
				entries := bmcIPMap[*m.Ipmi.Address]
				entries = append(entries, *m)
				bmcIPMap[*m.Ipmi.Address] = entries
			}
		}

		if m.Allocation != nil && m.Allocation.Role != nil && *m.Allocation.Role == models.V1MachineAllocationRoleFirewall {
			// collecting ASN overlaps
			for _, n := range m.Allocation.Networks {
				if n.Asn == nil {
					continue
				}

				machines, ok := asnMap[*n.Asn]
				if !ok {
					machines = []models.V1MachineIPMIResponse{}
				}

				alreadyContained := false
				for _, mm := range machines {
					if *mm.ID == *m.ID {
						alreadyContained = true
						break
					}
				}

				if alreadyContained {
					continue
				}

				machines = append(machines, *m)
				asnMap[*n.Asn] = machines
			}
		}

		if len(issues) > 0 {
			res[*m.ID] = MachineWithIssues{
				Machine: *m,
				Issues:  issues,
			}
		}
	}

	includeASN := true
	for _, o := range omit {
		if o == IssueASNUniqueness.ShortName {
			includeASN = false
			break
		}
	}

	if includeASN {
		for asn, ms := range asnMap {
			if len(ms) < 2 {
				continue
			}

			for _, m := range ms {
				var sharedIDs []string
				for _, mm := range ms {
					if *m.ID == *mm.ID {
						continue
					}
					sharedIDs = append(sharedIDs, *mm.ID)
				}

				mWithIssues, ok := res[*m.ID]
				if !ok {
					mWithIssues = MachineWithIssues{
						Machine: m,
					}
				}
				issue := IssueASNUniqueness
				issue.Description = fmt.Sprintf("ASN (%d) not unique, shared with %s", asn, sharedIDs)
				mWithIssues.Issues = append(mWithIssues.Issues, issue)
				res[*m.ID] = mWithIssues
			}
		}
	}

	includeDistinctBMC := true
	for _, o := range omit {
		if o == IssueNonDistinctBMCIP.ShortName {
			includeDistinctBMC = false
			break
		}
	}

	if includeDistinctBMC {
		for ip, ms := range bmcIPMap {
			if len(ms) < 2 {
				continue
			}

			for _, m := range ms {
				var sharedIDs []string
				for _, mm := range ms {
					if *m.ID == *mm.ID {
						continue
					}
					sharedIDs = append(sharedIDs, *mm.ID)
				}

				mWithIssues, ok := res[*m.ID]
				if !ok {
					mWithIssues = MachineWithIssues{
						Machine: m,
					}
				}
				issue := IssueNonDistinctBMCIP
				issue.Description = fmt.Sprintf("BMC IP (%s) not unique, shared with %s", ip, sharedIDs)
				mWithIssues.Issues = append(mWithIssues.Issues, issue)
				res[*m.ID] = mWithIssues
			}
		}
	}

	return res
}
