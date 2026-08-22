// This file verifies generated server and client interceptor data, including
// selected payload and result attribute references.
package service

import (
	"bytes"
	"flag"
	"go/format"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/service/testdata"
)

var updateGolden = flag.Bool("update-interceptors", false, "update golden files for interceptor tests")

func TestInterceptors(t *testing.T) {
	cases := []struct {
		Name              string
		DSL               func()
		expectedFileCount int
	}{
		{"no-interceptors", testdata.NoInterceptorsDSL, 0},
		{"single-api-server-interceptor", testdata.SingleAPIServerInterceptorDSL, 2},
		{"single-service-server-interceptor", testdata.SingleServiceServerInterceptorDSL, 2},
		{"single-method-server-interceptor", testdata.SingleMethodServerInterceptorDSL, 2},
		{"single-client-interceptor", testdata.SingleClientInterceptorDSL, 2},
		{"multiple-interceptors", testdata.MultipleInterceptorsExampleDSL, 3},
		{"interceptor-with-read-payload", testdata.InterceptorWithReadPayloadDSL, 3},
		{"interceptor-with-write-payload", testdata.InterceptorWithWritePayloadDSL, 3},
		{"interceptor-with-read-write-payload", testdata.InterceptorWithReadWritePayloadDSL, 3},
		{"interceptor-with-read-result", testdata.InterceptorWithReadResultDSL, 3},
		{"interceptor-with-write-result", testdata.InterceptorWithWriteResultDSL, 3},
		{"interceptor-with-read-write-result", testdata.InterceptorWithReadWriteResultDSL, 3},
		{"streaming-interceptors", testdata.StreamingInterceptorsDSL, 3},
		{"streaming-interceptors-with-read-payload-and-read-streaming-payload", testdata.StreamingInterceptorsWithReadPayloadAndReadStreamingPayloadDSL, 3},
		{"streaming-interceptors-with-read-streaming-result", testdata.StreamingInterceptorsWithReadStreamingResultDSL, 3},
		{"streaming-interceptors-with-read-payload", testdata.StreamingInterceptorsWithReadPayloadDSL, 2},
		{"streaming-interceptors-with-read-result", testdata.StreamingInterceptorsWithReadResultDSL, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := runDSL(t, c.DSL)
			plan := mustServicePlan(t, root)
			require.Len(t, root.Services, 1)

			fs := interceptorsFiles(plan, plan.facts.services[0])

			require.Len(t, fs, c.expectedFileCount)
			for _, f := range fs {
				buf := new(bytes.Buffer)
				for _, s := range f.SectionTemplates[1:] {
					require.NoError(t, s.Write(buf))
				}
				bs, err := format.Source(buf.Bytes())
				require.NoError(t, err, buf.String())
				code := strings.ReplaceAll(string(bs), "\r\n", "\n")

				golden := filepath.Join("testdata", "interceptors", c.Name+"_"+filepath.Base(f.Path)+".golden")
				compareOrUpdateGolden(t, code, golden)
			}
		})
	}
}

func TestInvalidInterceptors(t *testing.T) {
	cases := []struct {
		Name        string
		DSL         func()
		ErrContains string
	}{
		{
			Name:        "streaming-result-interceptor",
			DSL:         testdata.StreamingResultInterceptorDSL,
			ErrContains: "cannot be applied because the method result is streaming",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := runDSLWithError(t, c.DSL)
			require.Error(t, err)
			assert.Contains(t, err.Error(), c.ErrContains)
		})
	}
}

func compareOrUpdateGolden(t *testing.T, code, golden string) {
	t.Helper()
	if *updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0750))
		require.NoError(t, os.WriteFile(golden, []byte(code), 0640))
		return
	}
	data, err := os.ReadFile(golden)
	require.NoError(t, err)
	if runtime.GOOS == "windows" {
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	}
	assert.Equal(t, string(data), code)
}
