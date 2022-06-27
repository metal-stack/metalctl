package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metalctl/cmd/completion"
	"github.com/metal-stack/metalctl/pkg/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
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

// Execute is the entrypoint of the metal-go application
func Execute() {
	cmd := newRootCmd()
	err := cmd.Execute()
	if err != nil {
		if viper.GetBool("debug") {
			panic(err)
		}
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
func newRootCmd() *cobra.Command {
	name := "metalctl"
	rootCmd := &cobra.Command{
		Use:          name,
		Aliases:      []string{"m"},
		Short:        "a cli to manage metal devices.",
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
	rootCmd.PersistentFlags().StringP("order", "", "", "order by (comma separated) column(s), possible values: size|id|status|event|when|partition|project")
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
	must(rootCmd.RegisterFlagCompletionFunc("order", completion.OutputOrderListCompletion))

	c := getConfig(name)

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
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newLogoutCmd(c))
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newContextCmd(c))

	rootCmd.AddCommand(newUpdateCmd(c.name))

	must(viper.BindPFlags(rootCmd.PersistentFlags()))

	return rootCmd
}

type config struct {
	name      string
	driverURL string
	comp      *completion.Completion
	driver    *metalgo.Driver
	log       *zap.SugaredLogger
}

func getConfig(name string) *config {
	viper.SetEnvPrefix(strings.ToUpper(name))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType("yaml")
	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("config file path set explicitly, but unreadable:%v", err)
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", name))
		h, err := os.UserHomeDir()
		if err != nil {
			log.Printf("unable to figure out user home directory, skipping config lookup path: %v", err)
		} else {
			viper.AddConfigPath(fmt.Sprintf(h+"/.%s", name))
		}
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			usedCfg := viper.ConfigFileUsed()
			if usedCfg != "" {
				log.Fatalf("config %s file unreadable:%v", usedCfg, err)
			}
		}
	}

	ctx := api.MustDefaultContext()

	logger, err := newLogger()
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
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

	_, driver, err := metalgo.NewDriver(driverURL, apiToken, hmacKey)
	if err != nil {
		log.Fatal(err)
	}

	return &config{
		name:      name,
		comp:      completion.NewCompletion(driver),
		driver:    driver,
		driverURL: driverURL,
		log:       logger,
	}
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
