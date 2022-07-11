package tableprinters

import (
	"github.com/metal-stack/metalctl/pkg/api"
)

func (t *TablePrinter) ContextTable(data *api.Contexts, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"Name", "API URL", "Issuer URL"}
		rows   [][]string
	)

	for name, c := range data.Contexts {
		if name == data.CurrentContext {
			name = name + " [*]"
		}
		rows = append(rows, []string{name, c.ApiURL, c.IssuerURL})
	}

	return header, rows, nil
}
