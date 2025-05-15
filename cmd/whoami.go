package cmd

import (
	"fmt"
	"time"

	"github.com/metal-stack/metal-lib/jwt/sec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newWhoamiCmd(c *config) *cobra.Command {
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

			_, _ = fmt.Fprintf(c.out, "UserId: %s\n", user.Name)
			if user.Tenant != "" {
				_, _ = fmt.Fprintf(c.out, "Tenant: %s\n", user.Tenant)
			}
			if user.Issuer != "" {
				_, _ = fmt.Fprintf(c.out, "Issuer: %s\n", user.Issuer)
			}
			_, _ = fmt.Fprintf(c.out, "Groups:\n")
			for _, g := range user.Groups {
				_, _ = fmt.Fprintf(c.out, " %s\n", g)
			}

			_, _ = fmt.Fprintf(c.out, "Expires at %s\n", time.Unix(parsedClaims.ExpiresAt, 0).Format("Mon Jan 2 15:04:05 MST 2006"))

			return nil
		},
	}
	return whoamiCmd
}
