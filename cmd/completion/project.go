package completion

import (
	"github.com/metal-stack/metal-go/api/client/project"
	"github.com/spf13/cobra"
)

func (c *Completion) ProjectListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Project().ListProjects(project.NewListProjectsParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Payload {
		names = append(names, p.Meta.ID+"\t"+p.TenantID+"/"+p.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
