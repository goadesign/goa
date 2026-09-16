// This file renders complete Swagger 2.0 documents from prepared HTTP designs
// and compares the JSON and YAML output produced with run-owned example state.
package openapiv2_test

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv2 "goa.design/goa/v3/http/codegen/openapi/v2"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestSections(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", t.Name())
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"empty", testdata.EmptyDSL},
		{"file-service", testdata.FileServiceDSL},
		{"valid", testdata.SimpleDSL},
		{"multiple-services", testdata.MultipleServicesDSL},
		{"multiple-views", testdata.MultipleViewsDSL},
		{"explicit-view", testdata.ExplicitViewDSL},
		{"released-response-collection-names", testdata.ReleasedResponseCollectionNamesDSL},
		{"security", testdata.SecurityDSL},
		{"server-host-with-variables", testdata.ServerHostWithVariablesDSL},
		{"with-spaces", testdata.WithSpacesDSL},
		{"with-map", testdata.WithMapDSL},
		{"with-any", testdata.WithAnyDSL},
		{"path-with-wildcards", testdata.PathWithWildcardDSL},
		{"path-with-multiple-wildcards", testdata.PathWithMultipleWildcardDSL},
		{"path-with-multiple-explicit-wildcards", testdata.PathWithMultipleExplicitWildcardDSL},
		{"headers", testdata.HeadersDSL},
		{"typename", testdata.TypenameDSL},
		{"not-generate-server", testdata.NotGenerateServerDSL},
		{"not-generate-host", testdata.NotGenerateHostDSL},
		{"not-generate-attribute", testdata.NotGenerateAttributeDSL},
		{"json-prefix", testdata.JSONPrefixDSL},
		{"json-indent", testdata.JSONIndentDSL},
		{"json-prefix-indent", testdata.JSONPrefixIndentDSL},
		{"additional-properties-type", testdata.AdditionalPropertiesTypeDSL},
		{"additional-properties-payload-result", testdata.AdditionalPropertiesPayloadResultDSL},
		{"additional-properties-embedded-payload-result", testdata.AdditionalPropertiesPayloadResultDSL},
		{"error-examples", testdata.ErrorExamplesDSL},
		{"shared-error-description", testdata.SharedErrorDescriptionDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			oFiles, err := openapiv2.Files(root, openapi.DefaultPath20)
			if err != nil {
				t.Fatalf("OpenAPI failed with %s", err)
			}
			for i, o := range oFiles {
				tname := fmt.Sprintf("file%d", i)
				s := o.SectionTemplates
				t.Run(tname, func(t *testing.T) {
					if len(s) != 1 {
						t.Fatalf("expected 1 section, got %d", len(s))
					}
					if s[0].Source == "" {
						t.Fatalf("empty section template")
					}
					if s[0].Data == nil {
						t.Fatalf("nil data")
					}
					var buf bytes.Buffer
					tmpl := template.Must(template.New("openapi").Funcs(s[0].FuncMap).Parse(s[0].Source))
					if err := tmpl.Execute(&buf, s[0].Data); err != nil {
						t.Fatalf("failed to render template: %s", err)
					}
					if filepath.Ext(o.Path) == ".json" {
						if err := validateSwagger(buf.Bytes()); err != nil {
							t.Errorf("invalid swagger: %s", err)
						}
					}

					golden := filepath.Join(goldenPath, fmt.Sprintf("%s_%s.golden", c.Name, tname))
					if filepath.Ext(o.Path) == ".json" {
						testutil.AssertJSON(t, golden, buf.Bytes())
					} else {
						testutil.AssertString(t, golden, buf.String())
					}
				})
			}
		})
	}
}

func TestValidations(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", t.Name())
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"string", testdata.StringValidationDSL},
		{"integer", testdata.IntValidationDSL},
		{"array", testdata.ArrayValidationDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			oFiles, err := openapiv2.Files(root, openapi.DefaultPath20)
			require.NoError(t, err, "OpenAPI failed")
			require.NotEmpty(t, oFiles, "No swagger files")
			for i, o := range oFiles {
				tname := fmt.Sprintf("file%d", i)
				s := o.SectionTemplates
				t.Run(tname, func(t *testing.T) {
					require.Len(t, s, 1, "expected 1 section, got %d", len(s))
					require.NotEmpty(t, s[0].Source, "empty section template")
					require.NotNil(t, s[0].Data, "nil data")
					var buf bytes.Buffer
					tmpl := template.Must(template.New("openapi").Funcs(s[0].FuncMap).Parse(s[0].Source))
					require.NoError(t, tmpl.Execute(&buf, s[0].Data), "failed to render template")
					if filepath.Ext(o.Path) == ".json" {
						require.NoError(t, validateSwagger(buf.Bytes()), "invalid swagger")
					}

					golden := filepath.Join(goldenPath, fmt.Sprintf("%s_%s.golden", c.Name, tname))
					if filepath.Ext(o.Path) == ".json" {
						testutil.AssertJSON(t, golden, buf.Bytes())
					} else {
						testutil.AssertString(t, golden, buf.String())
					}
				})
			}
		})
	}
}

