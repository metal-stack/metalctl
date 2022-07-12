package tableprinters

import (
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/tag"
)

func (t *TablePrinter) IPTable(data []*models.V1IPResponse, wide bool) ([]string, [][]string, error) {
	var (
		rows [][]string
	)

	header := []string{"IP", "Description", "Name", "Network", "Project", "Type", "Tags"}
	if wide {
		header = []string{"IP", "Allocation UUID", "Description", "Name", "Network", "Project", "Type", "Tags"}
	}

	for _, i := range data {
		ipaddress := pointer.Deref(i.Ipaddress)
		ipType := pointer.Deref(i.Type)
		network := pointer.Deref(i.Networkid)
		project := pointer.Deref(i.Projectid)

		var shortTags []string
		for _, t := range i.Tags {
			parts := strings.Split(t, "=")
			if strings.HasPrefix(t, tag.MachineID+"=") {
				shortTags = append(shortTags, "machine:"+parts[1])
			} else if strings.HasPrefix(t, tag.ClusterServiceFQN+"=") {
				shortTags = append(shortTags, "service:"+parts[1])
			} else {
				shortTags = append(shortTags, t)
			}
		}

		name := genericcli.TruncateMiddle(i.Name, 30)
		description := genericcli.TruncateMiddle(i.Description, 30)
		allocationUUID := ""
		if i.Allocationuuid != nil {
			allocationUUID = *i.Allocationuuid
		}

		if wide {
			rows = append(rows, []string{ipaddress, allocationUUID, i.Description, i.Name, network, project, ipType, strings.Join(i.Tags, "\n")})
		} else {
			rows = append(rows, []string{ipaddress, description, name, network, project, ipType, strings.Join(shortTags, "\n")})
		}
	}

	return header, rows, nil
}
