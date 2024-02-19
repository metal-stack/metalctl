package tableprinters

import (
	"fmt"
	"strings"
	"time"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/spf13/viper"
)

func (t *TablePrinter) ImageTable(data []*models.V1ImageResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows      [][]string
		header    = []string{"ID", "Name", "Description", "Features", "Expiration", "Status", "UsedBy"}
		showUsage = viper.GetBool("show-usage")
	)

	for _, image := range data {
		id := pointer.SafeDeref(image.ID)
		features := strings.Join(image.Features, ",")
		name := image.Name
		description := image.Description
		status := image.Classification

		expiration := ""
		if image.ExpirationDate != nil {
			expiration = HumanizeDuration(time.Until(time.Time(*image.ExpirationDate)))
		}

		usedBy := fmt.Sprintf("%d", len(image.Usedby))
		if wide {
			usedBy = strings.Join(image.Usedby, "\n")
		}
		if !showUsage {
			usedBy = ""
		}

		rows = append(rows, []string{id, name, description, features, expiration, status, usedBy})
	}

	return header, rows, nil
}
