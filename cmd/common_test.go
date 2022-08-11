package cmd

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gopkg.in/yaml.v3"
)

func newTestConfig(t *testing.T, out io.Writer, mocks *client.MetalMockFns) (*config, *client.MetalMockClient) {
	mock, c := client.NewMetalMockClient(mocks)
	return &config{
		fs:     afero.NewMemMapFs(),
		out:    out,
		client: c,
		log:    zaptest.NewLogger(t).Sugar(),
	}, mock
}

func outputFormats[R any]() []outputFormat[R] {
	return []outputFormat[R]{
		&jsonOutputFormat[R]{},
		&yamlOutputFormat[R]{},
	}
}

type outputFormat[R any] interface {
	Name() string
	Validate(t *testing.T, output []byte, want R)
}

type jsonOutputFormat[R any] struct{}

func (o *jsonOutputFormat[R]) Name() string {
	return "json"
}

func (o *jsonOutputFormat[R]) Validate(t *testing.T, output []byte, want R) {
	var got R
	err := json.Unmarshal(output, &got)
	require.NoError(t, err, string(output))

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("diff (+got -want):\n %s", diff)
	}
}

type yamlOutputFormat[R any] struct{}

func (o *yamlOutputFormat[R]) Name() string {
	return "yaml"
}

func (o *yamlOutputFormat[R]) Validate(t *testing.T, output []byte, want R) {
	var got R
	err := yaml.Unmarshal(output, &got)
	require.NoError(t, err, string(output))

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("diff (+got -want):\n %s", diff)
	}
}
