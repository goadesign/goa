// This file verifies that generated service and HTTP unions store only the
// branch selected by their public constructors, setters, and JSON decoder.
// The same selected value reaches Go templates, including zero-valued branches.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
)

func TestGeneratedUnionsStoreOnlyTheSelectedBranch(t *testing.T) {
	registry := testRegistry(
		"gen",
		testGenerator(planServiceData, testServiceFiles),
		testGenerator(planTransportData, testTransportFiles),
	)

	codegen.RunDSL(t, exclusiveUnionDSL)
	directory := t.TempDir()
	generated := filepath.Join(directory, codegen.Gendir)
	writeGeneratedModule(t, generated, "generated.local/gen")
	if _, err := generate(directory, "gen", false, registry); err != nil {
		t.Fatalf("generate exclusive unions: %v", err)
	}

	writeGeneratedTest(t, filepath.Join(generated, "exclusive_union", "union_storage_test.go"), serviceUnionStorageTest+unionValueContractTest)
	writeGeneratedTest(t, filepath.Join(generated, "http", "exclusive_union", "server", "union_storage_test.go"), httpUnionStorageTest+unionValueContractTest)
	runGeneratedTests(t, generated)
}

// exclusiveUnionDSL uses the same OneOf in a service payload and an HTTP
// request so both generated union implementations receive identical tests.
func exclusiveUnionDSL() {
	dsl.API("exclusive union", func() {})
	inactive := dsl.Type("Inactive", func() {})
	kindValue := dsl.Type("KindValue", func() {})
	selection := dsl.Type("Selection", func() {
		dsl.OneOf("choice", func() {
			dsl.TypeName("ExclusiveChoice")
			dsl.Attribute("text", dsl.String)
			dsl.Attribute("count", dsl.Int)
			dsl.Attribute("enabled", dsl.Boolean)
			dsl.Attribute("kind", kindValue)
			dsl.Attribute("inactive", inactive)
		})
		dsl.Required("choice")
	})
	dsl.Service("ExclusiveUnion", func() {
		dsl.Method("Select", func() {
			dsl.Payload(selection)
			dsl.Result(dsl.String)
			dsl.HTTP(func() {
				dsl.POST("/selection")
				dsl.Response(dsl.StatusOK)
			})
		})
	})
}

// writeGeneratedTest adds a runtime contract test beside generated source.
func writeGeneratedTest(t *testing.T, path, source string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write generated contract test %s: %v", path, err)
	}
}

const serviceUnionStorageTest = `package exclusiveunion

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"text/template"

	goa "goa.design/goa/v3/pkg"
)

func TestUnionValue(t *testing.T) {
	checkUnionValues(t, []unionValueCase{
		{"empty text", NewExclusiveChoiceText(""), ExclusiveChoiceBranchText(""), "", ""},
		{"zero", NewExclusiveChoiceCount(0), ExclusiveChoiceBranchCount(0), "0", ""},
		{"false", NewExclusiveChoiceEnabled(false), ExclusiveChoiceBranchEnabled(false), "false", ""},
		{"object", NewExclusiveChoiceInactive(&Inactive{}), &Inactive{}, "{}", ""},
		{"unselected", ExclusiveChoice{}, nil, "", goa.InvalidEnumValue},
		{"nil object", NewExclusiveChoiceInactive(nil), nil, "", goa.MissingField},
	})
}

func TestUnionStorage(t *testing.T) {
	var selected ExclusiveChoice
	selected.SetText("old")
	selected.SetCount(7)
	if selected.text != "" {
		t.Errorf("SetCount retained text %q", selected.text)
	}
	if selected.count != 7 {
		t.Errorf("SetCount stored %d, want 7", selected.count)
	}
	selected.SetKind(&KindValue{})
	if selected.kind2 == nil {
		t.Error("SetKind did not store its selected branch")
	}
	if selected.count != 0 {
		t.Errorf("SetKind retained count %d", selected.count)
	}

	if err := json.Unmarshal([]byte(` + "`" + `{"type":"text","value":"old"}` + "`" + `), &selected); err != nil {
		t.Errorf("decode text: %v", err)
	}
	if err := json.Unmarshal([]byte(` + "`" + `{"type":"count","value":9}` + "`" + `), &selected); err != nil {
		t.Errorf("decode count: %v", err)
	}
	if selected.text != "" {
		t.Errorf("decoding count retained text %q", selected.text)
	}
	if selected.count != 9 {
		t.Errorf("decoded count %d, want 9", selected.count)
	}
	if err := json.Unmarshal([]byte(` + "`" + `{"type":"text","value":false}` + "`" + `), &selected); err == nil {
		t.Error("decoding an invalid text value succeeded")
	}
	if selected.count != 9 {
		t.Errorf("failed decode changed count to %d", selected.count)
	}
	encoded, err := json.Marshal(selected)
	if err != nil {
		t.Errorf("encode selected count: %v", err)
	} else if string(encoded) != ` + "`" + `{"type":"count","value":9}` + "`" + ` {
		t.Errorf("encoded union %s", encoded)
	}

	var missing ExclusiveChoice
	assertServiceError(t, missing.Validate(), goa.InvalidEnumValue, "type")
	missing.SetInactive(nil)
	assertServiceError(t, missing.Validate(), goa.MissingField, "value")
	missing.SetInactive(&Inactive{})
	if err := missing.Validate(); err != nil {
		t.Errorf("Validate rejected a selected empty-message branch: %v", err)
	}
}

// assertServiceError checks the exact Goa error returned for an invalid union.
func assertServiceError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	if !errors.As(err, &serviceError) {
		t.Errorf("expected Goa service error, got %T: %v", err, err)
		return
	}
	if serviceError.Name != name {
		t.Errorf("error name %q, want %q", serviceError.Name, name)
	}
	if serviceError.Field == nil || *serviceError.Field != field {
		t.Errorf("error field %#v, want %q", serviceError.Field, field)
	}
}
`

