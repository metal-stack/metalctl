package tableprinters

import (
	"strings"
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func (t *TablePrinter) FirewallTable(data []*models.V1FirewallResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Age", "Hostname", "Project", "Networks", "IPs", "Partition"}
		rows   [][]string
	)

	for _, firewall := range data {
		partition := pointer.SafeDeref(pointer.SafeDeref(firewall.Partition).ID)
		alloc := pointer.SafeDeref(firewall.Allocation)
		project := pointer.SafeDeref(alloc.Project)
		hostname := pointer.SafeDeref(alloc.Hostname)

		var nwIPs []string
		var nws []string
		for _, nw := range alloc.Networks {
			nwIPs = append(nwIPs, nw.Ips...)
			nws = append(nws, *nw.Networkid)
		}
		ips := strings.Join(nwIPs, "\n")
		networks := strings.Join(nws, "\n")

		firewallID := *firewall.ID

		age := ""
		if alloc.Created != nil && !time.Time(*alloc.Created).IsZero() {
			age = humanizeDuration(time.Since(time.Time(*alloc.Created)))
		}

		rows = append(rows, []string{firewallID, age, hostname, project, networks, ips, partition})
	}

	return header, rows, nil
}
