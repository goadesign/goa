package codegen

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/example"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestExampleServerFiles(t *testing.T) {
	t.Run("package name check", func(t *testing.T) {
		cases := []struct {
			Name     string
			DSL      func()
			Expected string
		}{
			{
				Name:     "conflict with API name and service names including multipart",
				DSL:      ctestdata.ConflictWithAPINameAndServiceNamesIncludingMultipartDSL,
				Expected: "package alohaapi2",
			},
		}
		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				// reset global variable
				HTTPServices = make(ServicesData)
				service.Services = make(service.ServicesData)
				example.Servers = make(example.ServersData)
				codegen.RunDSL(t, c.DSL)
				require.Len(t, expr.Root.Services, 3)
				fs := ExampleServerFiles("", expr.Root)
				require.Len(t, fs, 2)
				for i, f := range fs {
					if i < len(fs)-1 {
						// Skip example http server.
						continue
					}
					require.Greater(t, len(f.SectionTemplates), 0)
					var b bytes.Buffer
					require.NoError(t, f.SectionTemplates[0].Write(&b))
					line, err := b.ReadBytes('\n')
					assert.NoError(t, err)
					got := string(bytes.TrimRight(line, "\n"))
					assert.Equal(t, c.Expected, got)
				}
			})
		}
	})

	t.Run("code check", func(t *testing.T) {
		cases := []struct {
			Name string
			DSL  func()
			Code string
		}{
			{"no-server", ctestdata.NoServerDSL, testdata.NoServerServerHandleCode},
			{"server-hosting-service-with-file-server", ctestdata.ServerHostingServiceWithFileServerDSL, testdata.ServerHostingServiceWithFileServerHandlerCode},
			{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL, testdata.ServerHostingServiceSubsetServerHandleCode},
			{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL, testdata.ServerHostingMultipleServicesServerHandleCode},
			{"streaming", testdata.StreamingMultipleServicesDSL, testdata.StreamingServerHandleCode},
		}
		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				// reset global variable
				HTTPServices = make(ServicesData)
				service.Services = make(service.ServicesData)
				example.Servers = make(example.ServersData)
				codegen.RunDSL(t, c.DSL)
				fs := ExampleServerFiles("", expr.Root)
				require.Len(t, fs, 1)
				require.Greater(t, len(fs[0].SectionTemplates), 0)
				var buf bytes.Buffer
				for _, s := range fs[0].SectionTemplates[1:] {
					require.NoError(t, s.Write(&buf))
				}
				code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
				assert.Equal(t, c.Code, code)
			})
		}
	})
}
