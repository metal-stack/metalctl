package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/client/health"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/metal-stack/metal-lib/rest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_HealthCmd(t *testing.T) {
	tests := []struct {
		name       string
		metalMocks *client.MetalMockFns
		want       *rest.HealthResponse
		wantErr    error
	}{
		{
			name: "print health",
			metalMocks: &client.MetalMockFns{
				Health: func(mock *mock.Mock) {
					mock.On("Health", testcommon.MatchIgnoreContext(t, health.NewHealthParams()), nil).Return(&health.HealthOK{
						Payload: &models.RestHealthResponse{
							Status:  pointer.Pointer(string(rest.HealthStatusHealthy)),
							Message: pointer.Pointer("ok"),
						},
					}, nil)
				},
			},
			want: &rest.HealthResponse{
				Status:   rest.HealthStatusHealthy,
				Message:  "ok",
				Services: make(map[string]rest.HealthResult),
			},
		},
		{
			name: "print health also on error response",
			metalMocks: &client.MetalMockFns{
				Health: func(mock *mock.Mock) {
					mock.On("Health", testcommon.MatchIgnoreContext(t, health.NewHealthParams()), nil).Return(nil, &health.HealthInternalServerError{
						Payload: &models.RestHealthResponse{
							Status:  pointer.Pointer(string(rest.HealthStatusUnhealthy)),
							Message: pointer.Pointer("error"),
						},
					})
				},
			},
			want: &rest.HealthResponse{
				Status:   rest.HealthStatusUnhealthy,
				Message:  "error",
				Services: make(map[string]rest.HealthResult),
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			config, mock := newTestConfig(t, &out, tt.metalMocks)

			cmd := newRootCmd(config)
			os.Args = []string{binaryName, "health"}

			err := cmd.Execute()
			if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n %s", diff)
			}

			var got *rest.HealthResponse
			err = yaml.Unmarshal(out.Bytes(), &got)
			require.NoError(t, err, out.String())

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}

			mock.AssertExpectations(t)
		})
	}
}
