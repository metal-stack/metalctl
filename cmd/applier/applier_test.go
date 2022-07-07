package applier

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/api/models"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var (
	projectCreateRequests = `---
name: test-a
description: Test Project A
tenantid: tenant-a
meta:
    apiversion: v1
    kind: Project
    id: 00000000-0000-0000-0000-000000000000
    annotations:
        annotation: a
    labels:
        - a
---
name: test-b
description: Test Project B
tenantid: tenant-b
meta:
    apiversion: v1
    kind: Project
    id: 00000000-0000-0000-0000-000000000001
    annotations:
        annotation: b
    labels:
        - b
`
)

func Test_readYAML(t *testing.T) {
	const testFile = "/test.yaml"

	tests := []struct {
		name    string
		mockFn  func(fs afero.Fs)
		want    []*models.V1ProjectCreateRequest
		wantErr error
	}{
		{
			name: "parsing empty file",
			mockFn: func(fs afero.Fs) {
				require.NoError(t, afero.WriteFile(fs, testFile, []byte(""), 0755))
			},
			want: nil,
		},
		{
			name: "parsing multi-document yaml",
			mockFn: func(fs afero.Fs) {
				require.NoError(t, afero.WriteFile(fs, testFile, []byte(projectCreateRequests), 0755))
			},
			want: []*models.V1ProjectCreateRequest{
				{
					Name:        "test-a",
					Description: "Test Project A",
					Meta: &models.V1Meta{
						Apiversion:  "v1",
						ID:          "00000000-0000-0000-0000-000000000000",
						Kind:        "Project",
						Annotations: map[string]string{"annotation": "a"},
						Labels:      []string{"a"},
					},
					TenantID: "tenant-a",
				},
				{
					Name:        "test-b",
					Description: "Test Project B",
					Meta: &models.V1Meta{
						Apiversion:  "v1",
						ID:          "00000000-0000-0000-0000-000000000001",
						Kind:        "Project",
						Annotations: map[string]string{"annotation": "b"},
						Labels:      []string{"b"},
					},
					TenantID: "tenant-b",
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.mockFn != nil {
				tt.mockFn(fs)
			}

			got, err := readYAML[*models.V1ProjectCreateRequest](fs, testFile)

			if diff := cmp.Diff(tt.wantErr, err); diff != "" {
				t.Errorf("error diff (+got -want):\n %s", diff)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}
		})
	}
}

func Test_readYAMLIndex(t *testing.T) {
	const testFile = "/test.yaml"

	tests := []struct {
		name    string
		mockFn  func(fs afero.Fs)
		index   int
		want    *models.V1ProjectCreateRequest
		wantErr error
	}{
		{
			name: "request zero index",
			mockFn: func(fs afero.Fs) {
				require.NoError(t, afero.WriteFile(fs, testFile, []byte(projectCreateRequests), 0755))
			},
			index: 0,
			want: &models.V1ProjectCreateRequest{
				Name:        "test-a",
				Description: "Test Project A",
				Meta: &models.V1Meta{
					Apiversion:  "v1",
					ID:          "00000000-0000-0000-0000-000000000000",
					Kind:        "Project",
					Annotations: map[string]string{"annotation": "a"},
					Labels:      []string{"a"},
				},
				TenantID: "tenant-a",
			},
		},
		{
			name: "request one index",
			mockFn: func(fs afero.Fs) {
				require.NoError(t, afero.WriteFile(fs, testFile, []byte(projectCreateRequests), 0755))
			},
			index: 1,
			want: &models.V1ProjectCreateRequest{
				Name:        "test-b",
				Description: "Test Project B",
				Meta: &models.V1Meta{
					Apiversion:  "v1",
					ID:          "00000000-0000-0000-0000-000000000001",
					Kind:        "Project",
					Annotations: map[string]string{"annotation": "b"},
					Labels:      []string{"b"},
				},
				TenantID: "tenant-b",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.mockFn != nil {
				tt.mockFn(fs)
			}

			got, err := readYAMLIndex[*models.V1ProjectCreateRequest](fs, testFile, tt.index)

			if diff := cmp.Diff(tt.wantErr, err); diff != "" {
				t.Errorf("error diff (+got -want):\n %s", diff)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}
		})
	}
}
