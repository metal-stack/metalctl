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

	"github.com/stretchr/testify/mock"
)

func Test_ImageListCmd(t *testing.T) {
	tests := []struct {
		name       string
		metalMocks *client.MetalMockFns
		want       []*models.V1ImageResponse
		wantTable  string
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
								Usedby:   []string{},
							},
							{
								Features: []string{"machine"},
								ID:       pointer.Pointer("debian"),
								Name:     "debian",
								Usedby:   []string{},
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
			wantTable: `
ID       NAME     DESCRIPTION   FEATURES   EXPIRATION   STATUS   USEDBY
debian   debian                 machine                          0
ubuntu   ubuntu                 machine                          0
`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for _, format := range outputFormats(tt.want, tt.wantTable) {
				format := format
				t.Run(format.Name(), func(t *testing.T) {
					var out bytes.Buffer
					config, mock := newTestConfig(t, &out, tt.metalMocks)

					cmd := newRootCmd(config)
					os.Args = []string{binaryName, "image", "list", "-o", format.Name()}

					err := cmd.Execute()
					if diff := cmp.Diff(tt.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
						t.Errorf("error diff (+got -want):\n %s", diff)
					}

					format.Validate(t, out.Bytes())

					mock.AssertExpectations(t)
				})
			}
		})
	}
}
