package openapiv3_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"

	"goa.design/goa/v3/codegen/testutil"
	httpgen "goa.design/goa/v3/http/codegen"
	"goa.design/goa/v3/http/codegen/openapi"
	openapiv3 "goa.design/goa/v3/http/codegen/openapi/v3"
	"goa.design/goa/v3/http/codegen/testdata"
)

func TestFiles(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", "golden")
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		// TestSections
		{"file-service", testdata.FileServiceDSL},
		{"file-service-swagger", testdata.FileServiceSwaggerDSL},
		{"file-service-wildcard", testdata.FileServiceWildcardDSL},
		{"valid", testdata.SimpleDSL},
		{"multiple-services", testdata.MultipleServicesDSL},
		{"multiple-views", testdata.MultipleViewsDSL},
		{"explicit-view", testdata.ExplicitViewDSL},
		{"security", testdata.SecurityDSL},
		{"bearer-security", testdata.BearerSecurityDSL},
		{"server-host-with-variables", testdata.ServerHostWithVariablesDSL},
		{"with-spaces", testdata.WithSpacesDSL},
		{"with-map", testdata.WithMapDSL},
		{"with-any", testdata.WithAnyDSL},
		{"path-with-wildcards", testdata.PathWithWildcardDSL},
		{"path-with-multiple-wildcards", testdata.PathWithMultipleWildcardDSL},
		{"path-with-multiple-explicit-wildcards", testdata.PathWithMultipleExplicitWildcardDSL},
		{"headers", testdata.HeadersDSL},
		{"with-tags", testdata.WithTagsDSL},
		{"with-tags-swagger", testdata.WithTagsSwaggerDSL},
		{"typename", testdata.TypenameDSL},
		{"not-generate-server", testdata.NotGenerateServerDSL},
		{"not-generate-host", testdata.NotGenerateHostDSL},
		{"not-generate-attribute", testdata.NotGenerateAttributeDSL},
		{"json-prefix", testdata.JSONPrefixDSL},
		{"json-indent", testdata.JSONIndentDSL},
		{"json-prefix-indent", testdata.JSONPrefixIndentDSL},
		// TestEndpoints
		{"endpoint", testdata.ExtensionDSL},
		{"endpoint-swagger", testdata.ExtensionSwaggerDSL},
		{"type-extension", testdata.TypeExtensionDSL},
		// Alias types stay inline in 3.0 documents (named in 3.2 only).
		{"alias-type", testdata.AliasTypeDSL},
		{"skip-response-body-encode-decode", testdata.SkipResponseBodyEncodeDecodeDSL},
		// TestValidations
		{"string", testdata.StringValidationDSL},
		{"integer", testdata.IntValidationDSL},
		{"array", testdata.ArrayValidationDSL},
		// Error examples
		{"error-examples", testdata.ErrorExamplesDSL},
		// Streaming endpoints: OpenAPI 3.2 constructs must not leak into
		// 3.0 documents.
		{"sse-string", testdata.SSEStringDSL},
		{"sse-all-fields", testdata.SSEAllFieldsDSL},
		{"sse-mixed-results", testdata.MixedResultsDSL},
		{"websocket", testdata.StreamingResultDSL},
		// OpenAPI 3.2 only meta: must not leak into 3.0 documents.
		{"v3.2-meta", testdata.OpenAPIV32MetaDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset global variables
			openapi.Definitions = make(map[string]*openapi.Schema)
			root := httpgen.RunHTTPDSL(t, c.DSL)
			oFiles := openapiv3.Files(root, openapi.Version30, openapi.DefaultPath30)
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
					validateSwagger(t, buf.Bytes())

					golden := filepath.Join(goldenPath, fmt.Sprintf("%s_%s.golden", strings.TrimSuffix(c.Name, "-swagger"), tname))
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

func TestFilesV32(t *testing.T) {
	var (
		goldenPath = filepath.Join("testdata", "golden", "v3.2")
	)
	cases := []struct {
		Name string
		DSL  func()
	}{
		{"valid", testdata.SimpleDSL},
		{"v3.2-meta", testdata.OpenAPIV32MetaDSL},
		{"with-tags", testdata.WithTagsDSL},
		{"server-host-with-variables", testdata.ServerHostWithVariablesDSL},
		{"sse-string", testdata.SSEStringDSL},
		{"sse-object", testdata.SSEObjectDSL},
		{"sse-data-field", testdata.SSEDataFieldDSL},
		{"sse-request-id", testdata.SSERequestIDDSL},
		{"sse-all-fields", testdata.SSEAllFieldsDSL},
		{"sse-mixed-results", testdata.MixedResultsDSL},
		{"websocket", testdata.StreamingResultDSL},
		{"alias-type", testdata.AliasTypeDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset global variables
			openapi.Definitions = make(map[string]*openapi.Schema)
			root := httpgen.RunHTTPDSL(t, c.DSL)
			oFiles := openapiv3.Files(root, openapi.Version32, openapi.DefaultPath32)
			wantPaths := []string{
				filepath.Join("gen", "http", "openapi3.2.json"),
				filepath.Join("gen", "http", "openapi3.2.yaml"),
			}
			for i, o := range oFiles {
				tname := fmt.Sprintf("file%d", i)
				s := o.SectionTemplates
				t.Run(tname, func(t *testing.T) {
					if o.Path != wantPaths[i] {
						t.Errorf("got path %q, expected %q", o.Path, wantPaths[i])
					}
					if len(s) != 1 {
						t.Fatalf("expected 1 section, got %d", len(s))
					}
					var buf bytes.Buffer
					tmpl := template.Must(template.New("openapi").Funcs(s[0].FuncMap).Parse(s[0].Source))
					if err := tmpl.Execute(&buf, s[0].Data); err != nil {
						t.Fatalf("failed to render template: %s", err)
					}
					// kin-openapi only validates 3.0.x documents so check the
					// document version and well-formedness structurally.
					golden := filepath.Join(goldenPath, fmt.Sprintf("%s_%s.golden", c.Name, tname))
					if filepath.Ext(o.Path) == ".json" {
						if !json.Valid(buf.Bytes()) {
							t.Errorf("invalid JSON:\n%s", buf.String())
						}
						if !strings.Contains(buf.String(), `"openapi":"3.2.0"`) {
							t.Errorf("missing or wrong OpenAPI version:\n%s", buf.String())
						}
						testutil.AssertJSON(t, golden, buf.Bytes())
					} else {
						if !strings.Contains(buf.String(), "openapi: 3.2.0") {
							t.Errorf("missing or wrong OpenAPI version:\n%s", buf.String())
						}
						testutil.AssertString(t, golden, buf.String())
					}
				})
			}
		})
	}
}

func validateSwagger(t *testing.T, b []byte) {
	swagger, err := openapi3.NewLoader().LoadFromData(b)
	if err == nil {
		err = swagger.Validate(context.Background())
	}
	if err != nil {
		t.Errorf("invalid spec: %s\nspec:\n%s", err.Error(), string(b))
	}
}
