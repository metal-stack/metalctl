package tableprinters

import (
	"fmt"
	"time"

	"github.com/metal-stack/metal-go/api/models"
)

func (t *TablePrinter) AuditTable(data []*models.V1AuditResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows   [][]string
		header = []string{"Time", "Request ID", "Detail", "Path", "Code", "Tenant", "User", "Phase"}
	)

	if wide {
		header = []string{"Time", "Request ID", "Detail", "Path", "Code", "Tenant", "User", "Phase", "Body"}
	}

	for _, trace := range data {
		var statusCode string
		if trace.StatusCode != 0 {
			statusCode = fmt.Sprintf("%d", trace.StatusCode)
		}
		row := []string{
			time.Time(trace.Timestamp).Format(time.StampMilli),
			trace.Rqid,
			trace.Detail,
			trace.Path,
			statusCode,
			trace.Tenant,
			trace.User,
			trace.Phase,
		}

		if wide {
			row = append(row, trace.Body)
		}

		rows = append(rows, row)
	}

	return header, rows, nil
}
