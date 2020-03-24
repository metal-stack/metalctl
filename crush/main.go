package main

import (
	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/metalctl/crush/cmd"
	"github.com/spf13/viper"
)

func main() {
	apiToken := getAPIToken()

	r := cmd.NewRunner("http://metal.test.fi-ts.io", apiToken, "")
	r.Run()
}

func getAPIToken() string {

	apiToken := viper.GetString("apitoken")

	// if there is no api token explicitly specified we try to pull it out of
	// the kubeconfig context
	if apiToken == "" {
		kubeconfig := viper.GetString("kubeConfig")
		authContext, err := auth.CurrentAuthContext(kubeconfig)
		// if there is an error, no kubeconfig exists for us ... this is not really an error
		// if metalctl is used in scripting with an hmac-key
		if err == nil {
			apiToken = authContext.IDToken
		}
	}
	return apiToken
}
