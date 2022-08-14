package cmd

import (
	"testing"

	"github.com/metal-stack/metal-go/api/client/health"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/metal-stack/metal-lib/rest"
	"github.com/stretchr/testify/mock"
)

func Test_HealthCmd(t *testing.T) {
	tests := []*test[*rest.HealthResponse]{
		{
			name: "health",
			cmd: func(want *rest.HealthResponse) []string {
				return []string{"health"}
			},
			mocks: &client.MetalMockFns{
				Health: func(mock *mock.Mock) {
					mock.On("Health", testcommon.MatchIgnoreContext(t, health.NewHealthParams()), nil).Return(&health.HealthOK{
						Payload: &models.RestHealthResponse{
							Status:   pointer.Pointer(string(rest.HealthStatusHealthy)),
							Message:  pointer.Pointer("ok"),
							Services: make(map[string]models.RestHealthResult),
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
			name: "health on error response",
			cmd: func(want *rest.HealthResponse) []string {
				return []string{"health"}
			},
			mocks: &client.MetalMockFns{
				Health: func(mock *mock.Mock) {
					mock.On("Health", testcommon.MatchIgnoreContext(t, health.NewHealthParams()), nil).Return(nil, &health.HealthInternalServerError{
						Payload: &models.RestHealthResponse{
							Status:   pointer.Pointer(string(rest.HealthStatusUnhealthy)),
							Message:  pointer.Pointer("error"),
							Services: make(map[string]models.RestHealthResult),
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
		tt.testCmd(t)
	}
}
