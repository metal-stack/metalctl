package cmd

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
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

type outputsFormatConfig[R any] struct {
	want           R       // for json and yaml
	table          *string // for table printer
	template       *string // for template printer
	templateOutput *string // for template printer
	markdownTable  *string // for markdown printer
}

func outputFormats[R any](c *outputsFormatConfig[R]) []outputFormat[R] {
	var formats []outputFormat[R]

	if !pointer.IsZero(c.want) {
		formats = append(formats, &jsonOutputFormat[R]{want: c.want}, &yamlOutputFormat[R]{want: c.want})
	}

	if c.table != nil {
		formats = append(formats, &tableOutputFormat[R]{table: *c.table})
	}

	if c.template != nil && c.templateOutput != nil {
		formats = append(formats, &templateOutputFormat[R]{template: *c.template, templateOutput: *c.templateOutput})
	}

	if c.markdownTable != nil {
		formats = append(formats, &markdownOutputFormat[R]{table: *c.markdownTable})
	}

	return formats
}

type outputFormat[R any] interface {
	Args() []string
	Validate(t *testing.T, output []byte)
}

type jsonOutputFormat[R any] struct {
	want R
}

func (o *jsonOutputFormat[R]) Args() []string {
	return []string{"-o", "json"}
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

func (o *yamlOutputFormat[R]) Args() []string {
	return []string{"-o", "yaml"}
}

func (o *yamlOutputFormat[R]) Validate(t *testing.T, output []byte) {
	var got R
	err := yaml.Unmarshal(output, &got)
	require.NoError(t, err)

	if diff := cmp.Diff(o.want, got); diff != "" {
		t.Errorf("diff (+got -want):\n %s", diff)
	}
}

type tableOutputFormat[R any] struct {
	table string
}

func (o *tableOutputFormat[R]) Args() []string {
	return []string{"-o", "table"}
}

func (o *tableOutputFormat[R]) Validate(t *testing.T, output []byte) {
	validateTableRows(t, o.table, string(output))
}

type templateOutputFormat[R any] struct {
	template       string
	templateOutput string
}

func (o *templateOutputFormat[R]) Args() []string {
	return []string{"-o", "template", "--template", o.template}
}

func (o *templateOutputFormat[R]) Validate(t *testing.T, output []byte) {
	t.Logf("got following template output:\n%s\n", string(output))

	if diff := cmp.Diff(strings.TrimSpace(o.templateOutput), strings.TrimSpace(string(output))); diff != "" {
		t.Errorf("diff (+got -want):\n %s", diff)
	}
}

type markdownOutputFormat[R any] struct {
	table string
}

func (o *markdownOutputFormat[R]) Args() []string {
	return []string{"-o", "markdown"}
}

func (o *markdownOutputFormat[R]) Validate(t *testing.T, output []byte) {
	validateTableRows(t, o.table, string(output))
}

func validateTableRows(t *testing.T, want, got string) {
	trimAll := func(ss []string) []string {
		var res []string
		for _, s := range ss {
			res = append(res, strings.TrimSpace(s))
		}
		return res
	}

	var (
		trimmedWant = strings.TrimSpace(want)
		trimmedGot  = strings.TrimSpace(string(got))

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
