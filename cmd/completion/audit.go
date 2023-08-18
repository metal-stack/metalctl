package completion

import (
	"github.com/spf13/cobra"
)

func (c *Completion) AuditTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"http", "grpc", "event"}, cobra.ShellCompDirectiveNoFileComp
}

func (c *Completion) AuditPhaseCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"request", "response", "single", "error", "opened", "closed"}, cobra.ShellCompDirectiveNoFileComp
}
