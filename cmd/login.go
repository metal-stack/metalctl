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
				handler = func(tokenInfo auth.TokenInfo) error {
					fmt.Fprintln(c.out, tokenInfo.IDToken)
					return nil
				}

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
				Log:          c.log.Desugar(),
			}

			fmt.Fprintln(c.out)

			return auth.OIDCFlow(config)
		},
	}
	loginCmd.Flags().Bool("print-only", false, "If true, the token is printed to stdout")
	return loginCmd
}
