package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/rest"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_BasicRootCmdStuff(t *testing.T) {
	// prevent env variables for metalctl being set from the outside, which could cause bad side-effects
	// for these tests as the mock client gets disabled (to point it to the test HTTP server instead)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, strings.ToUpper(binaryName)+"_") {
			t.Setenv(strings.Split(env, "=")[0], "")
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer") {
			assert.Equal(t, "Bearer i-am-token", authHeader)
		} else if strings.HasPrefix(authHeader, "Metal-Admin") {
			assert.Len(t, strings.Split(authHeader, " "), 2)
		} else {
			assert.Fail(t, "missing auth header")
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(mustMarshal(t, &models.RestHealthResponse{
			Status: pointer.Pointer(string(rest.HealthStatusHealthy)),
		}))
		if err != nil {
			t.Errorf("error writing response: %s", err)
		}
	}))
	defer ts.Close()

	tests := []*test[*rest.HealthResponse]{
		{
			name: "overwrite api-url and api-token from config-file",
			fsMocks: func(fs afero.Fs, want *rest.HealthResponse) {
				require.NoError(t, afero.WriteFile(fs, fmt.Sprintf("/etc/%s/config.yaml", binaryName), []byte(fmt.Sprintf(`---
api-url: "%s"
api-token: "i-am-token"
`, ts.URL)), 0755))
			},
			cmd: func(want *rest.HealthResponse) []string {
				return []string{"health"}
			},
			disableMockClient: true,
			want: &rest.HealthResponse{
				Status: rest.HealthStatusHealthy,
			},
		},
		{
			name: "overwrite api-url and api-token from user-given config-file path",
			fsMocks: func(fs afero.Fs, want *rest.HealthResponse) {
				require.NoError(t, afero.WriteFile(fs, "/config.yaml", []byte(fmt.Sprintf(`---
api-url: "%s"
api-token: "i-am-token"
`, ts.URL)), 0755))
			},
			cmd: func(want *rest.HealthResponse) []string {
				return []string{"health", "--config", "/config.yaml"}
			},
			disableMockClient: true,
			want: &rest.HealthResponse{
				Status: rest.HealthStatusHealthy,
			},
		},
		{
			name: "overwrite api-url and api-token from command line",
			cmd: func(want *rest.HealthResponse) []string {
				return []string{"health", "--api-url", ts.URL, "--api-token", "i-am-token"}
			},
			disableMockClient: true,
			want: &rest.HealthResponse{
				Status: rest.HealthStatusHealthy,
			},
		},
		{
			name: "overwrite api-url and api-token from environment",
			cmd: func(want *rest.HealthResponse) []string {
				t.Setenv("METALCTL_API_URL", ts.URL)
				t.Setenv("METALCTL_API_TOKEN", "i-am-token")
				return []string{"health"}
			},
			disableMockClient: true,
			want: &rest.HealthResponse{
				Status: rest.HealthStatusHealthy,
			},
		},
		{
			name: "use hmac",
			cmd: func(want *rest.HealthResponse) []string {
				t.Setenv("METALCTL_HMAC", "i-am-hmac")
				return []string{"health"}
			},
			disableMockClient: true,
			want: &rest.HealthResponse{
				Status: rest.HealthStatusHealthy,
			},
		},
	}
	for _, tt := range tests {
		tt.testCmd(t)
	}
}
