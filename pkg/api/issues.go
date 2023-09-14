package api

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

const (
	// IssueSeverityMinor is an issue that should be checked from time to time but has no bad effects for the user.
	IssueSeverityMinor IssueSeverity = "minor"
	// IssueSeverityMajor is an issue where user experience is affected or provider resources are wasted.
	// overall functionality is still maintained though. major issues should be resolved as soon as possible.
	IssueSeverityMajor IssueSeverity = "major"
	// IssueSeverityCritical is an issue that can lead to disfunction of the system and need to be handled as quickly as possible.
	IssueSeverityCritical IssueSeverity = "critical"

	IssueTypeNoPartition            IssueType = "no-partition"
	IssueTypeLivelinessDead         IssueType = "liveliness-dead"
	IssueTypeLivelinessUnknown      IssueType = "liveliness-unknown"
	IssueTypeLivelinessNotAvailable IssueType = "liveliness-not-available"
	IssueTypeFailedMachineReclaim   IssueType = "failed-machine-reclaim"
	IssueTypeCrashLoop              IssueType = "crashloop"
	IssueTypeLastEventError         IssueType = "last-event-error"
	IssueTypeBMCWithoutMAC          IssueType = "bmc-without-mac"
	IssueTypeBMCWithoutIP           IssueType = "bmc-without-ip"
	IssueTypeBMCInfoOutdated        IssueType = "bmc-info-outdated"
	IssueTypeASNUniqueness          IssueType = "asn-not-unique"
	IssueTypeNonDistinctBMCIP       IssueType = "bmc-no-distinct-ip"
)

type (
	IssueSeverity string
	IssueType     string

	// Issue formulates an issue of a machine
	Issue struct {
		Type        IssueType
		Severity    IssueSeverity
		Description string
		RefURL      string
		Details     string
	}
	// Issues is a list of issues
	Issues []Issue

	// IssueConfig contains configuration parameters for finding machine issues
	IssueConfig struct {
		Machines           []*models.V1MachineIPMIResponse
		Severity           IssueSeverity
		Only               []IssueType
		Omit               []IssueType
		LastErrorThreshold time.Duration
	}

	IssueNoPartition            struct{}
	IssueLivelinessDead         struct{}
	IssueLivelinessUnknown      struct{}
	IssueLivelinessNotAvailable struct{}
	IssueFailedMachineReclaim   struct{}
	IssueCrashLoop              struct{}
	IssueLastEventError         struct {
		details string
	}
	IssueBMCWithoutMAC   struct{}
	IssueBMCWithoutIP    struct{}
	IssueBMCInfoOutdated struct {
		details string
	}
	IssueASNUniqueness struct {
		details string
	}
	IssueNonDistinctBMCIP struct {
		details string
	}

	// MachineWithIssues summarizes a machine with issues
	MachineWithIssues struct {
		Machine *models.V1MachineIPMIResponse
		Issues  Issues
	}
	// MachineIssues is map of a machine response to a list of machine issues
	MachineIssues []*MachineWithIssues

	issue interface {
		// Evaluate decides whether a given machine has the machine issue.
		// the second argument contains additional information that may be required for the issue evaluation
		Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool
		// Spec returns the issue spec of this issue.
		Spec() *issueSpec
		// Details returns additional information on the issue after the evaluation.
		Details() string
	}

	issueSpec struct {
		Type        IssueType
		Severity    IssueSeverity
		Description string
		RefURL      string
	}

	machineIssueMap map[*models.V1MachineIPMIResponse]Issues
)

func DefaultLastErrorThreshold() time.Duration {
	return 7 * 24 * time.Hour
}

func AllIssueTypes() []IssueType {
	return []IssueType{
		IssueTypeNoPartition,
		IssueTypeLivelinessDead,
		IssueTypeLivelinessUnknown,
		IssueTypeLivelinessNotAvailable,
		IssueTypeFailedMachineReclaim,
		IssueTypeCrashLoop,
		IssueTypeLastEventError,
		IssueTypeBMCWithoutMAC,
		IssueTypeBMCWithoutIP,
		IssueTypeBMCInfoOutdated,
		IssueTypeASNUniqueness,
		IssueTypeNonDistinctBMCIP,
	}
}

func AllSevereties() []IssueSeverity {
	return []IssueSeverity{
		IssueSeverityMinor,
		IssueSeverityMajor,
		IssueSeverityCritical,
	}
}

