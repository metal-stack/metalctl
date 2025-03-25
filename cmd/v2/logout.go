package v2

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/fatih/color"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type logout struct {
	c *api.Config
}

func newLogoutCmd(c *api.Config) *cobra.Command {
	w := &logout{
		c: c,
	}

	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "logout",
		RunE: func(cmd *cobra.Command, args []string) error {
			return w.logout()
		},
	}

	logoutCmd.Flags().String("provider", "oidc", "the provider used to logout with")
	logoutCmd.Flags().String("context-name", "", "the context into which the token gets injected, if not specified it uses the current context or creates a context named default in case there is no current context set")

	genericcli.Must(logoutCmd.RegisterFlagCompletionFunc("provider", cobra.FixedCompletions([]string{"oidc"}, cobra.ShellCompDirectiveNoFileComp)))

	return logoutCmd
}

func (l *logout) logout() error {
	provider := viper.GetString("provider")
	if provider == "" {
		return errors.New("provider must be specified")
	}

	// ctxs, err := l.c.GetContexts()
	// if err != nil {
	// 	return err
	// }

	// ctxName := ctxs.CurrentContext
	// if viper.IsSet("context-name") {
	// 	ctxName = viper.GetString("context-name")
	// }

	// ctx, ok := ctxs.Get(ctxName)
	// if !ok {
	// 	defaultCtx := l.c.MustDefaultContext()
	// 	defaultCtx.Name = "default"

	// 	ctxs.PreviousContext = ctxs.CurrentContext
	// 	ctxs.CurrentContext = ctx.Name

	// 	ctxs.Contexts = append(ctxs.Contexts, &defaultCtx)

	// 	ctx = &defaultCtx
	// }

	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}

	server := http.Server{Addr: listener.Addr().String(), ReadTimeout: 2 * time.Second}

	go func() {
		fmt.Printf("Starting server at http://%s...\n", listener.Addr().String())
		err = server.Serve(listener) //nolint
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(fmt.Errorf("http server closed unexpectedly: %w", err))
		}
	}()

	url := fmt.Sprintf("%s/auth/logout/%s", l.c.ApiV2URL, provider)

	err = exec.Command("xdg-open", url).Run() //nolint
	if err != nil {
		return fmt.Errorf("error opening browser: %w", err)
	}

	err = server.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("unable to close http server: %w", err)
	}
	_ = listener.Close()

	fmt.Fprintf(l.c.Out, "%s logout successful! \n", color.GreenString("✔"))

	return nil
}
