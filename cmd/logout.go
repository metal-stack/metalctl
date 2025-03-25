package cmd

import (
	"fmt"

	"github.com/metal-stack/metal-lib/auth"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
)

func newLogoutCmd(c *api.Config) *cobra.Command {
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "logout user from OIDC SSO session",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := api.MustDefaultContext()

			err := auth.Logout(&auth.LogoutParams{
				IssuerURL: ctx.IssuerURL,
				Logger:    c.Log,
			})
			if err != nil {
				return err
			}

			fmt.Fprintln(c.Out, "OIDC session successfully logged out. Token is not revoked and is valid until expiration.")

			return nil
		},
	}
	return logoutCmd
}