func SeverityFromString(input string) (IssueSeverity, error) {
	switch IssueSeverity(input) {
	case IssueSeverityCritical:
		return IssueSeverityCritical, nil
	case IssueSeverityMajor:
		return IssueSeverityMajor, nil
	case IssueSeverityMinor:
		return IssueSeverityMinor, nil
	default:
		return "", fmt.Errorf("unknown issue severity: %s", input)
	}
}

func (s IssueSeverity) LowerThan(o IssueSeverity) bool {
	smap := map[IssueSeverity]int{
		IssueSeverityCritical: 10,
		IssueSeverityMajor:    5,
		IssueSeverityMinor:    0,
	}

	return smap[s] < smap[o]
}

func AllIssues() Issues {
	var res Issues

	for _, t := range AllIssueTypes() {
		i, err := newIssueFromType(t)
		if err != nil {
			continue
		}

		res = append(res, toIssue(i))
	}

	return res
}

func toIssue(i issue) Issue {
	return Issue{
		Type:        i.Spec().Type,
		Severity:    i.Spec().Severity,
		Description: i.Spec().Description,
		RefURL:      i.Spec().RefURL,
		Details:     i.Details(),
	}
}

func (mis MachineIssues) Get(id string) *MachineWithIssues {
	for _, m := range mis {
		m := m

		if m.Machine == nil || m.Machine.ID == nil {
			continue
		}

		if *m.Machine.ID == id {
			return m
		}
	}

	return nil
}

func FindIssues(c *IssueConfig) (MachineIssues, error) {
	res := machineIssueMap{}

	for _, t := range AllIssueTypes() {
		if !c.includeIssue(t) {
			continue
		}

		for _, m := range c.Machines {
			m := m

			i, err := newIssueFromType(t)
			if err != nil {
				return nil, err
			}

			if m.ID == nil {
				continue
			}

			if i.Evaluate(m, c) {
				res.add(m, toIssue(i))
			}
		}
	}

	return res.toList(), nil
}

func (c *IssueConfig) includeIssue(t IssueType) bool {
	issue, err := newIssueFromType(t)
	if err != nil {
		return false
	}

	if issue.Spec().Severity.LowerThan(c.Severity) {
		return false
	}

	for _, o := range c.Omit {
		if t == o {
			return false
		}
	}

	if len(c.Only) > 0 {
		for _, o := range c.Only {
			if t == o {
				return true
			}
		}
		return false
	}

	return true
}

func newIssueFromType(t IssueType) (issue, error) {
	switch t {
	case IssueTypeNoPartition:
		return &IssueNoPartition{}, nil
	case IssueTypeLivelinessDead:
		return &IssueLivelinessDead{}, nil
	case IssueTypeLivelinessUnknown:
		return &IssueLivelinessUnknown{}, nil
	case IssueTypeLivelinessNotAvailable:
		return &IssueLivelinessNotAvailable{}, nil
	case IssueTypeFailedMachineReclaim:
		return &IssueFailedMachineReclaim{}, nil
	case IssueTypeCrashLoop:
		return &IssueCrashLoop{}, nil
	case IssueTypeLastEventError:
		return &IssueLastEventError{}, nil
	case IssueTypeBMCWithoutMAC:
		return &IssueBMCWithoutMAC{}, nil
	case IssueTypeBMCWithoutIP:
		return &IssueBMCWithoutIP{}, nil
	case IssueTypeBMCInfoOutdated:
		return &IssueBMCInfoOutdated{}, nil
	case IssueTypeASNUniqueness:
		return &IssueASNUniqueness{}, nil
	case IssueTypeNonDistinctBMCIP:
		return &IssueNonDistinctBMCIP{}, nil
	default:
		return nil, fmt.Errorf("unknown issue type: %s", t)
	}
}

func (mim machineIssueMap) add(m *models.V1MachineIPMIResponse, issue Issue) {
	issues, ok := mim[m]
	if !ok {
		issues = Issues{}
	}
	issues = append(issues, issue)
	mim[m] = issues
}

func (mim machineIssueMap) toList() MachineIssues {
	var res MachineIssues

	for m, issues := range mim {
		res = append(res, &MachineWithIssues{
			Machine: m,
			Issues:  issues,
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return pointer.SafeDeref(res[i].Machine.ID) < pointer.SafeDeref(res[j].Machine.ID)
	})

	return res
}

func (i *IssueNoPartition) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeNoPartition,
		Severity:    IssueSeverityMajor,
		Description: "machine with no partition",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#no-partition",
	}
}

func (i *IssueNoPartition) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	return m.Partition == nil
}

