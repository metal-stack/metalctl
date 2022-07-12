package completion

import (
	metalgo "github.com/metal-stack/metal-go"
	"github.com/spf13/cobra"
)

type Completion struct {
	client metalgo.Client
}

func NewCompletion(client metalgo.Client) *Completion {
	return &Completion{
		client: client,
	}
}

func OutputFormatListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"table", "wide", "markdown", "json", "yaml", "template"}, cobra.ShellCompDirectiveNoFileComp
}
