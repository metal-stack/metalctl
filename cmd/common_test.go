package cmd

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gopkg.in/yaml.v3"
)

func newTestConfig(t *testing.T, out io.Writer, mocks *client.MetalMockFns, fsMocks func(fs afero.Fs)) (*config, *client.MetalMockClient) {
	mock, c := client.NewMetalMockClient(mocks)

	fs := afero.NewMemMapFs()
	if fsMocks != nil {
		fsMocks(fs)
	}

	return &config{
		fs:     fs,
		out:    out,
		client: c,
		log:    zaptest.NewLogger(t).Sugar(),
	}, mock
}

func mustMarshal(t *testing.T, d any) []byte {
	b, err := json.Marshal(d)
	require.NoError(t, err)
	return b
}

func outputFormats[R any](want R, table string) []outputFormat[R] {
	return []outputFormat[R]{
		&tableOutputFormat[R]{table: table},
		&jsonOutputFormat[R]{want: want},
		&yamlOutputFormat[R]{want: want},
	}
}

type outputFormat[R any] interface {
	Name() string
	Validate(t *testing.T, output []byte)
}

type jsonOutputFormat[R any] struct {
	want R
}

func (o *jsonOutputFormat[R]) Name() string {
	return "json"
}

func (o *jsonOutputFormat[R]) Validate(t *testing.T, output []byte) {
	var got R
	err := json.Unmarshal(output, &got)
	require.NoError(t, err, string(output))

	if diff := cmp.Diff(o.want, got); diff != "" {
		t.Errorf("diff (+got -want):\n %s", diff)
	}
}

type yamlOutputFormat[R any] struct {
	want R
}

func (o *yamlOutputFormat[R]) Name() string {
	return "yaml"
}

func (o *yamlOutputFormat[R]) Validate(t *testing.T, output []byte) {
	var got R
	err := yaml.Unmarshal(output, &got)
	require.NoError(t, err, string(output))

	if diff := cmp.Diff(o.want, got); diff != "" {
		t.Errorf("diff (+got -want):\n %s", diff)
	}
}

type tableOutputFormat[R any] struct {
	table string
}

func (o *tableOutputFormat[R]) Name() string {
	return "table"
}

func (o *tableOutputFormat[R]) Validate(t *testing.T, output []byte) {
	trimAll := func(ss []string) []string {
		var res []string
		for _, s := range ss {
			res = append(res, strings.TrimSpace(s))
		}
		return res
	}

	var (
		trimmedWant = strings.TrimSpace(o.table)
		trimmedGot  = strings.TrimSpace(string(output))

		wantRows = trimAll(strings.Split(trimmedWant, "\n"))
		gotRows  = trimAll(strings.Split(trimmedGot, "\n"))
	)

	t.Logf("got following table output:\n%s\n", trimmedGot)
	t.Log(cmp.Diff(trimmedWant, trimmedGot))

	require.Equal(t, len(wantRows), len(gotRows), "tables have different lengths")

	for i := range wantRows {
		wantFields := trimAll(strings.Split(wantRows[i], " "))
		gotFields := trimAll(strings.Split(gotRows[i], " "))

		require.Equal(t, len(wantFields), len(gotFields), "table fields have different lengths")

		for i := range wantFields {
			assert.Equal(t, wantFields[i], gotFields[i])
		}
	}
}
