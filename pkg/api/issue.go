package api

import (
	"fmt"

	"github.com/metal-stack/metal-go/api/models"
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
	IssueCrashLoop = Issue{
		ShortName:   "crashloop",
		Description: fmt.Sprintf("machine is in a provisioning crash loop (%s)", circle),
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#crashloop",
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
		IssueCrashLoop,
		IssueASNUniqueness,
		IssueBMCWithoutMAC,
		IssueBMCWithoutIP,
		IssueNonDistinctBMCIP,
	}
)