func (i *IssueNoPartition) Details() string {
	return ""
}

func (i *IssueLivelinessDead) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeLivelinessDead,
		Severity:    IssueSeverityMajor,
		Description: "the machine is not sending events anymore",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#liveliness-dead",
	}
}

func (i *IssueLivelinessDead) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	return m.Liveliness != nil && *m.Liveliness == "Dead"
}

func (i *IssueLivelinessDead) Details() string {
	return ""
}

func (i *IssueLivelinessUnknown) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeLivelinessUnknown,
		Severity:    IssueSeverityMajor,
		Description: "the machine is not sending LLDP alive messages anymore",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#liveliness-unknown",
	}
}

func (i *IssueLivelinessUnknown) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	return m.Liveliness != nil && *m.Liveliness == "Unknown"
}

func (i *IssueLivelinessUnknown) Details() string {
	return ""
}

func (i *IssueLivelinessNotAvailable) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeLivelinessNotAvailable,
		Severity:    IssueSeverityMinor,
		Description: "the machine liveliness is not available",
	}
}
func (i *IssueLivelinessNotAvailable) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	if m.Liveliness == nil {
		return true
	}

	allowed := map[string]bool{
		"Alive":   true,
		"Dead":    true,
		"Unknown": true,
	}

	return !allowed[*m.Liveliness]
}

func (i *IssueLivelinessNotAvailable) Details() string {
	return ""
}

func (i *IssueFailedMachineReclaim) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeFailedMachineReclaim,
		Severity:    IssueSeverityCritical,
		Description: "machine phones home but not allocated",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#failed-machine-reclaim",
	}
}

func (i *IssueFailedMachineReclaim) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	if pointer.SafeDeref(pointer.SafeDeref(m.Events).FailedMachineReclaim) {
		return true
	}

	// compatibility: before the provisioning FSM was renewed, this state could be detected the following way
	// we should keep this condition
	if m.Allocation == nil && pointer.SafeDeref(pointer.SafeDeref(pointer.FirstOrZero(pointer.SafeDeref(m.Events).Log)).Event) == "Phoned Home" {
		return true
	}

	return false
}

func (i *IssueFailedMachineReclaim) Details() string {
	return ""
}

func (i *IssueCrashLoop) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeCrashLoop,
		Severity:    IssueSeverityMajor,
		Description: fmt.Sprintf("machine is in a provisioning crash loop (%s)", Loop),
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#crashloop",
	}
}

func (i *IssueCrashLoop) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	if pointer.SafeDeref(pointer.SafeDeref(m.Events).CrashLoop) {
		if m.Events != nil && len(m.Events.Log) > 0 && *m.Events.Log[0].Event == "Waiting" {
			// Machine which are waiting are not considered to have issues
		} else {
			return true
		}
	}
	return false
}

func (i *IssueCrashLoop) Details() string {
	return ""
}

func (i *IssueLastEventError) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeLastEventError,
		Severity:    IssueSeverityMinor,
		Description: "the machine had an error during the provisioning lifecycle",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#last-event-error",
	}
}

func (i *IssueLastEventError) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	if c.LastErrorThreshold == 0 {
		return false
	}

	if pointer.SafeDeref(m.Events).LastErrorEvent != nil {
		timeSince := time.Since(time.Time(m.Events.LastErrorEvent.Time))
		if timeSince < c.LastErrorThreshold {
			i.details = fmt.Sprintf("occurred %s ago", timeSince.String())
			return true
		}
	}

	return false
}

func (i *IssueLastEventError) Details() string {
	return i.details
}

func (i *IssueBMCWithoutMAC) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeBMCWithoutMAC,
		Severity:    IssueSeverityMajor,
		Description: "BMC has no mac address",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#bmc-without-mac",
	}
}

func (i *IssueBMCWithoutMAC) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	return m.Ipmi != nil && (m.Ipmi.Mac == nil || *m.Ipmi.Mac == "")
}

func (i *IssueBMCWithoutMAC) Details() string {
	return ""
}

func (i *IssueBMCWithoutIP) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeBMCWithoutIP,
		Severity:    IssueSeverityMajor,
		Description: "BMC has no ip address",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#bmc-without-ip",
	}
}

func (i *IssueBMCWithoutIP) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	return m.Ipmi != nil && (m.Ipmi.Address == nil || *m.Ipmi.Address == "")
}

func (i *IssueBMCWithoutIP) Details() string {
	return ""
}

func (i *IssueBMCInfoOutdated) Details() string {
	return i.details
}

