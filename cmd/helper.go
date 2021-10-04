package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/metalctl/pkg/api"

	metalgo "github.com/metal-stack/metal-go"
	"gopkg.in/yaml.v3"

	"github.com/spf13/viper"
)

func atLeastOneViperStringFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetString(flag) != "" {
			return true
		}
	}
	return false
}

func atLeastOneViperStringSliceFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if len(viper.GetStringSlice(flag)) > 0 {
			return true
		}
	}
	return false
}

func atLeastOneViperBoolFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetBool(flag) {
			return true
		}
	}
	return false
}

func atLeastOneViperInt64FlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetInt64(flag) != 0 {
			return true
		}
	}
	return false
}

func viperString(flag string) *string {
	if viper.GetString(flag) == "" {
		return nil
	}
	value := viper.GetString(flag)
	return &value
}

func viperStringSlice(flag string) []string {
	value := viper.GetStringSlice(flag)
	if len(value) == 0 {
		return nil
	}
	return value
}

func viperBool(flag string) *bool {
	if !viper.GetBool(flag) {
		return nil
	}
	value := viper.GetBool(flag)
	return &value
}

func viperInt64(flag string) *int64 {
	if viper.GetInt64(flag) == 0 {
		return nil
	}
	value := viper.GetInt64(flag)
	return &value
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

func boolPtr(b bool) *bool {
	return &b
}

func labelsFromTags(tags []string) map[string]string {
	labels := make(map[string]string)
	for _, tag := range tags {
		parts := strings.Split(tag, "=")
		partslen := len(parts)
		switch partslen {
		case 1:
			labels[tag] = "true"
		case 2:
			labels[parts[0]] = parts[1]
		default:
			values := strings.Join(parts[1:], "")
			labels[parts[0]] = values
		}
	}
	return labels
}

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

func annotationsAsMap(annotations []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, a := range annotations {
		parts := strings.Split(strings.TrimSpace(a), "=")
		if len(parts) != 2 {
			return result, fmt.Errorf("given annotation %s does not contain exactly one =", a)
		}
		result[parts[0]] = parts[1]
	}
	return result, nil
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
