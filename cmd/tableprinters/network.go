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

	header := []string{"ID", "Name", "Project", "Partition", "Nat", "Shared", "Prefixes", "", "IPs"}
	if wide {
		header = []string{"ID", "Description", "Name", "Project", "Partition", "Nat", "Shared", "Prefixes", "Usage", "PrivateSuper", "Annotations"}
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
	shortIPv4IPUsage := nbr
	shortIPv4PrefixUsage := ""
	ipv4usage := ""

	// TODO activate
	// shortIPv6IPUsage := nbr
	// shortIPv6PrefixUsage := ""
	// ipv6usage := ""
	if n.Consumption != nil {
		consumption := *n.Consumption
		if consumption.IPV4 != nil {
			ipv4Consumption := *consumption.IPV4
			ipv4usage = fmt.Sprintf("IPs:     %v/%v", *ipv4Consumption.UsedIps, *ipv4Consumption.AvailableIps)

			ipv4Use := float64(*ipv4Consumption.UsedIps) / float64(*ipv4Consumption.AvailableIps)
			if ipv4Use >= 0.9 {
				shortIPv4IPUsage += color.RedString(dot)
			} else if ipv4Use >= 0.7 {
				shortIPv4IPUsage += color.YellowString(dot)
			} else {
				shortIPv4IPUsage += color.GreenString(dot)
			}

			if *ipv4Consumption.AvailablePrefixes > 0 {
				prefixUse := float64(*ipv4Consumption.UsedPrefixes) / float64(*ipv4Consumption.AvailablePrefixes)
				if prefixUse >= 0.9 {
					shortIPv4PrefixUsage = color.RedString(dot)
				}
				ipv4usage = fmt.Sprintf("%s\nPrefixes:%d/%d", ipv4usage, *ipv4Consumption.UsedPrefixes, *ipv4Consumption.AvailablePrefixes)
			}
		}
		// TODO activate
		// if consumption.IPV6 != nil {
		// 	ipv6Consumption := *consumption.IPV6
		// 	ipv6usage = fmt.Sprintf("IPs:     %v/%v", *ipv6Consumption.UsedIps, *ipv6Consumption.AvailableIps)

		// 	ipv6Use := float64(*ipv6Consumption.UsedIps) / float64(*ipv6Consumption.AvailableIps)
		// 	if ipv6Use >= 0.9 {
		// 		shortIPv6IPUsage += color.RedString(dot)
		// 	} else if ipv6Use >= 0.7 {
		// 		shortIPv6IPUsage += color.YellowString(dot)
		// 	} else {
		// 		shortIPv6IPUsage += color.GreenString(dot)
		// 	}

		// 	if *ipv6Consumption.AvailablePrefixes > 0 {
		// 		prefixUse := float64(*ipv6Consumption.UsedPrefixes) / float64(*ipv6Consumption.AvailablePrefixes)
		// 		if prefixUse >= 0.9 {
		// 			shortIPv6PrefixUsage = color.RedString(dot)
		// 		}
		// 		ipv6usage = fmt.Sprintf("%s\nPrefixes:%d/%d", ipv6usage, *ipv6Consumption.UsedPrefixes, *ipv6Consumption.AvailablePrefixes)
		// 	}
		// }
	}

	max := getMaxLineCount(n.Description, n.Name, n.Projectid, n.Partitionid, nat, prefixes, ipv4usage, privateSuper)
	for i := 0; i < max-1; i++ {
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
		return []string{id, n.Description, n.Name, n.Projectid, n.Partitionid, nat, shared, prefixes, ipv4usage, privateSuper, annotations}
	} else {
		return []string{id, n.Name, n.Projectid, n.Partitionid, nat, shared, prefixes, shortIPv4PrefixUsage, shortIPv4IPUsage}
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
