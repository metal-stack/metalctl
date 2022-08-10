package defaultscmds

import (
	"fmt"
	"log"
	"strings"

	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/printers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type DefaultCmd string

const (
	ListCmd     DefaultCmd = "list"
	DescribeCmd DefaultCmd = "describe"
	CreateCmd   DefaultCmd = "create"
	UpdateCmd   DefaultCmd = "update"
	DeleteCmd   DefaultCmd = "delete"
	ApplyCmd    DefaultCmd = "apply"
	EditCmd     DefaultCmd = "edit"
)

func AllCmds() map[DefaultCmd]bool {
	return map[DefaultCmd]bool{
		ListCmd:     true,
		DescribeCmd: true,
		CreateCmd:   true,
		UpdateCmd:   true,
		DeleteCmd:   true,
		ApplyCmd:    true,
		EditCmd:     true,
	}
}

func IncludeCmds(cmds ...DefaultCmd) map[DefaultCmd]bool {
	res := map[DefaultCmd]bool{}

	for _, c := range cmds {
		res[c] = true
	}

	return res
}

var bindPFlags = func(cmd *cobra.Command, args []string) {
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		log.Fatal(err.Error())
	}
}

type Config[C any, U any, R any] struct {
	GenericCLI *genericcli.GenericCLI[C, U, R]

	IncludeCmds map[DefaultCmd]bool

	BinaryName       string
	Singular, Plural string
	Description      string
	Aliases          []string

	CreateRequestFromCLI func() (C, error)
	UpdateRequestFromCLI func(args []string) (U, error)

	AvailableSortKeys []string

	ValidArgsFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

	ListCmdMutateFn     func(cmd *cobra.Command)
	DescribeCmdMutateFn func(cmd *cobra.Command)
	CreateCmdMutateFn   func(cmd *cobra.Command)
	UpdateCmdMutateFn   func(cmd *cobra.Command)
	DeleteCmdMutateFn   func(cmd *cobra.Command)
	ApplyCmdMutateFn    func(cmd *cobra.Command)
	EditCmdMutateFn     func(cmd *cobra.Command)
}

type cmds[C any, U any, R any] struct {
	rootCmd     *cobra.Command
	defaultCmds []*cobra.Command
}

