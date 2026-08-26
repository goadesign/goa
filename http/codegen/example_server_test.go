// This file verifies generated HTTP server examples.
package codegen

import (
	"bytes"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	ctestdata "goa.design/goa/v3/codegen/example/testdata"
	"goa.design/goa/v3/codegen/testutil"
	dsl "goa.design/goa/v3/dsl"
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
				root := codegen.RunDSL(t, c.DSL)
				require.Len(t, root.Services, 3)
				examples := linkedHTTPExamplePlanForRoot(t, root)
				fs := examples.ServerFiles()
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

	t.Run("multipart code check", func(t *testing.T) {
		cases := []struct {
			Name string
			DSL  func()
		}{
			{"object", testdata.PayloadMultipartValidationDSL},
			{"array", testdata.PayloadMultipartArrayTypeDSL},
			{"map", testdata.PayloadMultipartMapTypeDSL},
		}
		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				root := codegen.RunDSL(t, c.DSL)
				examples := linkedHTTPExamplePlanForRoot(t, root)
				var multipartFile *codegen.File
				for _, file := range examples.ServerFiles() {
					if file.Path == "multipart.go" {
						multipartFile = file
						break
					}
				}
				require.NotNil(t, multipartFile)
				var buf bytes.Buffer
				for _, section := range multipartFile.SectionTemplates {
					require.NoError(t, section.Write(&buf))
				}
				code := codegen.FormatTestCode(t, buf.String())
				golden := filepath.Join("testdata", "golden", "server-multipart-"+c.Name+".golden")
				testutil.CompareOrUpdateGolden(t, code, golden)
			})
		}
	})

	t.Run("code check", func(t *testing.T) {
		cases := []struct {
			Name string
			DSL  func()
		}{
			{"no-server", ctestdata.NoServerDSL},
			{"server-hosting-service-with-file-server", ctestdata.ServerHostingServiceWithFileServerDSL},
			{"server-hosting-service-subset", ctestdata.ServerHostingServiceSubsetDSL},
			{"server-hosting-multiple-services", ctestdata.ServerHostingMultipleServicesDSL},
			{"colliding-service-names", collidingServiceNamesDSL},
			{"streaming", testdata.StreamingMultipleServicesDSL},
		}
		for _, c := range cases {
			t.Run(c.Name, func(t *testing.T) {
				root := codegen.RunDSL(t, c.DSL)
				examples := linkedHTTPExamplePlanForRoot(t, root)
				fs := examples.ServerFiles()
				require.Len(t, fs, 1)
				require.Greater(t, len(fs[0].SectionTemplates), 0)
				var buf bytes.Buffer
				for _, s := range fs[0].SectionTemplates[1:] {
					require.NoError(t, s.Write(&buf))
				}
				code := codegen.FormatTestCode(t, "package foo\n"+buf.String())
				golden := filepath.Join("testdata", "golden", "server-"+c.Name+".golden")
				testutil.CompareOrUpdateGolden(t, code, golden)
			})
		}
	})
}

func TestExampleServerUsesServicePathsForLocalNames(t *testing.T) {
	root := codegen.RunDSL(t, collidingServiceNamesDSL)
	examples := linkedHTTPExamplePlanForRoot(t, root)
	files := examples.ServerFiles()
	require.Len(t, files, 1)

	var output bytes.Buffer
	for _, section := range files[0].SectionTemplates {
		require.NoError(t, section.Write(&output))
	}
	first := examples.transport.services.Get("read_value").Service
	second := examples.transport.services.Get("read-value").Service
	firstBase := codegen.Goify(first.PathName, false)
	secondBase := codegen.Goify(second.PathName, false)
	require.NotEqual(t, firstBase, secondBase)
	require.Contains(t, output.String(), firstBase+"Endpoints")
	require.Contains(t, output.String(), secondBase+"Endpoints")
	require.Contains(t, output.String(), firstBase+"Server")
	require.Contains(t, output.String(), secondBase+"Server")
}

// TestExampleServerImportsOnlyConfiguredServices verifies that an example
// server does not reserve packages for services it does not expose or for a
// WebSocket configurer it does not write.
func TestExampleServerImportsOnlyConfiguredServices(t *testing.T) {
	root := codegen.RunDSL(t, ctestdata.ServerHostingServiceSubsetDSL)
	examples := linkedHTTPExamplePlanForRoot(t, root)
	files := examples.ServerFiles()
	require.Len(t, files, 1)

	imports := importPaths(files[0].SectionTemplates[0].Data.(map[string]any)["Imports"].([]*codegen.ImportSpec))
	ignored := examples.transport.services.Get("IgnoredService").Service
	require.NotContains(t, imports, path.Join(examples.transport.services.GenPkg(), ignored.PathName))
	require.NotContains(t, imports, "github.com/gorilla/websocket")
}

// collidingServiceNamesDSL defines two services whose names become the same Go
// name. Their generated package paths remain distinct.
func collidingServiceNamesDSL() {
	dsl.API("collision", func() {
		dsl.Server("collision", func() {
			dsl.Services("read_value", "read-value")
		})
	})
	dsl.Service("read_value", func() {
		dsl.Method("read", func() {
			dsl.HTTP(func() { dsl.GET("/underscore") })
		})
	})
	dsl.Service("read-value", func() {
		dsl.Method("read", func() {
			dsl.HTTP(func() { dsl.GET("/dash") })
		})
	})
}
