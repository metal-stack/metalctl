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
	// "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/api/go/client"
	"github.com/metal-stack/metal-lib/pkg/genericcli"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type login struct {
	c client.Client
}

func NewLoginCmd(c client.Client) *cobra.Command {
	l := &login{c: c}
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "login",
		RunE: func(cmd *cobra.Command, args []string) error {
			return l.login()
		},
	}

	loginCmd.Flags().String("provider", "oidc", "the provider used to login with")
	loginCmd.Flags().String("context-name", "", "the context into which the token gets injected, if not specified it uses the current context or creates a context named default in case there is no current context set")

	genericcli.Must(loginCmd.RegisterFlagCompletionFunc("provider", cobra.FixedCompletions([]string{"github", "azure", "google"}, cobra.ShellCompDirectiveNoFileComp)))

	return loginCmd
}

func (l *login) login() error {
	ctxs, err := api.GetContexts()
	if err != nil {
		return err
	}

	provider := viper.GetString("provider")
	if provider == "" {
		return errors.New("provider must be specified")
	}

	ctxName := ctxs.CurrentContext
	if viper.IsSet("context-name") {
		ctxName = viper.GetString("context-name")
	}

	ctx, ok := ctxs.Get(ctxName)
	if !ok {
		defaultCtx := api.MustDefaultContext()
		defaultCtxName := "default"

		ctxs.PreviousContext = ctxs.CurrentContext
		ctxs.CurrentContext = defaultCtxName

		ctxs.Contexts[defaultCtxName] = defaultCtx

		ctx = &defaultCtx
	}

	if ctx.V2Config == nil || ctx.V2Config.ApiURL == nil {
		return errors.New("no proper v2 config found, please ensure your current context contains v2/api-url")
	}

	tokenChan := make(chan string)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		tokenChan <- r.URL.Query().Get("token")

		http.Redirect(w, r, "https://metal-stack.io", http.StatusSeeOther)
	})

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

	url := fmt.Sprintf("%s/auth/%s?redirect-url=http://%s/callback", *ctx.V2Config.ApiURL, provider, listener.Addr().String())
	fmt.Printf("open %q in browser", url)

	err = exec.Command("xdg-open", url).Run() //nolint
	if err != nil {
		return fmt.Errorf("error opening browser: %w", err)
	}

	token := <-tokenChan

	err = server.Shutdown(context.Background())
	if err != nil {
		return fmt.Errorf("unable to close http server: %w", err)
	}
	_ = listener.Close()

	if token == "" {
		return errors.New("no token was retrieved")
	}

	ctx.V2Config.Token = token

	// if ctx.DefaultProject == "" {
	// 	projects, err := l.c.Client.Apiv1().Project().List(context.Background(), connect.NewRequest(&apiv2.ProjectServiceListRequest{}))
	// 	if err != nil {
	// 		return fmt.Errorf("unable to retrieve project list: %w", err)
	// 	}

	// 	idx := slices.IndexFunc(projects.Msg.Projects, func(p *apiv1.Project) bool {
	// 		return p.IsDefaultProject
	// 	})

	// 	if idx >= 0 {
	// 		ctx.DefaultProject = projects.Msg.Projects[idx].Uuid
	// 	}
	// }

	err = api.WriteContexts(ctxs)
	if err != nil {
		return err
	}

	fmt.Printf("%s login successful! Updated and activated context \n", color.GreenString("✔"))

	return nil
}
