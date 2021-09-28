package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newLoginCmd(c *config) *cobra.Command {

	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "login user and receive token",
		Long:  "login and receive token that will be used to authenticate commands.",
		RunE: func(cmd *cobra.Command, args []string) error {

			var console io.Writer
			var handler auth.TokenHandlerFunc
			if viper.GetBool("print-only") {
				// do not print to console
				handler = printTokenHandler
			} else {
				cs, err := api.GetContexts()
				if err != nil {
					return err
				}
				console = os.Stdout
				handler = auth.NewUpdateKubeConfigHandler(viper.GetString("kubeconfig"), console, auth.WithContextName(formatContextName(cloudContext, cs.CurrentContext)))
			}

			ctx := api.MustDefaultContext()
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
	loginCmd.Flags().Bool("print-only", false, "If true, the token is printed to stdout")
	return loginCmd
}

func printTokenHandler(tokenInfo auth.TokenInfo) error {
	fmt.Println(tokenInfo.IDToken)
	return nil
}
