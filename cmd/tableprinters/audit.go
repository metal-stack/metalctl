package tableprinters

import (
	"fmt"
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
)

func (t *TablePrinter) AuditTable(data []*models.V1AuditResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows   [][]string
		header = []string{"Time", "Request ID", "Component", "Detail", "Path", "Code", "User"}
	)

	if wide {
		header = []string{"Time", "Request ID", "Component", "Detail", "Path", "Code", "User", "Tenant", "Body"}
	}

	for _, trace := range data {
		var statusCode string
		if trace.StatusCode != 0 {
			statusCode = fmt.Sprintf("%d", trace.StatusCode)
		}
		row := []string{
			time.Time(trace.Timestamp).Format(time.DateTime),
			trace.Rqid,
			trace.Component,
			trace.Detail,
			trace.Path,
			statusCode,
			trace.User,
		}

		if wide {
			row = append(row, trace.Tenant)

			body := genericcli.TruncateEnd(trace.Body, 40)
			row = append(row, body)
		}

		rows = append(rows, row)
	}

	return header, rows, nil
}
