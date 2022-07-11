package completion

import (
	metalgo "github.com/metal-stack/metal-go"
	"github.com/spf13/cobra"
)

type Completion struct {
	driver *metalgo.Driver
}

func NewCompletion(driver *metalgo.Driver) *Completion {
	return &Completion{
		driver: driver,
	}
}

func OutputFormatListCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{"table", "wide", "markdown", "json", "yaml", "template"}, cobra.ShellCompDirectiveNoFileComp
}
