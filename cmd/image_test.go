package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/client/image"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_ImageListCmd(t *testing.T) {
	tests := []struct {
		name       string
		metalMocks *client.MetalMockFns
		want       []*models.V1ImageResponse
		wantErr    error
	}{
		{
			name: "list images",
			metalMocks: &client.MetalMockFns{
				Image: func(mock *mock.Mock) {
					mock.On("ListImages", testcommon.MatchIgnoreContext(t, image.NewListImagesParams().WithShowUsage(pointer.Pointer(false))), nil).Return(&image.ListImagesOK{
						Payload: []*models.V1ImageResponse{
							{
								Features: []string{"machine"},
								ID:       pointer.Pointer("ubuntu"),
								Name:     "ubuntu",
							},
							{
								Features: []string{"machine"},
								ID:       pointer.Pointer("debian"),
								Name:     "debian",
							},
						},
					}, nil)
				},
			},
			want: []*models.V1ImageResponse{
				{
					Features: []string{"machine"},
					ID:       pointer.Pointer("debian"),
					Name:     "debian",
					Usedby:   []string{},
				},
				{
					Features: []string{"machine"},
					ID:       pointer.Pointer("ubuntu"),
					Name:     "ubuntu",
					Usedby:   []string{},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			config, mock := newTestConfig(t, &out, tt.metalMocks)

			cmd := newRootCmd(config)
			os.Args = []string{binaryName, "image", "list", "-o", "yaml"}

			err := cmd.Execute()
			if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n %s", diff)
			}

			var got []*models.V1ImageResponse
			err = yaml.Unmarshal(out.Bytes(), &got)
			require.NoError(t, err, out.String())

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}

			mock.AssertExpectations(t)
		})
	}
}
