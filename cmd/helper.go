package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/metalctl/pkg/api"
)

func parseNetworks(values []string) ([]*models.V1MachineAllocationNetwork, error) {
	nets := []*models.V1MachineAllocationNetwork{}
	for _, netWithFlag := range values {
		id, autoAcquire, err := splitNetwork(netWithFlag)
		if err != nil {
			return nil, err
		}

		net := models.V1MachineAllocationNetwork{
			Autoacquire: &autoAcquire,
			Networkid:   &id,
		}
		nets = append(nets, &net)
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
