package cmd

import (
	"fmt"
	"time"

	"github.com/metal-stack/metal-lib/jwt/sec"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newWhoamiCmd(c *api.Config) *cobra.Command {
	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "shows current user",
		Long:  "shows the current user, that will be used to authenticate commands.",
		RunE: func(cmd *cobra.Command, args []string) error {

			authContext, err := getAuthContext(viper.GetString("kubeconfig"))
			if err != nil {
				return err
			}

			if !authContext.AuthProviderOidc {
				return fmt.Errorf("active user %s has no oidc authProvider, check config", authContext.User)
			}

			user, parsedClaims, err := sec.ParseTokenUnvalidatedUnfiltered(authContext.IDToken)
			if err != nil {
				return err
			}

			fmt.Fprintf(c.Out, "UserId: %s\n", user.Name)
			if user.Tenant != "" {
				fmt.Fprintf(c.Out, "Tenant: %s\n", user.Tenant)
			}
			if user.Issuer != "" {
				fmt.Fprintf(c.Out, "Issuer: %s\n", user.Issuer)
			}
			fmt.Fprintf(c.Out, "Groups:\n")
			for _, g := range user.Groups {
				fmt.Fprintf(c.Out, " %s\n", g)
			}

			fmt.Fprintf(c.Out, "Expires at %s\n", time.Unix(parsedClaims.ExpiresAt, 0).Format("Mon Jan 2 15:04:05 MST 2006"))

			return nil
		},
	}
	return whoamiCmd
}