func (i *IssueBMCInfoOutdated) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	if m.Ipmi == nil {
		i.details = "machine ipmi has never been set"
		return true
	}

	if m.Ipmi.LastUpdated == nil || m.Ipmi.LastUpdated.IsZero() {
		// "last_updated has not been set yet"
		return false
	}

	lastUpdated := time.Since(time.Time(*m.Ipmi.LastUpdated))

	if lastUpdated > 20*time.Minute {
		i.details = fmt.Sprintf("last updated %s ago", lastUpdated.String())
		return true
	}

	return false
}

func (*IssueBMCInfoOutdated) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeBMCInfoOutdated,
		Severity:    IssueSeverityMajor,
		Description: "BMC has not been updated from either metal-hammer or metal-bmc",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#bmc-info-outdated",
	}
}

func (i *IssueASNUniqueness) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeASNUniqueness,
		Severity:    IssueSeverityMinor,
		Description: "The ASN is not unique (only impact on firewalls)",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#asn-not-unique",
	}
}

func (i *IssueASNUniqueness) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	var (
		machineASNs  = map[int64][]*models.V1MachineIPMIResponse{}
		overlaps     []string
		isNoFirewall = func(m *models.V1MachineIPMIResponse) bool {
			return m.Allocation == nil || m.Allocation.Role == nil || *m.Allocation.Role != models.V1MachineAllocationRoleFirewall
		}
	)

	if isNoFirewall(m) {
		return false
	}

	for _, n := range m.Allocation.Networks {
		if n.Asn == nil {
			continue
		}

		machineASNs[*n.Asn] = nil
	}

	for _, machineFromAll := range c.Machines {
		if pointer.SafeDeref(machineFromAll.ID) == pointer.SafeDeref(m.ID) {
			continue
		}
		otherMachine := machineFromAll

		if isNoFirewall(otherMachine) {
			continue
		}

		for _, n := range otherMachine.Allocation.Networks {
			if n.Asn == nil {
				continue
			}

			_, ok := machineASNs[*n.Asn]
			if !ok {
				continue
			}

			machineASNs[*n.Asn] = append(machineASNs[*n.Asn], otherMachine)
		}
	}

	var asnList []int64
	for asn := range machineASNs {
		asnList = append(asnList, asn)
	}
	sort.Slice(asnList, func(i, j int) bool {
		return asnList[i] < asnList[j]
	})

	for _, asn := range asnList {
		overlappingMachines, ok := machineASNs[asn]
		if !ok || len(overlappingMachines) == 0 {
			continue
		}

		var sharedIDs []string
		for _, m := range overlappingMachines {
			m := m
			sharedIDs = append(sharedIDs, *m.ID)
		}

		overlaps = append(overlaps, fmt.Sprintf("- ASN (%d) not unique, shared with %s", asn, sharedIDs))
	}

	if len(overlaps) == 0 {
		return false
	}

	sort.Slice(overlaps, func(i, j int) bool {
		return overlaps[i] < overlaps[j]
	})

	i.details = strings.Join(overlaps, "\n")

	return true
}

func (i *IssueASNUniqueness) Details() string {
	return i.details
}

func (i *IssueNonDistinctBMCIP) Spec() *issueSpec {
	return &issueSpec{
		Type:        IssueTypeNonDistinctBMCIP,
		Description: "BMC IP address is not distinct",
		RefURL:      "https://docs.metal-stack.io/stable/installation/troubleshoot/#bmc-no-distinct-ip",
	}
}

func (i *IssueNonDistinctBMCIP) Evaluate(m *models.V1MachineIPMIResponse, c *IssueConfig) bool {
	if m.Ipmi == nil || m.Ipmi.Address == nil {
		return false
	}

	var (
		bmcIP    = *m.Ipmi.Address
		overlaps []string
	)

	for _, machineFromAll := range c.Machines {
		if pointer.SafeDeref(machineFromAll.ID) == pointer.SafeDeref(m.ID) {
			continue
		}
		otherMachine := machineFromAll

		if otherMachine.Ipmi == nil || otherMachine.Ipmi.Address == nil {
			continue
		}

		if bmcIP == *otherMachine.Ipmi.Address {
			overlaps = append(overlaps, *otherMachine.ID)
		}
	}

	if len(overlaps) == 0 {
		return false
	}

	i.details = fmt.Sprintf("BMC IP (%s) not unique, shared with %s", bmcIP, overlaps)

	return true
}

func (i *IssueNonDistinctBMCIP) Details() string {
	return i.details
}