const httpUnionStorageTest = `package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"text/template"

	goa "goa.design/goa/v3/pkg"
)

func TestUnionValue(t *testing.T) {
	checkUnionValues(t, []unionValueCase{
		{"empty text", NewExclusiveChoiceRequestBodyText(""), "", "", ""},
		{"zero", NewExclusiveChoiceRequestBodyCount(0), 0, "0", ""},
		{"false", NewExclusiveChoiceRequestBodyEnabled(false), false, "false", ""},
		{"object", NewExclusiveChoiceRequestBodyInactive(&InactiveRequestBody{}), &InactiveRequestBody{}, "{}", ""},
		{"unselected", ExclusiveChoiceRequestBody{}, nil, "", goa.InvalidEnumValue},
		{"nil object", NewExclusiveChoiceRequestBodyInactive(nil), nil, "", goa.MissingField},
	})
}

func TestUnionStorage(t *testing.T) {
	var selected ExclusiveChoiceRequestBody
	selected.SetText("old")
	selected.SetCount(7)
	if selected.text != "" {
		t.Errorf("SetCount retained text %q", selected.text)
	}
	if selected.count != 7 {
		t.Errorf("SetCount stored %d, want 7", selected.count)
	}
	selected.SetKind(&KindValueRequestBody{})
	if selected.kind2 == nil {
		t.Error("SetKind did not store its selected branch")
	}
	if selected.count != 0 {
		t.Errorf("SetKind retained count %d", selected.count)
	}

	if err := json.Unmarshal([]byte(` + "`" + `{"type":"text","value":"old"}` + "`" + `), &selected); err != nil {
		t.Errorf("decode text: %v", err)
	}
	if err := json.Unmarshal([]byte(` + "`" + `{"type":"count","value":9}` + "`" + `), &selected); err != nil {
		t.Errorf("decode count: %v", err)
	}
	if selected.text != "" {
		t.Errorf("decoding count retained text %q", selected.text)
	}
	if selected.count != 9 {
		t.Errorf("decoded count %d, want 9", selected.count)
	}
	if err := json.Unmarshal([]byte(` + "`" + `{"type":"text","value":false}` + "`" + `), &selected); err == nil {
		t.Error("decoding an invalid text value succeeded")
	}
	if selected.count != 9 {
		t.Errorf("failed decode changed count to %d", selected.count)
	}
	encoded, err := json.Marshal(selected)
	if err != nil {
		t.Errorf("encode selected count: %v", err)
	} else if string(encoded) != ` + "`" + `{"type":"count","value":9}` + "`" + ` {
		t.Errorf("encoded union %s", encoded)
	}

	var missing ExclusiveChoiceRequestBody
	assertServiceError(t, missing.Validate(), goa.InvalidEnumValue, "type")
	missing.SetInactive(nil)
	assertServiceError(t, missing.Validate(), goa.MissingField, "value")
	missing.SetInactive(&InactiveRequestBody{})
	if err := missing.Validate(); err != nil {
		t.Errorf("Validate rejected a selected empty-message branch: %v", err)
	}
}

// assertServiceError checks the exact Goa error returned for an invalid union.
func assertServiceError(t *testing.T, err error, name, field string) {
	t.Helper()
	var serviceError *goa.ServiceError
	if !errors.As(err, &serviceError) {
		t.Errorf("expected Goa service error, got %T: %v", err, err)
		return
	}
	if serviceError.Name != name {
		t.Errorf("error name %q, want %q", serviceError.Name, name)
	}
	if serviceError.Field == nil || *serviceError.Field != field {
		t.Errorf("error field %#v, want %q", serviceError.Field, field)
	}
}
`

const unionValueContractTest = `
type (
	unionValueCase struct {
		name string
		selected interface { Value() (any, error) }
		want any
		rendered string
		errorName string
	}
)

// checkUnionValues reads each selected branch through Go and a Go template.
// Invalid selections must also fail JSON encoding with the same Goa error.
func checkUnionValues(t *testing.T, cases []unionValueCase) {
	t.Helper()
	renderer, err := template.New("value").Parse("{{.Value}}")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			value, valueError := test.selected.Value()
			var output bytes.Buffer
			templateError := renderer.Execute(&output, test.selected)
			_, jsonError := json.Marshal(test.selected)
			if test.errorName != "" {
				field := "type"
				if test.errorName == goa.MissingField {
					field = "value"
				}
				assertServiceError(t, valueError, test.errorName, field)
				assertServiceError(t, templateError, test.errorName, field)
				assertServiceError(t, jsonError, test.errorName, field)
				return
			}
			if valueError != nil || templateError != nil || jsonError != nil {
				t.Errorf("valid selection failed: Value=%v, template=%v, JSON=%v", valueError, templateError, jsonError)
			}
			if !reflect.DeepEqual(test.want, value) {
				t.Errorf("Value returned %#v, want %#v", value, test.want)
			}
			if output.String() != test.rendered {
				t.Errorf("template returned %q, want %q", output.String(), test.rendered)
			}
		})
	}
}
`
