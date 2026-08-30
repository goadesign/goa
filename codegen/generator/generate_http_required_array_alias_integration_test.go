// This file verifies that HTTP validation keeps null array elements visible
// without changing primitive alias elements in service values.
package generator

import (
	"path/filepath"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

// TestGenerateHTTPRequiredPrimitiveAliasArray checks the generated service and
// HTTP packages for an array whose string alias elements cannot be null.
func TestGenerateHTTPRequiredPrimitiveAliasArray(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
	)
	codegen.RunDSL(t, func() {
		alias := dsl.Type("Alias", dsl.String, func() {
			dsl.Pattern("^[a-z]*$")
		})
		plainAlias := dsl.Type("PlainAlias", dsl.String)
		nested := dsl.Type("Nested", func() {
			dsl.Field(1, "values", dsl.ArrayOfRequired(alias))
			dsl.Required("values")
		})
		payload := dsl.Type("StorePayload", func() {
			dsl.Field(1, "names", dsl.ArrayOfRequired(dsl.String))
			dsl.Field(2, "values", dsl.ArrayOfRequired(alias))
			dsl.Field(3, "nested", nested)
			dsl.Required("names", "values", "nested")
		})
		searchPayload := dsl.Type("SearchPayload", func() {
			dsl.Attribute("values", dsl.ArrayOfRequired(alias))
			dsl.Required("values")
		})
		dsl.Service("Aliases", func() {
			dsl.Method("Store", func() {
				dsl.Payload(payload)
				dsl.HTTP(func() {
					dsl.POST("/aliases")
				})
				dsl.GRPC(func() {})
			})
			dsl.Method("Search", func() {
				dsl.Payload(searchPayload)
				dsl.HTTP(func() {
					dsl.GET("/aliases")
					dsl.Param("values")
				})
			})
			dsl.Method("ListNames", func() {
				dsl.Result(dsl.ArrayOfRequired(dsl.String))
				dsl.HTTP(func() {
					dsl.GET("/names")
				})
			})
			dsl.Method("ListAliases", func() {
				dsl.Result(dsl.ArrayOfRequired(plainAlias))
				dsl.HTTP(func() {
					dsl.GET("/alias-values")
				})
			})
		})
	})

	directory := filepath.Join(t.TempDir(), codegen.Gendir)
	writeGeneratedModule(t, directory, "generated.local/gen")
	_, err := generate(filepath.Dir(directory), "gen", false, registry)
	if err != nil {
		t.Fatalf("generate required primitive alias array: %v", err)
	}
	writeGeneratedContractTest(
		t,
		directory,
		filepath.Join("http", "aliases", "server"),
		requiredPrimitiveAliasArrayRuntimeTest,
	)
	writeGeneratedContractTest(
		t,
		directory,
		filepath.Join("http", "aliases", "client"),
		requiredPrimitiveAliasArrayResponseRuntimeTest,
	)
	runGeneratedTests(t, directory)
}

const requiredPrimitiveAliasArrayRuntimeTest = `package server

import (
	"encoding/json"
	"testing"

	aliases "generated.local/gen/aliases"
	goa "goa.design/goa/v3/pkg"
)

func TestRequiredPrimitiveArrayElements(t *testing.T) {
	var valid StoreRequestBody
	if err := json.Unmarshal([]byte("{\"names\":[\"\"],\"values\":[\"\"],\"nested\":{\"values\":[\"\"]}}"), &valid); err != nil {
		t.Fatalf("decode valid body: %v", err)
	}
	if err := ValidateStoreRequestBody(&valid); err != nil {
		t.Fatalf("validate empty strings: %v", err)
	}
	payload := NewStorePayload(&valid)
	if len(payload.Names) != 1 || payload.Names[0] != "" {
		t.Fatalf("converted names = %#v", payload.Names)
	}
	if len(payload.Values) != 1 || payload.Values[0] != aliases.Alias("") {
		t.Fatalf("converted aliases = %#v", payload.Values)
	}
	if payload.Nested == nil || len(payload.Nested.Values) != 1 || payload.Nested.Values[0] != aliases.Alias("") {
		t.Fatalf("converted nested aliases = %#v", payload.Nested)
	}

	assertNullElement := func(body string, context string) {
		t.Helper()
		var decoded StoreRequestBody
		if err := json.Unmarshal([]byte(body), &decoded); err != nil {
			t.Fatalf("decode null element: %v", err)
		}
		err := ValidateStoreRequestBody(&decoded)
		if err == nil {
			t.Fatal("null element passed validation")
		}
		want := goa.MissingFieldError(context, "[*]").Error()
		if err.Error() != want {
			t.Fatalf("validation error = %q, want %q", err, want)
		}
	}
	assertNullElement("{\"names\":[null],\"values\":[\"ok\"],\"nested\":{\"values\":[\"ok\"]}}", "body.names")
	assertNullElement("{\"names\":[\"ok\"],\"values\":[null],\"nested\":{\"values\":[\"ok\"]}}", "body.values")
	assertNullElement("{\"names\":[\"ok\"],\"values\":[\"ok\"],\"nested\":{\"values\":[null]}}", "body.nested.values")
}
`

const requiredPrimitiveAliasArrayResponseRuntimeTest = `package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	aliases "generated.local/gen/aliases"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

func TestRequiredPrimitiveArrayResponseElements(t *testing.T) {
	decodeNames := DecodeListNamesResponse(goahttp.ResponseDecoder, false)
	decoded, err := decodeNames(response("[\"\"]"))
	if err != nil {
		t.Fatalf("decode empty string: %v", err)
	}
	names := decoded.([]string)
	if len(names) != 1 || names[0] != "" {
		t.Fatalf("decoded names = %#v", names)
	}
	assertNullResponse(t, decodeNames, "body")

	decodeAliases := DecodeListAliasesResponse(goahttp.ResponseDecoder, false)
	decoded, err = decodeAliases(response("[\"\"]"))
	if err != nil {
		t.Fatalf("decode empty alias: %v", err)
	}
	aliasValues := decoded.([]aliases.PlainAlias)
	if len(aliasValues) != 1 || aliasValues[0] != aliases.PlainAlias("") {
		t.Fatalf("decoded aliases = %#v", aliasValues)
	}
	assertNullResponse(t, decodeAliases, "body")
}

func assertNullResponse(t *testing.T, decode func(*http.Response) (any, error), context string) {
	t.Helper()
	_, err := decode(response("[null]"))
	if err == nil {
		t.Fatal("null element passed validation")
	}
	want := goa.MissingFieldError(context, "[*]").Error()
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("validation error = %q, want it to contain %q", err, want)
	}
}

func response(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    (&http.Request{}).WithContext(context.Background()),
	}
}
`