func TestExtensions(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", t.Name())
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"endpoint", testdata.ExtensionDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
			oFiles, err := openapiv2.Files(root, openapi.DefaultPath20)
			require.NoError(t, err, "OpenAPI failed")
			require.NotEmpty(t, oFiles, "No swagger files")
			for i, o := range oFiles {
				tname := fmt.Sprintf("file%d", i)
				s := o.SectionTemplates
				t.Run(tname, func(t *testing.T) {
					require.Len(t, s, 1, "expected 1 section, got %d", len(s))
					require.NotEmpty(t, s[0].Source, "empty section template")
					require.NotNil(t, s[0].Data, "nil data")
					var buf bytes.Buffer
					tmpl := template.Must(template.New("openapi").Funcs(s[0].FuncMap).Parse(s[0].Source))
					require.NoError(t, tmpl.Execute(&buf, s[0].Data), "failed to render template")
					if filepath.Ext(o.Path) == ".json" {
						require.NoError(t, validateSwagger(buf.Bytes()), "invalid swagger")
					}

					golden := filepath.Join(goldenPath, fmt.Sprintf("%s_%s.golden", c.Name, tname))
					if filepath.Ext(o.Path) == ".json" {
						testutil.AssertJSON(t, golden, buf.Bytes())
					} else {
						testutil.AssertString(t, golden, buf.String())
					}
				})
			}
		})
	}
}

func TestNamedPrimitiveParamsAndHeadersUseOpenAPIBaseTypes(t *testing.T) {
	root := expr.RunDSL(t, func() {
		var UUID = dsl.Type("UUID", dsl.String, func() {
			dsl.Format(dsl.FormatUUID)
		})
		var Time = dsl.Type("Time", dsl.String, func() {
			dsl.Format(dsl.FormatDateTime)
		})
		var UploadStatus = dsl.ResultType("application/vnd.upload-status", func() {
			dsl.Attributes(func() {
				dsl.Attribute("expires", Time, "RFC3339 expiration timestamp.")
				dsl.Attribute("offset", dsl.Int64, "Current upload offset in bytes.")
			})
		})

		dsl.API("test", func() {
			dsl.Server("test", func() {
				dsl.Host("localhost", func() {
					dsl.URI("http://localhost:80")
				})
			})
		})

		dsl.Service("repro", func() {
			dsl.Method("show", func() {
				dsl.Payload(func() {
					dsl.Attribute("ids", dsl.ArrayOf(UUID), "UUID filter values.")
				})
				dsl.Result(UploadStatus)
				dsl.HTTP(func() {
					dsl.GET("/repro")
					dsl.Param("ids")
					dsl.Response(dsl.StatusOK, func() {
						dsl.Header("expires:Upload-Expires")
						dsl.Header("offset:Upload-Offset")
						dsl.Body(dsl.Empty)
					})
				})
			})
		})
	})

	spec, err := openapiv2.NewV2(root, root.API.Servers[0].Hosts[0])
	require.NoError(t, err)

	path, ok := spec.Paths["/repro"]
	require.True(t, ok)

	get := path.(*openapiv2.Path).Get
	require.Len(t, get.Parameters, 1)
	require.Equal(t, "array", get.Parameters[0].Type)
	require.NotNil(t, get.Parameters[0].Items)
	require.Equal(t, "string", get.Parameters[0].Items.Type)
	require.Equal(t, "uuid", get.Parameters[0].Items.Format)

	headers := get.Responses["200"].Headers
	require.Equal(t, "string", headers["Upload-Expires"].Type)
	require.Equal(t, "date-time", headers["Upload-Expires"].Format)
	require.Equal(t, "integer", headers["Upload-Offset"].Type)
	require.Equal(t, "int64", headers["Upload-Offset"].Format)
}

// validateSwagger asserts that the given bytes contain a valid Swagger spec.
func validateSwagger(b []byte) error {
	doc := &openapi2.T{}
	if err := doc.UnmarshalJSON(b); err != nil {
		return err
	}
	if doc.Swagger == "" {
		return errors.New("nil swagger")
	}
	return nil
}
