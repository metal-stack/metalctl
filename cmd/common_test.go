package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/exp/slices"

	"bou.ke/monkey"
	"github.com/google/go-cmp/cmp"
	"github.com/metal-stack/metal-go/test/client"
	"github.com/metal-stack/metal-lib/pkg/pointer"
	"github.com/metal-stack/metal-lib/pkg/testcommon"
	"github.com/metal-stack/metalctl/cmd/completion"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gopkg.in/yaml.v3"
)

var testTime = time.Date(2022, time.May, 19, 1, 2, 3, 4, time.UTC)

func init() {
	_ = monkey.Patch(time.Now, func() time.Time { return testTime })
}

type test[R any] struct {
	name    string
	mocks   *client.MetalMockFns
	fsMocks func(fs afero.Fs, want R)
	cmd     func(want R) []string

	disableMockClient bool // can switch off mock client creation

	wantErr       error
	want          R       // for json and yaml
	wantTable     *string // for table printer
	wantWideTable *string // for wide table printer
	template      *string // for template printer
	wantTemplate  *string // for template printer
	wantMarkdown  *string // for markdown printer
}

func (c *test[R]) testCmd(t *testing.T) {
	require.NotEmpty(t, c.name, "test name must not be empty")
	require.NotEmpty(t, c.cmd, "cmd must not be empty")

	if c.wantErr != nil {
		mock, _, config := c.newMockConfig(t)

		cmd := newRootCmd(config)
		os.Args = append([]string{binaryName}, c.cmd(c.want)...)

		err := cmd.Execute()
		if diff := cmp.Diff(c.wantErr, err, testcommon.ErrorStringComparer()); diff != "" {
			t.Errorf("error diff (+got -want):\n %s", diff)
		}

		mock.AssertExpectations(t)
	}

	for _, format := range outputFormats(c) {
		format := format
		t.Run(fmt.Sprintf("%v", format.Args()), func(t *testing.T) {
			mock, out, config := c.newMockConfig(t)

			cmd := newRootCmd(config)
			os.Args = append([]string{binaryName}, c.cmd(c.want)...)
			os.Args = append(os.Args, format.Args()...)

			err := cmd.Execute()
			assert.NoError(t, err)

			format.Validate(t, out.Bytes())

			mock.AssertExpectations(t)
		})
	}
}

func (c *test[R]) newMockConfig(t *testing.T) (*client.MetalMockClient, *bytes.Buffer, *config) {
	mock, client := client.NewMetalMockClient(c.mocks)

	fs := afero.NewMemMapFs()
	if c.fsMocks != nil {
		c.fsMocks(fs, c.want)
	}

	var (
		out    bytes.Buffer
		config = &config{
			fs:     fs,
			client: client,
			out:    &out,
			log:    zaptest.NewLogger(t).Sugar(),
			comp:   &completion.Completion{},
		}
	)

	if c.disableMockClient {
		config.client = nil
	}

	return mock, &out, config
}

func assertExhaustiveArgs(t *testing.T, args []string, exclude ...string) {
	assertContainsPrefix := func(ss []string, prefix string) error {
		for _, s := range ss {
			if strings.HasPrefix(s, prefix) {
				return nil
			}
		}
		return fmt.Errorf("not exhaustive: does not contain " + prefix)
	}

	root := newRootCmd(&config{comp: &completion.Completion{}})
	cmd, args, err := root.Find(args)
	require.NoError(t, err)

	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if slices.Contains(exclude, f.Name) {
			return
		}
		assert.NoError(t, assertContainsPrefix(args, "--"+f.Name), "please ensure you all available args are used in order to increase coverage or exclude them explicitly")
	})
}

func mustMarshal(t *testing.T, d any) []byte {
	b, err := json.MarshalIndent(d, "", "    ")
	require.NoError(t, err)
	return b
}

func mustMarshalToMultiYAML[R any](t *testing.T, data []R) []byte {
	var parts []string
	for _, elem := range data {
		parts = append(parts, string(mustMarshal(t, elem)))
	}
	return []byte(strings.Join(parts, "\n---\n"))
}

func mustJsonDeepCopy[O any](t *testing.T, object O) O {
	raw, err := json.Marshal(&object)
	require.NoError(t, err)
	var copy O
	err = json.Unmarshal(raw, &copy)
	require.NoError(t, err)
	return copy
}

func outputFormats[R any](c *test[R]) []outputFormat[R] {
	var formats []outputFormat[R]

	if !pointer.IsZero(c.want) {
		formats = append(formats, &jsonOutputFormat[R]{want: c.want}, &yamlOutputFormat[R]{want: c.want})
	}

	if c.wantTable != nil {
		formats = append(formats, &tableOutputFormat[R]{table: *c.wantTable})
	}

	if c.wantWideTable != nil {
		formats = append(formats, &wideTableOutputFormat[R]{table: *c.wantWideTable})
	}

	if c.template != nil && c.wantTemplate != nil {
		formats = append(formats, &templateOutputFormat[R]{template: *c.template, templateOutput: *c.wantTemplate})
	}

	if c.wantMarkdown != nil {
		formats = append(formats, &markdownOutputFormat[R]{table: *c.wantMarkdown})
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

	if diff := cmp.Diff(o.want, got, testcommon.StrFmtDateComparer()); diff != "" {
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

	if diff := cmp.Diff(o.want, got, testcommon.StrFmtDateComparer()); diff != "" {
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

type wideTableOutputFormat[R any] struct {
	table string
}

func (o *wideTableOutputFormat[R]) Args() []string {
	return []string{"-o", "wide"}
}

func (o *wideTableOutputFormat[R]) Validate(t *testing.T, output []byte) {
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
	t.Logf("got following template output:\n\n%s\n\nconsider using this for test comparison if it looks correct.", string(output))

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

	t.Logf("got following table output:\n\n%s\n\nconsider using this for test comparison if it looks correct.", trimmedGot)

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
