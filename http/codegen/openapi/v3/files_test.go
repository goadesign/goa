package openapiv3_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
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
		{"server-host-with-variables", testdata.ServerHostWithVariablesDSL},
		{"with-spaces", testdata.WithSpacesDSL},
		{"with-map", testdata.WithMapDSL},
		{"with-any", testdata.WithAnyDSL},
		{"constructor-union", testdata.ConstructorUnionHTTPDSL},
		{"constructor-union-custom-keys", testdata.ConstructorUnionCustomKeysHTTPDSL},
		{"constructor-union-user-example-second-branch", testdata.ConstructorUnionUserExampleSecondBranchHTTPDSL},
		{"nested-constructor-union", testdata.NestedConstructorUnionHTTPDSL},
		{"nested-top-level-constructor-union", testdata.NestedTopLevelConstructorUnionHTTPDSL},
		{"nested-top-level-constructor-union-custom-keys", testdata.NestedTopLevelConstructorUnionCustomKeysHTTPDSL},
		{"recursive-constructor-union", testdata.RecursiveConstructorUnionHTTPDSL},
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
		{"skip-response-body-encode-decode", testdata.SkipResponseBodyEncodeDecodeDSL},
		// TestValidations
		{"string", testdata.StringValidationDSL},
		{"integer", testdata.IntValidationDSL},
		{"array", testdata.ArrayValidationDSL},
		// Error examples
		{"error-examples", testdata.ErrorExamplesDSL},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			// Reset global variables
			openapi.Definitions = make(map[string]*openapi.Schema)
			root := httpgen.RunHTTPDSL(t, c.DSL)
			oFiles, err := openapiv3.Files(root)
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

func TestConstructorUnionFilesStableAcrossBranchOrder(t *testing.T) {
	forward := renderOpenAPIFiles(t, testdata.ConstructorUnionHTTPDSL)
	reordered := renderOpenAPIFiles(t, testdata.ConstructorUnionHTTPReorderedDSL)
	if len(forward) != len(reordered) {
		t.Fatalf("expected same number of rendered files, got %d and %d", len(forward), len(reordered))
	}
	for i := range forward {
		if forward[i] != reordered[i] {
			t.Fatalf("expected identical OpenAPI output for reordered constructor union branches in file %d", i)
		}
	}
}

func TestConstructorUnionFilesStableAcrossUnrelatedDeclarationOrder(t *testing.T) {
	forward := renderOpenAPIFiles(t, testdata.ConstructorUnionUnrelatedDeclarationOrderHTTPDSL)
	reordered := renderOpenAPIFiles(t, testdata.ConstructorUnionUnrelatedDeclarationOrderReorderedHTTPDSL)
	if len(forward) != len(reordered) {
		t.Fatalf("expected same number of rendered files, got %d and %d", len(forward), len(reordered))
	}
	forwardSig := constructorUnionRenderedSignature(t, forward)
	reorderedSig := constructorUnionRenderedSignature(t, reordered)
	if !reflect.DeepEqual(forwardSig, reorderedSig) {
		t.Fatalf("expected stable constructor-union request/response schema signature across unrelated declaration order, got %#v and %#v", forwardSig, reorderedSig)
	}
}

func TestConstructorUnionFilesStableAcrossServiceTraversalOrder(t *testing.T) {
	forward := renderOpenAPIFiles(t, testdata.ConstructorUnionTraversalOrderHTTPDSL)
	reordered := renderOpenAPIFiles(t, testdata.ConstructorUnionTraversalOrderReorderedHTTPDSL)
	if len(forward) != len(reordered) {
		t.Fatalf("expected same number of rendered files, got %d and %d", len(forward), len(reordered))
	}
	forwardSig := traversalRenderedSignature(t, forward)
	reorderedSig := traversalRenderedSignature(t, reordered)
	if !reflect.DeepEqual(forwardSig, reorderedSig) {
		t.Fatalf("expected stable constructor-union component refs across service traversal order, got %#v and %#v", forwardSig, reorderedSig)
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

func renderOpenAPIFiles(t *testing.T, dslFn func()) []string {
	t.Helper()

	openapi.Definitions = make(map[string]*openapi.Schema)
	root := httpgen.RunHTTPDSL(t, dslFn)
	oFiles, err := openapiv3.Files(root)
	if err != nil {
		t.Fatalf("OpenAPI failed with %s", err)
	}

	rendered := make([]string, len(oFiles))
	for i, o := range oFiles {
		sections := o.SectionTemplates
		if len(sections) != 1 {
			t.Fatalf("expected 1 section, got %d", len(sections))
		}
		var buf bytes.Buffer
		tmpl := template.Must(template.New("openapi").Funcs(sections[0].FuncMap).Parse(sections[0].Source))
		if err := tmpl.Execute(&buf, sections[0].Data); err != nil {
			t.Fatalf("failed to render template: %s", err)
		}
		validateSwagger(t, buf.Bytes())
		rendered[i] = buf.String()
	}

	return rendered
}

func constructorUnionRenderedSignature(t *testing.T, rendered []string) map[string]any {
	t.Helper()

	for _, doc := range rendered {
		if !strings.HasPrefix(strings.TrimSpace(doc), "{") {
			continue
		}
		var spec map[string]any
		if err := json.Unmarshal([]byte(doc), &spec); err != nil {
			t.Fatalf("failed to parse rendered OpenAPI JSON: %s", err)
		}
		paths := nestedMap(spec, "paths")
		post := nestedMap(paths, "/", "post")
		reqSchema := nestedMap(post, "requestBody", "content", "application/json", "schema")
		respSchema := nestedMap(post, "responses", "200", "content", "application/json", "schema")
		components := nestedMap(spec, "components", "schemas")
		names := make([]string, 0)
		for name := range components {
			if strings.Contains(name, "JSONPayloadOrTextPayload") {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		return map[string]any{
			"requestMapping":  nestedMap(reqSchema, "discriminator", "mapping"),
			"requestOneOf":    reqSchema["oneOf"],
			"requestExample":  reqSchema["example"],
			"responseMapping": nestedMap(respSchema, "discriminator", "mapping"),
			"responseOneOf":   respSchema["oneOf"],
			"responseExample": respSchema["example"],
			"componentNames":  names,
		}
	}
	t.Fatal("expected rendered JSON OpenAPI document")
	return nil
}

func traversalRenderedSignature(t *testing.T, rendered []string) map[string]any {
	t.Helper()

	for _, doc := range rendered {
		if !strings.HasPrefix(strings.TrimSpace(doc), "{") {
			continue
		}
		var spec map[string]any
		if err := json.Unmarshal([]byte(doc), &spec); err != nil {
			t.Fatalf("failed to parse rendered OpenAPI JSON: %s", err)
		}
		components := nestedMap(spec, "components", "schemas")
		names := make([]string, 0, len(components))
		for name := range components {
			if strings.Contains(name, "Traversal") {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		signature := make(map[string]any, len(names))
		for _, name := range names {
			signature[name] = components[name]
		}
		return signature
	}
	t.Fatal("expected rendered JSON OpenAPI document")
	return nil
}

func nestedMap(m map[string]any, keys ...string) map[string]any {
	current := m
	for _, key := range keys {
		next, ok := current[key].(map[string]any)
		if !ok {
			return nil
		}
		current = next
	}
	return current
}
