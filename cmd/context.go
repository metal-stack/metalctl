package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/metal-stack/metalctl/cmd/output"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
)

func newContextCmd(c *config) *cobra.Command {
	contextCmd := &cobra.Command{
		Use:               "context <name>",
		Aliases:           []string{"ctx"},
		Short:             "manage metalctl context",
		Long:              "context defines the backend to which metalctl talks to. You can switch back and forth with \"-\"",
		ValidArgsFunction: c.comp.ContextListCompletion,
		Example: `
~/.metalctl/config.yaml
---
current: prod
contexts:
  prod:
    url: https://api.metal-stack.io/metal
    issuer_url: https://dex.metal-stack.io/dex
    client_id: metal_client
    client_secret: 456
  dev:
    url: https://api.metal-stack.dev/metal
    issuer_url: https://dex.metal-stack.dev/dex
    client_id: metal_client
    client_secret: 123
...
`,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 1 {
				return contextSet(args)
			}
			if len(args) == 0 {
				return c.contextList()
			}
			return nil
		},
		PreRun: bindPFlags,
	}

	contextShortCmd := &cobra.Command{
		Use:   "short",
		Short: "only show the default context name",
		RunE: func(cmd *cobra.Command, args []string) error {
			return contextShort()
		},
		PreRun: bindPFlags,
	}
	contextCmd.AddCommand(contextShortCmd)
	return contextCmd
}

func contextShort() error {
	ctxs, err := api.GetContexts()
	if err != nil {
		return err
	}
	fmt.Println(ctxs.CurrentContext)
	return nil
}

func contextSet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no context name given")
	}
	if args[0] == "-" {
		return previous()
	}
	ctxs, err := api.GetContexts()
	if err != nil {
		return err
	}
	nextCtx := args[0]
	_, ok := ctxs.Contexts[nextCtx]
	if !ok {
		return fmt.Errorf("context %s not found", nextCtx)
	}
	if nextCtx == ctxs.CurrentContext {
		fmt.Printf("%s context \"%s\" already active\n", color.GreenString("✔"), color.GreenString(ctxs.CurrentContext))
		return nil
	}
	ctxs.PreviousContext = ctxs.CurrentContext
	ctxs.CurrentContext = nextCtx
	return api.WriteContexts(ctxs)
}

func previous() error {
	ctxs, err := api.GetContexts()
	if err != nil {
		return err
	}
	prev := ctxs.PreviousContext
	if prev == "" {
		return fmt.Errorf("no previous context found")
	}
	curr := ctxs.CurrentContext
	ctxs.PreviousContext = curr
	ctxs.CurrentContext = prev
	return api.WriteContexts(ctxs)
}

func (c *config) contextList() error {
	ctxs, err := api.GetContexts()
	if err != nil {
		return err
	}
	return output.New().Print(ctxs)
}
