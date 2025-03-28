package tableprinters

import (
	"strconv"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
)

func (t *TablePrinter) TenantTable(data []*models.V1TenantResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Name", "Description"}
	if wide {
		header = []string{"ID", "Name", "Description", "Labels", "Annotations", "Quotas"}
	}

	for _, pr := range data {
		var (
			clusterQuota = "∞"
			machineQuota = "∞"
			ipQuota      = "∞"
		)

		if pr.Quotas != nil {
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
		}

		quotas := []string{
			clusterQuota + " Cluster(s)",
			machineQuota + " Machine(s)",
			ipQuota + " IP(s)",
		}

		labels := strings.Join(pr.Meta.Labels, "\n")

		as := genericcli.MapToLabels(pr.Meta.Annotations)
		annotations := strings.Join(as, "\n")

		if wide {
			rows = append(rows, []string{pr.Meta.ID, pr.Name, pr.Description, labels, annotations, strings.Join(quotas, "\n")})
		} else {
			rows = append(rows, []string{pr.Meta.ID, pr.Name, pr.Description})
		}
	}

	return header, rows, nil
}
