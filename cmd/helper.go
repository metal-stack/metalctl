package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/cmd/tableprinters"
	"github.com/metal-stack/metalctl/pkg/api"

	metalgo "github.com/metal-stack/metal-go"
	"gopkg.in/yaml.v3"

	"github.com/spf13/viper"
)

func newPrinterFromCLI() genericcli.Printer {
	var printer genericcli.Printer
	var err error

	switch format := viper.GetString("output-format"); format {
	case "yaml":
		printer = genericcli.NewYAMLPrinter()
	case "json":
		printer = genericcli.NewJSONPrinter()
	case "table", "wide", "markdown":
		cfg := &genericcli.TablePrinterConfig{
			ToHeaderAndRows: tableprinters.ToHeaderAndRows,
			Wide:            format == "wide",
			Markdown:        format == "markdown",
			NoHeaders:       viper.GetBool("no-headers"),
		}
		printer, err = genericcli.NewTablePrinter(cfg)
		if err != nil {
			log.Fatalf("unable to initialize printer: %v", err)
		}
	case "template":
		printer, err = genericcli.NewTemplatePrinter(viper.GetString("template"))
		if err != nil {
			log.Fatalf("unable to initialize printer: %v", err)
		}
	default:
		log.Fatalf("unknown output format: %q", format)
	}

	if viper.IsSet("force-color") {
		enabled := viper.GetBool("force-color")
		if enabled {
			color.NoColor = false
		} else {
			color.NoColor = true
		}
	}

	return printer
}

func defaultToYAMLPrinter() genericcli.Printer {
	if viper.IsSet("output-format") {
		return newPrinterFromCLI()
	}
	return genericcli.NewYAMLPrinter()
}

func parseNetworks(values []string) ([]metalgo.MachineAllocationNetwork, error) {
	nets := []metalgo.MachineAllocationNetwork{}
	for _, netWithFlag := range values {
		id, autoAcquire, err := splitNetwork(netWithFlag)
		if err != nil {
			return nil, err
		}

		net := metalgo.MachineAllocationNetwork{
			Autoacquire: autoAcquire,
			NetworkID:   id,
		}
		nets = append(nets, net)
	}
	return nets, nil
}

func splitNetwork(value string) (string, bool, error) {
	splitNets := strings.SplitN(value, ":", 2)
	id := splitNets[0]
	if len(splitNets) > 1 {
		mode := strings.ToLower(splitNets[1])
		switch mode {
		// case NETWORK:auto
		case "auto":
			return id, true, nil
		// case NETWORK:noauto
		case "noauto":
			return id, false, nil
		// case NETWORK:<illegal value>
		default:
			return "", false, fmt.Errorf("illegal mode: %s", mode)
		}
	}
	// case: NETWORK, defaults to NETWORK:auto
	return id, true, nil
}

// func shortID(machineID string) string {
// 	result := strings.ReplaceAll(machineID, "00000000-", "")
// 	result = strings.ReplaceAll(result, "0000-", "")
// 	return result
// }

// func longID(shortID string) string {
// 	machineIDPattern := []byte("00000000-0000-0000-0000-000000000000")
// 	longIDLength := len(machineIDPattern)
// 	result := machineIDPattern
// 	shortIDSlice := []byte(strings.TrimSpace(shortID))
// 	for i := len(shortID) - 1; i >= 0; i-- {
// 		pos := longIDLength - i - 1
// 		shortPos := len(shortIDSlice) - i - 1
// 		result[pos] = shortIDSlice[shortPos]
// 	}
// 	return string(result)
// }

// readFrom will either read from stdin (-) or a file path an marshall from yaml to data
func readFrom(from string, data interface{}, f func(target interface{})) error {
	var reader io.Reader
	var err error
	switch from {
	case "-":
		reader = os.Stdin
	default:
		reader, err = os.Open(from)
		if err != nil {
			return fmt.Errorf("unable to open %s %w", from, err)
		}
	}
	dec := yaml.NewDecoder(reader)
	for {
		err := dec.Decode(data)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("decode error: %w", err)
		}
		f(data)
	}
	return nil
}

const cloudContext = "cloudctl"

// getAuthContext reads AuthContext from given kubeconfig
func getAuthContext(kubeconfig string) (*auth.AuthContext, error) {
	cs, err := api.GetContexts()
	if err != nil {
		return nil, err
	}
	authContext, err := auth.GetAuthContext(kubeconfig, formatContextName(cloudContext, cs.CurrentContext))
	if err != nil {
		return nil, err
	}

	if !authContext.AuthProviderOidc {
		return nil, fmt.Errorf("active user %s has no oidc authProvider, check config", authContext.User)
	}

	return &authContext, nil
}

// formatContextName returns the contextName for the given suffix. suffix can be empty.
func formatContextName(prefix string, suffix string) string {
	contextName := prefix
	if suffix != "" {
		contextName = fmt.Sprintf("%s-%s", cloudContext, suffix)
	}
	return contextName
}

// Prompt the user to given compare text
func Prompt(msg, compare string) error {
	fmt.Print(msg + " ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	text := scanner.Text()
	if text != compare {
		return fmt.Errorf("unexpected answer given (%q), aborting...", text)
	}
	return nil
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
