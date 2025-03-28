package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	metalgo "github.com/metal-stack/metal-go"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metal-lib/pkg/genericcli/printers"
	"github.com/metal-stack/metalctl/cmd/completion"
	"github.com/metal-stack/metalctl/pkg/api"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

type config struct {
	fs              afero.Fs
	out             io.Writer
	driverURL       string
	comp            *completion.Completion
	client          metalgo.Client
	log             *slog.Logger
	describePrinter printers.Printer
	listPrinter     printers.Printer
}

const (
	binaryName = "metalctl"
)

var (
	defaultSSHKeys = [...]string{"id_ed25519", "id_ecdsa", "id_rsa", "id_dsa"}
	// emptyBody is kind of hack because post with "nil" will result into 406 error from the api
	emptyBody = []string{}
)

func Execute() {
	// the config will be provided with more values on cobra init
	// cobra flags do not work so early in the game
	c := &config{
		fs:   afero.NewOsFs(),
		out:  os.Stdout,
		comp: &completion.Completion{},
	}

	err := newRootCmd(c).Execute()
	if err != nil {
		if viper.GetBool("debug") {
			panic(err)
		}
		os.Exit(1)
	}
}

func newRootCmd(c *config) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          binaryName,
		Aliases:      []string{"m"},
		Short:        "a cli to manage entities in the metal-stack api",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			viper.SetFs(c.fs)
			genericcli.Must(viper.BindPFlags(cmd.Flags()))
			genericcli.Must(viper.BindPFlags(cmd.PersistentFlags()))
			// we cannot instantiate the config earlier because
			// cobra flags do not work so early in the game
			genericcli.Must(readConfigFile())
			genericcli.Must(initConfigWithViperCtx(c))
		},
	}

	markdownCmd := &cobra.Command{
		Use:   "markdown",
		Short: "create markdown documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doc.GenMarkdownTree(rootCmd, "./docs")
		},
		DisableAutoGenTag: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			recursiveAutoGenDisable(rootCmd)
		},
	}

	rootCmd.PersistentFlags().StringP("config", "c", "", `alternative config file path, (default is ~/.metalctl/config.yaml).
Example config.yaml:

---
apitoken: "alongtoken"
...

`)
	rootCmd.PersistentFlags().StringP("api-url", "", "", "api server address. Can be specified with METALCTL_API_URL environment variable.")
	rootCmd.PersistentFlags().String("api-token", "", "api token to authenticate. Can be specified with METALCTL_API_TOKEN environment variable.")
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

	genericcli.Must(rootCmd.RegisterFlagCompletionFunc("output-format", completion.OutputFormatListCompletion))

	rootCmd.AddCommand(newAuditCmd(c))
	rootCmd.AddCommand(newFirmwareCmd(c))
	rootCmd.AddCommand(newMachineCmd(c))
	rootCmd.AddCommand(newFirewallCmd(c))
	rootCmd.AddCommand(newProjectCmd(c))
	rootCmd.AddCommand(newTenantCmd(c))
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
	rootCmd.AddCommand(newWhoamiCmd(c))
	rootCmd.AddCommand(newContextCmd(c))
	rootCmd.AddCommand(newVPNCmd(c))
	rootCmd.AddCommand(newUpdateCmd(c))

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

func initConfigWithViperCtx(c *config) error {
	ctx := api.MustDefaultContext()

	c.listPrinter = newPrinterFromCLI(c.out)
	c.describePrinter = defaultToYAMLPrinter(c.out)

	if c.log == nil {
		opts := &slog.HandlerOptions{}
		if viper.GetBool("debug") {
			opts.Level = slog.LevelDebug
		}
		jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
		c.log = slog.New(jsonHandler)
	}

	if c.client != nil {
		return nil
	}

	driverURL := viper.GetString("api-url")
	if driverURL == "" && ctx.ApiURL != "" {
		driverURL = ctx.ApiURL
	}

	var (
		clientOptions []metalgo.ClientOption
		err           error
	)

	hmacKey := viper.GetString("hmac")
	if hmacKey == "" && ctx.HMAC != nil {
		hmacKey = *ctx.HMAC
	}
	hmacAuthType := viper.GetString("hmac-auth-type")
	if hmacAuthType == "" && ctx.HMACAuthType != "" {
		hmacAuthType = ctx.HMACAuthType
	}
	clientOptions = append(clientOptions, metalgo.HMACAuth(hmacKey, hmacAuthType))

	apiToken := viper.GetString("api-token")
	// if there is no api token explicitly specified we try to pull it out of the kubeconfig context
	if apiToken == "" {
		authContext, err := getAuthContext(viper.GetString("kubeconfig"))
		// if there is an error, no kubeconfig exists for us ... this is not really an error
		// since metalctl can be used in scripting with an hmac-key
		if err == nil {
			apiToken = authContext.IDToken
		}
	}
	clientOptions = append(clientOptions, metalgo.BearerToken(apiToken))

	certificateAuthorityData := viper.GetString("certificate-authority-data")
	if certificateAuthorityData == "" && ctx.CertificateAuthorityData != "" {
		certificateAuthorityData = ctx.CertificateAuthorityData
	}
	if certificateAuthorityData != "" {
		tlsClientConfig, err := createTLSClientConfig([]byte(certificateAuthorityData))
		if err != nil {
			return err
		}

		clientOptions = append(clientOptions, metalgo.TLSClientConfig(tlsClientConfig))
	}

	client, err := metalgo.NewClient(driverURL, clientOptions...)
	if err != nil {
		return err
	}

	c.comp.SetClient(client)
	c.driverURL = driverURL
	c.client = client

	return nil
}

func createTLSClientConfig(caData []byte) (*tls.Config, error) {
	caCert := make([]byte, base64.StdEncoding.DecodedLen(len(caData)))
	_, err := base64.StdEncoding.Decode(caCert, caData)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsClientConfig := tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS12,
	}

	return &tlsClientConfig, nil
}

func recursiveAutoGenDisable(cmd *cobra.Command) {
	cmd.DisableAutoGenTag = true
	for _, child := range cmd.Commands() {
		recursiveAutoGenDisable(child)
	}
}
