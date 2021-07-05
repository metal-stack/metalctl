package cmd

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

const (
	cfgFileType = "yaml"
	// name of the application, used for help, config location and env config variable names.
	programName = "metalctl"
)

var (
	ctx            Context
	printer        Printer
	detailer       Detailer
	driverURL      string
	driver         *metalgo.Driver
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

	rootCmd = &cobra.Command{
		Use:     programName,
		Aliases: []string{"m"},
		Short:   "a cli to manage metal devices.",
		Long:    "",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initPrinter()
		},
		SilenceUsage: true,
	}

	markdownCmd = &cobra.Command{
		Use:   "markdown",
		Short: "create markdown documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := doc.GenMarkdownTree(rootCmd, "./docs")
			if err != nil {
				return err
			}
			return nil
		},
	}
)

// Execute is the entrypoint of the metal-go application
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		if viper.GetBool("debug") {
			st := errors.WithStack(err)
			fmt.Printf("%+v", st)
		}
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("config", "c", "", `alternative config file path, (default is ~/.metalctl/config.yaml).
Example config.yaml:

---
apitoken: "alongtoken"
...

`)
	rootCmd.PersistentFlags().StringP("url", "u", "", "api server address. Can be specified with METALCTL_URL environment variable.")
	rootCmd.PersistentFlags().String("apitoken", "", "api token to authenticate. Can be specified with METALCTL_APITOKEN environment variable.")
	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kube-config to use for authentication and authorization. Is updated by login.")
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

	err := rootCmd.RegisterFlagCompletionFunc("output-format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputFormatListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	err = rootCmd.RegisterFlagCompletionFunc("order", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return outputOrderListCompletion()
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	rootCmd.AddCommand(firmwareCmd)
	rootCmd.AddCommand(machineCmd)
	rootCmd.AddCommand(firewallCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(sizeCmd)
	rootCmd.AddCommand(filesystemLayoutCmd)
	rootCmd.AddCommand(imageCmd)
	rootCmd.AddCommand(partitionCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(networkCmd)
	rootCmd.AddCommand(markdownCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(versionCmd)

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(contextCmd)

	rootCmd.AddCommand(updateCmd)

	err = viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("error setup root cmd:%v", err)
	}
}

func initConfig() {
	viper.SetEnvPrefix(strings.ToUpper(programName))
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.SetConfigType(cfgFileType)
	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("config file path set explicitly, but unreadable:%v", err)
		}
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(fmt.Sprintf("/etc/%s", programName))
		h, err := homedir.Dir()
		if err != nil {
			log.Printf("unable to figure out user home directory, skipping config lookup path: %v", err)
		} else {
			viper.AddConfigPath(fmt.Sprintf(h+"/.%s", programName))
		}
		viper.AddConfigPath(".")
		if err := viper.ReadInConfig(); err != nil {
			usedCfg := viper.ConfigFileUsed()
			if usedCfg != "" {
				log.Fatalf("config %s file unreadable:%v", usedCfg, err)
			}
		}
	}

	ctx = mustDefaultContext()
	driverURL = viper.GetString("url")
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
		authContext, err := getAuthContext(viper.GetString("kubeConfig"))
		// if there is an error, no kubeconfig exists for us ... this is not really an error
		// since metalctl can be used in scripting with an hmac-key
		if err == nil {
			apiToken = authContext.IDToken
		}
	}

	var err error
	driver, err = metalgo.NewDriver(driverURL, apiToken, hmacKey)
	if err != nil {
		log.Fatal(err)
	}
}

func initPrinter() {
	var err error
	printer, err = NewPrinter(
		viper.GetString("output-format"),
		viper.GetString("order"),
		viper.GetString("template"),
		viper.GetBool("no-headers"),
	)
	if err != nil {
		log.Fatalf("unable to initialize printer:%v", err)
	}
	detailer, err = NewDetailer(viper.GetString("output-format"))
	if err != nil {
		log.Fatalf("unable to initialize detailer:%v", err)
	}
}

func searchSSHKey() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("unable to determine current user for expanding userdata path:%w", err)
	}
	homeDir := currentUser.HomeDir
	defaultDir := filepath.Join(homeDir, "/.ssh/")
	var key string
	for _, k := range defaultSSHKeys {
		possibleKey := filepath.Join(defaultDir, k)
		_, err := os.ReadFile(possibleKey)
		if err == nil {
			fmt.Printf("using SSH identity: %s. Another identity can be specified with --sshidentity/-p\n",
				possibleKey)
			key = possibleKey
			break
		}
	}

	if key == "" {
		return "", fmt.Errorf("failure to locate a SSH identity in default location (%s). "+
			"Another identity can be specified with --sshidentity/-p\n", defaultDir)
	}
	return key, nil
}

func readFromFile(filePath string) (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("unable to determine current user for expanding userdata path:%w", err)
	}
	homeDir := currentUser.HomeDir

	if filePath == "~" {
		filePath = homeDir
	} else if strings.HasPrefix(filePath, "~/") {
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read from given file %s error:%w", filePath, err)
	}
	return strings.TrimSpace(string(content)), nil
}
