package api

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/fatih/color"
	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metalctl/cmd/completion"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	v2client "github.com/metal-stack/api/go/client"
)

type Config struct {
	FS              afero.Fs
	Out             io.Writer
	ApiURL          string
	ApiV2URL        string
	Comp            *completion.Completion
	Client          metalgo.Client
	V2Client        v2client.Client
	Log             *slog.Logger
	DescribePrinter printers.Printer
	ListPrinter     printers.Printer
}

func NewContextCmd(c *Config) *cobra.Command {
	contextCmd := &cobra.Command{
		Use:               "context <name>",
		Aliases:           []string{"ctx"},
		Short:             "manage metalctl context",
		Long:              "context defines the backend to which metalctl talks to. You can switch back and forth with \"-\"",
		ValidArgsFunction: ContextListCompletion,
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
				return ContextSet(args)
			}
			if len(args) == 0 {
				return c.contextList()
			}
			return nil
		},
	}

	contextShortCmd := &cobra.Command{
		Use:   "short",
		Short: "only show the default context name",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ContextShort()
		},
	}
	contextCmd.AddCommand(contextShortCmd)
	return contextCmd
}

func ContextShort() error {
	ctxs, err := GetContexts()
	if err != nil {
		return err
	}
	fmt.Println(ctxs.CurrentContext)
	return nil
}

func ContextSet(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no context name given")
	}
	if args[0] == "-" {
		return previous()
	}
	ctxs, err := GetContexts()
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
	return WriteContexts(ctxs)
}

func previous() error {
	ctxs, err := GetContexts()
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
	return WriteContexts(ctxs)
}

func (c *Config) contextList() error {
	ctxs, err := GetContexts()
	if err != nil {
		return err
	}
	return c.ListPrinter.Print(ctxs)
}

func (c *Config) NewRequestContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
