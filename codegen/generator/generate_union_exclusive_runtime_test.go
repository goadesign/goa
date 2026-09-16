// This file verifies that generated service and HTTP unions store only the
// branch selected by their public constructors, setters, and JSON decoder.
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

	writeGeneratedTest(t, filepath.Join(generated, "exclusive_union", "union_storage_test.go"), serviceUnionStorageTest)
	writeGeneratedTest(t, filepath.Join(generated, "http", "exclusive_union", "server", "union_storage_test.go"), httpUnionStorageTest)
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
	"encoding/json"
	"errors"
	"testing"

	goa "goa.design/goa/v3/pkg"
)

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
	"encoding/json"
	"errors"
	"testing"

	goa "goa.design/goa/v3/pkg"
)

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