func New[C any, U any, R any](c *Config[C, U, R]) *cmds[C, U, R] {
	if len(c.IncludeCmds) == 0 {
		c.IncludeCmds = AllCmds()
	}

	cmds := &cmds[C, U, R]{
		rootCmd: &cobra.Command{
			Use:     c.Singular,
			Short:   fmt.Sprintf("manage %s entities", c.Singular),
			Long:    c.Description,
			Aliases: c.Aliases,
		},
	}

	var defaultCmds []*cobra.Command

	if _, ok := c.IncludeCmds[ListCmd]; ok {
		cmd := &cobra.Command{
			Use:     "list",
			Aliases: []string{"ls"},
			Short:   fmt.Sprintf("list all %s", c.Plural),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.GenericCLI.ListAndPrint(printers.NewPrinterFromCLI())
			},
			PreRun: bindPFlags,
		}

		if len(c.AvailableSortKeys) > 0 {
			cmd.Flags().StringSlice("order", []string{}, fmt.Sprintf("order by (comma separated) column(s), sort direction can be changed by appending :asc or :desc behind the column identifier. possible values: %s", strings.Join(c.AvailableSortKeys, "|")))
			must(cmd.RegisterFlagCompletionFunc("order", cobra.FixedCompletions(c.AvailableSortKeys, cobra.ShellCompDirectiveNoFileComp)))
		}

		if c.ListCmdMutateFn != nil {
			c.ListCmdMutateFn(cmd)
		}

		defaultCmds = append(defaultCmds, cmd)
	}

	if _, ok := c.IncludeCmds[DescribeCmd]; ok {
		cmd := &cobra.Command{
			Use:     "describe <id>",
			Aliases: []string{"get"},
			Short:   fmt.Sprintf("describes the %s", c.Singular),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.GenericCLI.DescribeAndPrint(args, printers.DefaultToYAMLPrinter())
			},
			ValidArgsFunction: c.ValidArgsFunc,
		}

		if c.DescribeCmdMutateFn != nil {
			c.DescribeCmdMutateFn(cmd)
		}

		defaultCmds = append(defaultCmds, cmd)
	}

	if _, ok := c.IncludeCmds[CreateCmd]; ok {
		cmd := &cobra.Command{
			Use:   "create",
			Short: fmt.Sprintf("creates the %s", c.Singular),
			RunE: func(cmd *cobra.Command, args []string) error {
				if c.CreateRequestFromCLI != nil && !viper.IsSet("file") {
					rq, err := c.CreateRequestFromCLI()
					if err != nil {
						return err
					}
					return c.GenericCLI.CreateAndPrint(rq, printers.DefaultToYAMLPrinter())
				}
				return c.GenericCLI.CreateFromFileAndPrint(viper.GetString("file"), printers.DefaultToYAMLPrinter())
			},
			PreRun: bindPFlags,
		}

		if c.CreateRequestFromCLI != nil {
			cmd.Flags().StringP("file", "f", "", c.helpText("create"))
		}

		if c.CreateCmdMutateFn != nil {
			c.CreateCmdMutateFn(cmd)
		}

		defaultCmds = append(defaultCmds, cmd)
	}

	if _, ok := c.IncludeCmds[UpdateCmd]; ok {
		cmd := &cobra.Command{
			Use:   "update",
			Short: fmt.Sprintf("updates the %s", c.Singular),
			RunE: func(cmd *cobra.Command, args []string) error {
				if c.UpdateRequestFromCLI != nil && !viper.IsSet("file") {
					rq, err := c.UpdateRequestFromCLI(args)
					if err != nil {
						return err
					}
					return c.GenericCLI.UpdateAndPrint(rq, printers.DefaultToYAMLPrinter())
				}
				return c.GenericCLI.UpdateFromFileAndPrint(viper.GetString("file"), printers.DefaultToYAMLPrinter())
			},
			PreRun:            bindPFlags,
			ValidArgsFunction: c.ValidArgsFunc,
		}

		if c.UpdateRequestFromCLI != nil {
			cmd.Flags().StringP("file", "f", "", c.helpText("update"))
		}

		if c.UpdateCmdMutateFn != nil {
			c.UpdateCmdMutateFn(cmd)
		}

		defaultCmds = append(defaultCmds, cmd)
	}

	if _, ok := c.IncludeCmds[DeleteCmd]; ok {
		cmd := &cobra.Command{
			Use:     "delete <id>",
			Short:   fmt.Sprintf("deletes the %s", c.Singular),
			Aliases: []string{"destroy", "rm", "remove"},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.GenericCLI.DeleteAndPrint(args, printers.DefaultToYAMLPrinter())
			},
			PreRun:            bindPFlags,
			ValidArgsFunction: c.ValidArgsFunc,
		}

		if c.DeleteCmdMutateFn != nil {
			c.DeleteCmdMutateFn(cmd)
		}

		defaultCmds = append(defaultCmds, cmd)
	}

	if _, ok := c.IncludeCmds[ApplyCmd]; ok {
		cmd := &cobra.Command{
			Use:   "apply",
			Short: fmt.Sprintf("applies one or more %s from a given file", c.Plural),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.GenericCLI.ApplyFromFileAndPrint(viper.GetString("file"), printers.NewPrinterFromCLI())
			},
			PreRun: bindPFlags,
		}

		cmd.Flags().StringP("file", "f", "", c.helpText("apply"))
		must(cmd.MarkFlagRequired("file"))

		if c.ApplyCmdMutateFn != nil {
			c.ApplyCmdMutateFn(cmd)
		}

		defaultCmds = append(defaultCmds, cmd)
	}

	if _, ok := c.IncludeCmds[EditCmd]; ok {
		cmd := &cobra.Command{
			Use:   "edit <id>",
			Short: fmt.Sprintf("updates the %s through an editor", c.Singular),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.GenericCLI.EditAndPrint(args, printers.DefaultToYAMLPrinter())
			},
			PreRun:            bindPFlags,
			ValidArgsFunction: c.ValidArgsFunc,
		}

		if c.EditCmdMutateFn != nil {
			c.EditCmdMutateFn(cmd)
		}

		defaultCmds = append(defaultCmds, cmd)
	}

	return cmds
}

func (c *cmds[C, U, R]) Build(additionalCmds ...*cobra.Command) *cobra.Command {
	c.rootCmd.AddCommand(c.defaultCmds...)
	c.rootCmd.AddCommand(additionalCmds...)
	return c.rootCmd
}

func (c *Config[C, U, R]) helpText(command string) string {
	return fmt.Sprintf(`filename of the create or update request in yaml format, or - for stdin.

Example:
# %[2]s %[1]s describe %[1]s-1 -o yaml > %[1]s.yaml
# vi %[1]s.yaml
## either via stdin
# cat %[1]s.yaml | %[2]s %[1]s %[3]s -f -
## or via file
# %[2]s %[1]s %[3]s -f %[1]s.yaml
	`, c.Singular, c.BinaryName, command)
}

func must(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}
