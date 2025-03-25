package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/metal-stack/metal-go/api/client/version"
	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/metal-stack/v"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newLoginCmd(c *api.Config) *cobra.Command {
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
					fmt.Fprintln(c.Out, tokenInfo.IDToken)
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
				Log:          c.Log,
			}

			fmt.Fprintln(c.Out)

			err := auth.OIDCFlow(config)
			if err != nil {
				return err
			}

			resp, err := c.Client.Version().Info(version.NewInfoParams(), clientNoAuth())
			if err != nil {
				return err
			}
			if resp.Payload != nil && resp.Payload.MinClientVersion != nil {
				minVersion := *resp.Payload.MinClientVersion
				parsedMinVersion, err := semver.NewVersion(minVersion)
				if err != nil {
					return fmt.Errorf("required metalctl minimum version:%q is not semver parsable:%w", minVersion, err)
				}

				// This is a developer build
				if !strings.HasPrefix(v.Version, "v") {
					return nil
				}

				thisVersion, err := semver.NewVersion(v.Version)
				if err != nil {
					return fmt.Errorf("metalctl version:%q is not semver parsable:%w", v.Version, err)
				}

				if thisVersion.LessThan(parsedMinVersion) {
					return fmt.Errorf("your metalctl version:%s is smaller than the required minimum version:%s, please run `metalctl update do` to get this version", thisVersion, minVersion)
				}

				if !thisVersion.Equal(parsedMinVersion) {
					fmt.Fprintln(c.Out)
					fmt.Fprintf(c.Out, "WARNING: Your metalctl version %q might not compatible with the metal-api (supported version is %q). Please run `metalctl update do` to update to the supported version.", thisVersion, minVersion)
					fmt.Fprintln(c.Out)
				}
			}

			return nil

		},
	}
	loginCmd.Flags().Bool("print-only", false, "If true, the token is printed to stdout")
	return loginCmd
}
