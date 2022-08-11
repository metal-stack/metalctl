package cmd

import (
	"bytes"
	"os"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/client/version"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/metal-stack/metalctl/pkg/api"
	"github.com/metal-stack/v"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func Test_VersionCmd(t *testing.T) {
	tests := []struct {
		name       string
		metalMocks *client.MetalMockFns
		want       *api.Version
		wantErr    error
	}{
		{
			name: "print version",
			metalMocks: &client.MetalMockFns{
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			config, mock := newTestConfig(t, &out, tt.metalMocks)

			v.Version = "client v1.0.0"

			cmd := newRootCmd(config)
			os.Args = []string{binaryName, "version"}

			err := cmd.Execute()
			if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n %s", diff)
			}

			var got *api.Version
			err = yaml.Unmarshal(out.Bytes(), &got)
			require.NoError(t, err, out.String())

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}

			mock.AssertExpectations(t)
		})
	}
}
