package completion

import (
	"github.com/metal-stack/metal-go/api/client/tenant"
	"github.com/spf13/cobra"
)

func (c *Completion) TenantListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	resp, err := c.client.Tenant().ListTenants(tenant.NewListTenantsParams(), nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	var names []string
	for _, p := range resp.Payload {
		names = append(names, p.Meta.ID+"\t/"+p.Name)
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
