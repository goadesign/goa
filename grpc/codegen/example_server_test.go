// This file verifies generated gRPC server examples.
package codegen

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/codegen"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/testutil"
)

func TestExampleServerFiles(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"no-server", ctestdata.NoServerDSL},
		{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL},
		{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := codegen.RunDSL(t, c.DSL)
			examples := createExamplePlan(root, "generated.local/gen")
			fs := examples.ServerFiles()
			require.Greater(t, len(fs), 0)
			require.Greater(t, len(fs[0].SectionTemplates), 0)
			header := sectionCode(t, fs[0].SectionTemplates[0])
			require.NotContains(t, header, `goagrpc "goa.design/goa/v3/grpc"`)
			require.NotContains(t, header, `"generated.local"`)
			var buf bytes.Buffer
			for _, s := range fs[0].SectionTemplates[1:] {
				require.NoError(t, s.Write(&buf))
			}
			code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
			if strings.Contains(code, "GetServiceInfo") {
				t.Errorf("generated server discovers methods at runtime:\n%s", code)
			}
			if !strings.Contains(code, "serving gRPC method") {
				t.Errorf("generated server does not log its planned methods:\n%s", code)
			}
			golden := filepath.Join("testdata", "server-"+c.Name+".golden")
			testutil.AssertGo(t, golden, code)
		})
	}
}
