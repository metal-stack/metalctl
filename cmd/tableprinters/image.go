package tableprinters

import (
	"fmt"
	"strings"
	"time"

	"github.com/metal-stack/api/go/enum"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
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
			expiration = humanizeDuration(time.Until(time.Time(*image.ExpirationDate)))
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

func (t *TablePrinter) V2ImageTable(data []*apiv2.Image, wide bool) ([]string, [][]string, error) {
	var (
		rows   [][]string
		header = []string{"ID", "Name", "Description", "Features", "Expiration", "Status"}
	)

	for _, image := range data {
		var (
			features []string
		)

		for _, f := range image.Features {
			feature, err := enum.GetStringValue(f)
			if err != nil {
				return nil, nil, err
			}
			features = append(features, feature)
		}

		rows = append(rows, []string{image.Id, pointer.SafeDeref(image.Name), pointer.SafeDeref(image.Description), strings.Join(features, ","), image.ExpiresAt.String(), image.Classification.String()})
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
	})

	return header, rows, nil
}
