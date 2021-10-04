package api

import (
	"fmt"

	"github.com/metal-stack/metal-lib/auth"
)

const CloudContext = "metalctl"

// getAuthContext reads AuthContext from given kubeconfig
func GetAuthContext(kubeconfig string) (*auth.AuthContext, error) {
	cs, err := GetContexts()
	if err != nil {
		return nil, err
	}
	authContext, err := auth.GetAuthContext(kubeconfig, FormatContextName(CloudContext, cs.CurrentContext))
	if err != nil {
		return nil, err
	}

	if !authContext.AuthProviderOidc {
		return nil, fmt.Errorf("active user %s has no oidc authProvider, check config", authContext.User)
	}

	return &authContext, nil
}

// formatContextName returns the contextName for the given suffix. suffix can be empty.
func FormatContextName(prefix string, suffix string) string {
	contextName := prefix
	if suffix != "" {
		contextName = fmt.Sprintf("%s-%s", CloudContext, suffix)
	}
	return contextName
}
