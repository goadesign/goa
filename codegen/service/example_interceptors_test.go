package service

import (
	"bytes"
	"go/format"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/service/testdata"
	"goa.design/goa/v3/expr"
)

func TestExampleInterceptorsFiles(t *testing.T) {
	cases := []struct {
		Name          string
		DSL           func()
		ExpectedFiles []string
	}{
		{
			Name: "no-interceptors",
			DSL:  testdata.NoInterceptorExampleDSL,
		},
		{
			Name: "server-interceptor",
			DSL:  testdata.ServerInterceptorExampleDSL,
			ExpectedFiles: []string{
				filepath.Join("interceptors", "server_interceptor_service_server.go"),
			},
		},
		{
			Name: "client-interceptor",
			DSL:  testdata.ClientInterceptorExampleDSL,
			ExpectedFiles: []string{
				filepath.Join("interceptors", "client_interceptor_service_client.go"),
			},
		},
		{
			Name: "server-interceptor-by-name",
			DSL:  testdata.ServerInterceptorByNameExampleDSL,
			ExpectedFiles: []string{
				filepath.Join("interceptors", "server_interceptor_by_name_service_server.go"),
			},
		},
		{
			Name: "multiple-interceptors",
			DSL:  testdata.MultipleInterceptorsExampleDSL,
			ExpectedFiles: []string{
				filepath.Join("interceptors", "multiple_interceptors_service_server.go"),
				filepath.Join("interceptors", "multiple_interceptors_service_client.go"),
			},
		},
		{
			Name: "multiple-services",
			DSL:  testdata.MultipleServicesInterceptorsExampleDSL,
			ExpectedFiles: []string{
				filepath.Join("interceptors", "multiple_services_interceptors_service_server.go"),
				filepath.Join("interceptors", "multiple_services_interceptors_service_client.go"),
				filepath.Join("interceptors", "multiple_services_interceptors_service2_server.go"),
				filepath.Join("interceptors", "multiple_services_interceptors_service2_client.go"),
			},
		},
		{
			Name: "api-interceptors",
			DSL:  testdata.APIInterceptorExampleDSL,
			ExpectedFiles: []string{
				filepath.Join("interceptors", "api_interceptor_service_server.go"),
				filepath.Join("interceptors", "api_interceptor_service_client.go"),
			},
		},
		{
			Name: "chained-interceptors",
			DSL:  testdata.ChainedInterceptorExampleDSL,
			ExpectedFiles: []string{
				filepath.Join("interceptors", "chained_interceptor_service_server.go"),
				filepath.Join("interceptors", "chained_interceptor_service_client.go"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Run DSL
			root := expr.RunDSL(t, c.DSL)
			require.NotNil(t, root)

			// Generate files
			fs := ExampleInterceptorsFiles("", root)
			require.Len(t, fs, len(c.ExpectedFiles))

			// Verify file paths
			paths := make([]string, len(fs))
			for i, f := range fs {
				paths[i] = f.Path
			}
			assert.ElementsMatch(t, c.ExpectedFiles, paths)

			// Verify file content
			for _, f := range fs {
				buf := new(bytes.Buffer)
				for _, s := range f.SectionTemplates {
					require.NoError(t, s.Write(buf))
				}
				bs, err := format.Source(buf.Bytes())
				require.NoError(t, err, buf.String())
				code := string(bs)

				// Use the base name of the generated file without extension for the golden file
				baseName := filepath.Base(f.Path)
				ext := filepath.Ext(baseName)
				goldenName := baseName[:len(baseName)-len(ext)] + ".golden"
				golden := filepath.Join("testdata", "example_interceptors", goldenName)
				compareOrUpdateGolden(t, code, golden)
			}
		})
	}
}
