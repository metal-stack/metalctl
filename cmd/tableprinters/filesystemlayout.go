package tableprinters

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/olekukonko/tablewriter"
)

func (t *TablePrinter) FSLTable(data []*models.V1FilesystemLayoutResponse, wide bool) ([]string, [][]string, error) {
	var (
		header = []string{"ID", "Description", "Filesystems", "Sizes", "Images"}
		rows   [][]string
	)

	for _, fsl := range data {
		imageConstraints := []string{}
		for os, v := range fsl.Constraints.Images {
			imageConstraints = append(imageConstraints, os+" "+v)
		}

		fsls := fsl.Filesystems
		sort.Slice(fsls, func(i, j int) bool { return depth(fsls[i].Path) < depth(fsls[j].Path) })

		fss := bytes.NewBufferString("")
		w := tabwriter.NewWriter(fss, 0, 0, 0, ' ', 0)
		for i, fs := range fsls {
			fmt.Fprintf(w, "%s\t  \t%s", fs.Path, *fs.Device)
			if i != len(fsls)-1 {
				fmt.Fprintln(w)
			}
		}
		err := w.Flush()
		if err != nil {
			return nil, nil, err
		}

		rows = append(rows, []string{pointer.SafeDeref(fsl.ID), fsl.Description, fss.String(), strings.Join(fsl.Constraints.Sizes, "\n"), strings.Join(imageConstraints, "\n")})
	}

	t.t.MutateTable(func(table *tablewriter.Table) {
		table.SetAutoWrapText(false)
	})

	return header, rows, nil
}
