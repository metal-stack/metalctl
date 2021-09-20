package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/metal-stack/metal-lib/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login user and receive token",
	Long:  "login and receive token that will be used to authenticate commands.",
	RunE: func(cmd *cobra.Command, args []string) error {

		var console io.Writer
		var handler auth.TokenHandlerFunc
		if viper.GetBool("printOnly") {
			// do not print to console
			handler = printTokenHandler
		} else {
			cs, err := getContexts()
			if err != nil {
				return err
			}
			console = os.Stdout
			handler = auth.NewUpdateKubeConfigHandler(viper.GetString("kubeconfig"), console, auth.WithContextName(formatContextName(cloudContext, cs.CurrentContext)))
		}

		scopes := auth.DexScopes
		if ctx.IssuerType == "generic" {
			scopes = auth.GenericScopes
		} else if ctx.CustomScopes != "" {
			cs := strings.Split(ctx.CustomScopes, ",")
			for i := range cs {
				cs[i] = strings.TrimSpace(cs[i])
			}
			scopes = cs
		}

		config := auth.Config{
			ClientID:     ctx.ClientID,
			ClientSecret: ctx.ClientSecret,
			IssuerURL:    ctx.IssuerURL,
			Scopes:       scopes,
			TokenHandler: handler,
			Console:      console,
			Debug:        viper.GetBool("debug"),
		}

		fmt.Println()

		return auth.OIDCFlow(config)
	},
	PreRun: bindPFlags,
}

func printTokenHandler(tokenInfo auth.TokenInfo) error {

	fmt.Println(tokenInfo.IDToken)
	return nil
}

func init() {

	loginCmd.Flags().Bool("printOnly", false, "If true, the token is printed to stdout")
}
