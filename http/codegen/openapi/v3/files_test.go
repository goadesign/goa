// This file renders complete OpenAPI 3.0 and 3.2 documents from prepared HTTP
// designs and compares output produced with run-owned example state.
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
	"goa.design/goa/v3/expr"
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
		{"released-response-collection-names", testdata.ReleasedResponseCollectionNamesDSL},
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
		{"shared-error-description", testdata.SharedErrorDescriptionDSL},
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
			root := expr.RunDSL(t, c.DSL)
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
		{"shared-error-description", testdata.SharedErrorDescriptionDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			root := expr.RunDSL(t, c.DSL)
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

// TestAuthoredExampleFixturesIgnoreRandomizer verifies that golden designs
// which do not test generated examples describe every displayed value.
func TestAuthoredExampleFixturesIgnoreRandomizer(t *testing.T) {
	cases := []struct {
		Name    string
		DSL     func()
		Version openapi.Version
	}{
		{"alias-type", testdata.AliasTypeDSL, openapi.Version30},
		{"array", testdata.ArrayValidationDSL, openapi.Version30},
		{"headers", testdata.HeadersDSL, openapi.Version30},
		{"not-generate-host", testdata.NotGenerateHostDSL, openapi.Version30},
		{"not-generate-server", testdata.NotGenerateServerDSL, openapi.Version30},
		{"path-with-wildcards", testdata.PathWithWildcardDSL, openapi.Version30},
		{"path-with-multiple-wildcards", testdata.PathWithMultipleWildcardDSL, openapi.Version30},
		{"path-with-multiple-explicit-wildcards", testdata.PathWithMultipleExplicitWildcardDSL, openapi.Version30},
		{"sse-all-fields", testdata.SSEAllFieldsDSL, openapi.Version32},
		{"sse-data-field", testdata.SSEDataFieldDSL, openapi.Version32},
		{"sse-mixed-results", testdata.MixedResultsDSL, openapi.Version32},
		{"sse-object", testdata.SSEObjectDSL, openapi.Version32},
		{"sse-request-id", testdata.SSERequestIDDSL, openapi.Version32},
		{"sse-string", testdata.SSEStringDSL, openapi.Version32},
		{"type-extension", testdata.TypeExtensionDSL, openapi.Version30},
		{"websocket", testdata.StreamingResultDSL, openapi.Version32},
		{"with-any", testdata.WithAnyDSL, openapi.Version30},
		{"with-map", testdata.WithMapDSL, openapi.Version30},
		{"with-spaces", testdata.WithSpacesDSL, openapi.Version30},
		{"with-tags", testdata.WithTagsDSL, openapi.Version32},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			firstRoot := expr.RunDSL(t, c.DSL)
			first := openapiv3.NewWithValues(
				firstRoot,
				c.Version,
				expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("first")),
				openapi.Values{},
			)
			firstJSON, err := json.Marshal(first)
			if err != nil {
				t.Fatalf("failed to encode first document: %s", err)
			}

			secondRoot := expr.RunDSL(t, c.DSL)
			second := openapiv3.NewWithValues(
				secondRoot,
				c.Version,
				expr.NewExampleGenerator(expr.NewFakerRandomizerFactory("second")),
				openapi.Values{},
			)
			secondJSON, err := json.Marshal(second)
			if err != nil {
				t.Fatalf("failed to encode second document: %s", err)
			}

			if !bytes.Equal(firstJSON, secondJSON) {
				t.Error("OpenAPI examples changed with the randomizer")
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
