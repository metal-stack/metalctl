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

	usage := fmt.Sprintf("IPs:     %v/%v", *n.Usage.UsedIps, *n.Usage.AvailableIps)

	ipUse := float64(*n.Usage.UsedIps) / float64(*n.Usage.AvailableIps)
	shortIPUsage := nbr
	if ipUse >= 0.9 {
		shortIPUsage += color.RedString(dot)
	} else if ipUse >= 0.7 {
		shortIPUsage += color.YellowString(dot)
	} else {
		shortIPUsage += color.GreenString(dot)
	}

	shortPrefixUsage := ""
	if *n.Usage.AvailablePrefixes > 0 {
		prefixUse := float64(*n.Usage.UsedPrefixes) / float64(*n.Usage.AvailablePrefixes)
		if prefixUse >= 0.9 {
			shortPrefixUsage = exclamationMark
		}
		usage = fmt.Sprintf("%s\nPrefixes:%d/%d", usage, *n.Usage.UsedPrefixes, *n.Usage.AvailablePrefixes)
	}

	max := getMaxLineCount(n.Description, n.Name, n.Projectid, n.Partitionid, nat, prefixes, usage, privateSuper)
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
		return []string{id, n.Description, n.Name, n.Projectid, n.Partitionid, nat, shared, prefixes, usage, privateSuper, annotations}
	} else {
		return []string{id, n.Name, n.Projectid, n.Partitionid, nat, shared, prefixes, shortPrefixUsage, shortIPUsage}
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
