package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/completion"
	"github.com/metal-stack/metalctl/pkg/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

type config struct {
	driverURL string
	comp      *completion.Completion
	client    metalgo.Client
	log       *zap.SugaredLogger
}

const (
	binaryName = "metalctl"
)

var (
	defaultSSHKeys = [...]string{"id_ed25519", "id_rsa", "id_dsa"}

	// will bind all viper flags to subcommands and
	// prevent overwrite of identical flag names from other commands
	// see https://github.com/spf13/viper/issues/233#issuecomment-386791444
	bindPFlags = func(cmd *cobra.Command, args []string) {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			log.Fatal(err.Error())
		}
	}
)

func Execute() {
	cmd := newRootCmd()
	err := cmd.Execute()
	if err != nil {
		if viper.GetBool("debug") {
			panic(err)
		}
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	// the config will be provided with values on cobra init
	// cobra flags do not work so early in the game
	c := &config{
		comp: &completion.Completion{},
	}

	rootCmd := &cobra.Command{
		Use:          binaryName,
		Aliases:      []string{"m"},
		Short:        "a cli to manage metal-stack api",
		Long:         "",
		SilenceUsage: true,
	}

	markdownCmd := &cobra.Command{
		Use:   "markdown",
		Short: "create markdown documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doc.GenMarkdownTree(rootCmd, "./docs")
		},
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", `alternative config file path, (default is ~/.metalctl/config.yaml).
Example config.yaml:

---
apitoken: "alongtoken"
...

`)
	rootCmd.PersistentFlags().StringP("url", "u", "", "api server address. Can be specified with METALCTL_URL environment variable.")
	rootCmd.PersistentFlags().String("apitoken", "", "api token to authenticate. Can be specified with METALCTL_APITOKEN environment variable.")
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kube-config to use for authentication and authorization. Is updated by login. Uses default path if not specified.")

	rootCmd.PersistentFlags().StringP("output-format", "o", "table", "output format (table|wide|markdown|json|yaml|template), wide is a table with more columns.")
	rootCmd.PersistentFlags().StringP("template", "", "", `output template for template output-format, go template format.
For property names inspect the output of -o json or -o yaml for reference.
Example for machines:

metalctl machine list -o template --template "{{ .id }}:{{ .size.id  }}"

`)
	rootCmd.PersistentFlags().Bool("no-headers", false, "do not print headers of table output format (default print headers)")

	rootCmd.PersistentFlags().Bool(forceFlag, false, "skips security prompts (which can be dangerous to set blindly because actions can lead to data loss or additional costs)")
	rootCmd.PersistentFlags().Bool("debug", false, "debug output")
	rootCmd.PersistentFlags().Bool("force-color", false, "force colored output even without tty")

	must(rootCmd.RegisterFlagCompletionFunc("output-format", completion.OutputFormatListCompletion))

	must(viper.BindPFlags(rootCmd.Flags()))
	must(viper.BindPFlags(rootCmd.PersistentFlags()))

	rootCmd.AddCommand(newFirmwareCmd(c))
	rootCmd.AddCommand(newMachineCmd(c))
	rootCmd.AddCommand(newFirewallCmd(c))
	rootCmd.AddCommand(newProjectCmd(c))
	rootCmd.AddCommand(newSizeCmd(c))
	rootCmd.AddCommand(newFilesystemLayoutCmd(c))
	rootCmd.AddCommand(newImageCmd(c))
	rootCmd.AddCommand(newPartitionCmd(c))
	rootCmd.AddCommand(newSwitchCmd(c))
	rootCmd.AddCommand(newNetworkCmd(c))
	rootCmd.AddCommand(markdownCmd)
	rootCmd.AddCommand(newHealthCmd(c))
	rootCmd.AddCommand(newVersionCmd(c))
	rootCmd.AddCommand(newLoginCmd(c))
	rootCmd.AddCommand(newLogoutCmd(c))
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newContextCmd(c))

	rootCmd.AddCommand(newUpdateCmd())

	cobra.OnInitialize(func() {
		must(readConfigFile())
		// we cannot instantiate the client earlier because
		// cobra flags do not work so early in the game
		must(initClient(c))
	})

	return rootCmd
}

func readConfigFile() error {
	viper.SetEnvPrefix(strings.ToUpper(binaryName))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")
	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("config file path set explicitly, but unreadable: %w", err)
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", binaryName))

		h, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to figure out user home directory, skipping config lookup path: %w", err)
		} else {
			viper.AddConfigPath(fmt.Sprintf(h+"/.%s", binaryName))
		}

		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			usedCfg := viper.ConfigFileUsed()
			if usedCfg != "" {
				return fmt.Errorf("config %s file unreadable: %w", usedCfg, err)
			}
		}
	}

	return nil
}

func initClient(c *config) error {
	ctx := api.MustDefaultContext()

	logger, err := newLogger()
	if err != nil {
		return fmt.Errorf("error creating logger: %w", err)
	}

	driverURL := viper.GetString("url")
	if driverURL == "" && ctx.ApiURL != "" {
		driverURL = ctx.ApiURL
	}
	hmacKey := viper.GetString("hmac")
	if hmacKey == "" && ctx.HMAC != nil {
		hmacKey = *ctx.HMAC
	}
	apiToken := viper.GetString("apitoken")

	// if there is no api token explicitly specified we try to pull it out of the kubeconfig context
	if apiToken == "" {
		authContext, err := getAuthContext(viper.GetString("kubeconfig"))
		// if there is an error, no kubeconfig exists for us ... this is not really an error
		// since metalctl can be used in scripting with an hmac-key
		if err == nil {
			apiToken = authContext.IDToken
		}
	}

	client, _, err := metalgo.NewDriver(driverURL, apiToken, hmacKey)
	if err != nil {
		return err
	}

	c.comp.SetClient(client)
	c.driverURL = driverURL
	c.client = client
	c.log = logger

	return nil
}

func newLogger() (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	if viper.GetBool("debug") {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	l, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return l.Sugar(), nil
}

type defaultCmdsConfig[C any, U any, R any] struct {
	gcli *genericcli.GenericCLI[C, U, R]

	singular, plural string
	description      string
	aliases          []string

	createRequestFromCLI func() (C, error)
	updateRequestFromCLI func(args []string) (U, error)

	availableSortKeys []string

	validArgsFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

type defaultCmds struct {
	rootCmd     *cobra.Command
	listCmd     *cobra.Command
	describeCmd *cobra.Command
	createCmd   *cobra.Command
	updateCmd   *cobra.Command
	deleteCmd   *cobra.Command
	applyCmd    *cobra.Command
	editCmd     *cobra.Command
}

func (d *defaultCmds) buildRootCmd(additionalCmds ...*cobra.Command) *cobra.Command {
	d.rootCmd.AddCommand(
		d.listCmd,
		d.describeCmd,
		d.createCmd,
		d.updateCmd,
		d.deleteCmd,
		d.applyCmd,
		d.editCmd,
	)
	d.rootCmd.AddCommand(additionalCmds...)
	return d.rootCmd
}

func newDefaultCmds[C any, U any, R any](c *defaultCmdsConfig[C, U, R]) *defaultCmds {
	cmds := &defaultCmds{
		rootCmd: &cobra.Command{
			Use:     c.singular,
			Short:   fmt.Sprintf("manage %s entities", c.singular),
			Long:    c.description,
			Aliases: c.aliases,
		},
		listCmd: &cobra.Command{
			Use:     "list",
			Aliases: []string{"ls"},
			Short:   fmt.Sprintf("list all %s", c.plural),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.gcli.ListAndPrint(newPrinterFromCLI())
			},
			PreRun: bindPFlags,
		},
		describeCmd: &cobra.Command{
			Use:     "describe <id>",
			Aliases: []string{"get"},
			Short:   fmt.Sprintf("describes the %s", c.singular),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.gcli.DescribeAndPrint(args, defaultToYAMLPrinter())
			},
			ValidArgsFunction: c.validArgsFunc,
		},
		createCmd: &cobra.Command{
			Use:   "create",
			Short: fmt.Sprintf("creates the %s", c.singular),
			RunE: func(cmd *cobra.Command, args []string) error {
				if c.createRequestFromCLI != nil && !viper.IsSet("file") {
					rq, err := c.createRequestFromCLI()
					if err != nil {
						return err
					}
					return c.gcli.CreateAndPrint(rq, defaultToYAMLPrinter())
				}
				return c.gcli.CreateFromFileAndPrint(viper.GetString("file"), defaultToYAMLPrinter())
			},
			PreRun: bindPFlags,
		},
		updateCmd: &cobra.Command{
			Use:   "update",
			Short: fmt.Sprintf("updates the %s", c.singular),
			RunE: func(cmd *cobra.Command, args []string) error {
				if c.updateRequestFromCLI != nil && !viper.IsSet("file") {
					rq, err := c.updateRequestFromCLI(args)
					if err != nil {
						return err
					}
					return c.gcli.UpdateAndPrint(rq, defaultToYAMLPrinter())
				}
				return c.gcli.UpdateFromFileAndPrint(viper.GetString("file"), defaultToYAMLPrinter())
			},
			PreRun:            bindPFlags,
			ValidArgsFunction: c.validArgsFunc,
		},
		deleteCmd: &cobra.Command{
			Use:     "delete <id>",
			Short:   fmt.Sprintf("deletes the %s", c.singular),
			Aliases: []string{"destroy", "rm", "remove"},
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.gcli.DeleteAndPrint(args, defaultToYAMLPrinter())
			},
			PreRun:            bindPFlags,
			ValidArgsFunction: c.validArgsFunc,
		},
		applyCmd: &cobra.Command{
			Use:   "apply",
			Short: fmt.Sprintf("applies one or more %s from a given file", c.plural),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.gcli.ApplyFromFileAndPrint(viper.GetString("file"), newPrinterFromCLI())
			},
			PreRun: bindPFlags,
		},
		editCmd: &cobra.Command{
			Use:   "edit <id>",
			Short: fmt.Sprintf("updates the %s through an editor", c.singular),
			RunE: func(cmd *cobra.Command, args []string) error {
				return c.gcli.EditAndPrint(args, defaultToYAMLPrinter())
			},
			PreRun:            bindPFlags,
			ValidArgsFunction: c.validArgsFunc,
		},
	}

	helpText := func(command string) string {
		return fmt.Sprintf(`filename of the create or update request in yaml format, or - for stdin.

Example:
# %[2]s %[1]s describe %[1]s-1 -o yaml > %[1]s.yaml
# vi %[1]s.yaml
## either via stdin
# cat %[1]s.yaml | %[2]s %[1]s %[3]s -f -
## or via file
# %[2]s %[1]s %[3]s -f %[1]s.yaml
	`, c.singular, binaryName, command)
	}

	cmds.applyCmd.Flags().StringP("file", "f", "", helpText("apply"))
	must(cmds.applyCmd.MarkFlagRequired("file"))

	if c.createRequestFromCLI != nil {
		cmds.createCmd.Flags().StringP("file", "f", "", helpText("create"))
	}

	if c.updateRequestFromCLI != nil {
		cmds.updateCmd.Flags().StringP("file", "f", "", helpText("update"))
	}

	if len(c.availableSortKeys) > 0 {
		cmds.listCmd.Flags().StringSlice("order", []string{}, fmt.Sprintf("order by (comma separated) column(s), sort direction can be changed by appending :asc or :desc behind the column identifier. possible values: %s", strings.Join(c.availableSortKeys, "|")))
		must(cmds.listCmd.RegisterFlagCompletionFunc("order", cobra.FixedCompletions(c.availableSortKeys, cobra.ShellCompDirectiveNoFileComp)))
	}

	return cmds
}
