// This file verifies generated gRPC command-line client examples.
package codegen

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/grpc/codegen/testdata"
)

func TestExampleCLIFiles(t *testing.T) {
	cases := []struct {
		Name    string
		DSL     func()
		PkgPath string
	}{
		{"no-server", ctestdata.NoServerDSL, "generated.local/gen"},
		{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL, "generated.local/gen"},
		{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL, "generated.local/gen"},
		{"no-server-pkgpath", ctestdata.NoServerDSL, "my/pkg/path"},
		{"server-hosting-service-subset-pkgpath", ctestdata.ServerHostingServiceSubsetDSL, "my/pkg/path"},
		{"server-hosting-multiple-services-pkgpath", ctestdata.ServerHostingMultipleServicesDSL, "my/pkg/path"},
		{"interceptors", testdata.InterceptorsDSL, "generated.local/gen"},
		{"server-streaming", testdata.ServerStreamingRPCDSL, "generated.local/gen"},
		{"client-streaming", testdata.ClientStreamingRPCDSL, "generated.local/gen"},
		{"bidirectional-streaming", testdata.BidirectionalStreamingRPCDSL, "generated.local/gen"},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			examples := createExamplePlan(root, c.PkgPath)
			fs := examples.CLIFiles()
			require.Greater(t, len(fs), 0)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			header := sectionCode(t, fs[0].SectionTemplates[0])
			for _, absent := range []string{`"os"`, `"time"`, `goa "goa.design/goa/v3/pkg"`, `goagrpc "goa.design/goa/v3/grpc"`} {
				require.NotContains(t, header, absent)
			}
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, buf.String())
			golden := filepath.Join("testdata", "client-"+c.Name+".golden")
			testutil.AssertGo(t, golden, code)
		})
	}
}
