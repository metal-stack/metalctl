package cmd

import (
	"runtime"
	"testing"

	"github.com/metal-stack/metal-go/api/client/version"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/metal-stack/v"
	"github.com/stretchr/testify/mock"
)

func Test_VersionCmd(t *testing.T) {
	tests := []*test[*api.Version]{
		{
			name: "version",
			cmd: func(want *api.Version) []string {
				return []string{"version"}
			},
			mocks: &client.MetalMockFns{
				Version: func(mock *mock.Mock) {
					mock.On("Info", testcommon.MatchIgnoreContext(t, version.NewInfoParams()), nil).Return(&version.InfoOK{
						Payload: &models.RestVersion{
							Version: pointer.Pointer("server v1.0.0"),
						},
					}, nil)
				},
			},
			want: &api.Version{
				Client: "client v1.0.0, " + runtime.Version(),
				Server: &models.RestVersion{
					Version: pointer.Pointer("server v1.0.0"),
				},
			},
		},
	}
	for _, tt := range tests {
		v.Version = "client v1.0.0"
		tt.testCmd(t)
	}
}
