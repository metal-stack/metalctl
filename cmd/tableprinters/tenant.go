package tableprinters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
)

func (t *TablePrinter) TenantTable(data []*models.V1TenantResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Name", "Description", "Labels", "Annotations"}
	if wide {
		header = []string{"ID", "Name", "Description", "Quotas Clusters/Machines/IPs", "Labels", "Annotations"}
	}

	for _, pr := range data {
		quotas := "∞/∞/∞"
		if pr.Quotas != nil {
			clusterQuota := "∞"
			machineQuota := "∞"
			ipQuota := "∞"
			qs := pr.Quotas
			if qs.Cluster != nil {
				if qs.Cluster.Quota != 0 {
					clusterQuota = strconv.FormatInt(int64(qs.Cluster.Quota), 10)
				}
			}
			if qs.Machine != nil {
				if qs.Machine.Quota != 0 {
					machineQuota = strconv.FormatInt(int64(qs.Machine.Quota), 10)
				}
			}
			if qs.IP != nil {
				if qs.IP.Quota != 0 {
					ipQuota = strconv.FormatInt(int64(qs.IP.Quota), 10)
				}
			}
			quotas = fmt.Sprintf("%s/%s/%s", clusterQuota, machineQuota, ipQuota)
		}
		labels := strings.Join(pr.Meta.Labels, "\n")
		as := []string{}
		for k, v := range pr.Meta.Annotations {
			as = append(as, k+"="+v)
		}
		annotations := strings.Join(as, "\n")

		if wide {
			rows = append(rows, []string{pr.Meta.ID, pr.Name, pr.Description, quotas, labels, annotations})
		} else {
			rows = append(rows, []string{pr.Meta.ID, pr.Name, pr.Description, labels, annotations})
		}
	}

	return header, rows, nil
}
