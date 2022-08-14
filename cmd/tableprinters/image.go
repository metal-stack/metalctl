package tableprinters

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
)

func (t *TablePrinter) ImageTable(data []*models.V1ImageResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"ID", "Name", "Description", "Features", "Expiration", "Status"}
	if wide {
		header = []string{"ID", "Name", "Description", "Features", "Expiration", "Status", "UsedBy"}
	}

	sort.SliceStable(data, func(i, j int) bool { return *data[i].ID < *data[j].ID })
	for _, image := range data {
		id := pointer.SafeDeref(image.ID)
		features := strings.Join(image.Features, ",")
		name := image.Name
		description := image.Description
		status := image.Classification

		expiration := ""
		if image.ExpirationDate != nil {
			expiration = humanizeDuration(time.Until(time.Time(*image.ExpirationDate)))
		}

		usedBy := fmt.Sprintf("%d", len(image.Usedby))
		if wide {
			usedBy = strings.Join(image.Usedby, "\n")
		}

		if wide {
			rows = append(rows, []string{id, name, description, features, expiration, status, usedBy})
		} else {
			rows = append(rows, []string{id, name, description, features, expiration, status})
		}
	}

	return header, rows, nil
}
