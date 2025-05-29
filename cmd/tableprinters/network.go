package tableprinters

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

type network struct {
	parent   *models.V1NetworkResponse
	children []*models.V1NetworkResponse
}

type networks []*network

func (t *TablePrinter) NetworkTable(data []*models.V1NetworkResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Name", "Project", "Partition", "Nat", "Shared", "", "Prefixes", "IP Usage"}
	if wide {
		header = []string{"ID", "Description", "Name", "Project", "Partition", "Nat", "Shared", "Prefixes", "PrivateSuper", "Annotations"}
	}

	nn := &networks{}
	for _, n := range data {
		if n.Parentnetworkid == "" {
			*nn = append(*nn, &network{parent: n})
		}
	}
	for _, n := range data {
		if n.Parentnetworkid != "" {
			if !nn.appendChild(n.Parentnetworkid, n) {
				*nn = append(*nn, &network{parent: n})
			}
		}
	}
	for _, n := range *nn {
		rows = append(rows, addNetwork("", n.parent, wide))
		for i, c := range n.children {
			prefix := "├"
			if i == len(n.children)-1 {
				prefix = "└"
			}
			prefix += "─╴"
			rows = append(rows, addNetwork(prefix, c, wide))
		}
	}

	return header, rows, nil
}

func addNetwork(prefix string, n *models.V1NetworkResponse, wide bool) []string {
	id := fmt.Sprintf("%s%s", prefix, pointer.SafeDeref(n.ID))

	prefixes := strings.Join(n.Prefixes, ",")
	flag := false
	if n.Privatesuper != nil {
		flag = *n.Privatesuper
	}
	privateSuper := fmt.Sprintf("%t", flag)
	nat := fmt.Sprintf("%t", *n.Nat)

	shortIPUsage := nbr
	shortPrefixUsage := nbr
	ipv4Use := 0.0
	ipv4PrefixUse := 0.0
	ipv6Use := 0.0
	ipv6PrefixUse := 0.0

	if n.Consumption != nil {
		consumption := *n.Consumption
		if consumption.IPV4 != nil {
			ipv4Consumption := *consumption.IPV4
			ipv4Use = float64(*ipv4Consumption.UsedIps) / float64(*ipv4Consumption.AvailableIps)

			if *ipv4Consumption.AvailablePrefixes > 0 {
				ipv4PrefixUse = float64(*ipv4Consumption.UsedPrefixes) / float64(*ipv4Consumption.AvailablePrefixes)
			}
		}
		if consumption.IPV6 != nil {
			ipv6Consumption := *consumption.IPV6
			ipv6Use = float64(*ipv6Consumption.UsedIps) / float64(*ipv6Consumption.AvailableIps)

			if *ipv6Consumption.AvailablePrefixes > 0 {
				ipv6PrefixUse = float64(*ipv6Consumption.UsedPrefixes) / float64(*ipv6Consumption.AvailablePrefixes)
			}
		}

		if ipv4Use >= 0.9 || ipv6Use >= 0.9 {
			shortIPUsage = color.RedString(threequarterpie)
		} else if ipv4Use >= 0.7 || ipv6Use >= 0.7 {
			shortIPUsage = color.YellowString(halfpie)
		} else {
			shortIPUsage = color.GreenString(dot)
		}

		if ipv4PrefixUse >= 0.9 || ipv6PrefixUse >= 0.9 {
			shortPrefixUsage = color.RedString(threequarterpie)
		} else if ipv4PrefixUse >= 0.7 || ipv6PrefixUse >= 0.7 {
			shortPrefixUsage = color.YellowString(halfpie)
		} else {
			shortPrefixUsage = color.GreenString(dot)
		}
	}

	max := getMaxLineCount(n.Description, n.Name, n.Projectid, n.Partitionid, nat, prefixes, shortIPUsage, privateSuper)
	for range max - 1 {
		id += "\n│"
	}

	var as []string
	for k, v := range n.Labels {
		as = append(as, k+"="+v)
	}
	shared := "false"
	if n.Shared {
		shared = "true"
	}
	annotations := strings.Join(as, "\n")

	if wide {
		return []string{id, n.Description, n.Name, n.Projectid, n.Partitionid, nat, shared, prefixes, privateSuper, annotations}
	} else {
		return []string{id, n.Name, n.Projectid, n.Partitionid, nat, shared, shortPrefixUsage, prefixes, shortIPUsage}
	}
}

func (nn *networks) appendChild(parentID string, child *models.V1NetworkResponse) bool {
	for _, n := range *nn {
		if *n.parent.ID == parentID {
			n.children = append(n.children, child)
			return true
		}
	}
	return false
}
