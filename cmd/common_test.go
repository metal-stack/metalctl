package cmd

import (
	"io"
	"testing"

	"github.com/metal-stack/metal-go/test/client"
	"github.com/spf13/afero"
	"go.uber.org/zap/zaptest"
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
