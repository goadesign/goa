// This file checks complete multipart generation, including the starter
// decoder signatures and the validation that runs before payload construction.
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

// TestGenerateHTTPMultipartValidationCompilesAndRuns verifies generated
// object, array, and map bodies together with the generated starter decoder.
func TestGenerateHTTPMultipartValidationCompilesAndRuns(t *testing.T) {
	root := codegen.RunDSL(t, multipartValidationDSL)
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)
	exampleFiles, err := assembleExampleFilesForTest(plan)
	require.NoError(t, err)
	files = append(files, exampleFiles...)
	files, err = mergeFilesByPath(files)
	require.NoError(t, err)

	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	writeMultipartValidationRuntimeTest(t, directory)
	runGeneratedTests(t, directory)
}

// multipartValidationDSL defines one object body with mapped request values
// and two composite bodies that exercise generated callback signatures.
func multipartValidationDSL() {
	part := d.Type("Part", func() {
		d.Attribute("code", d.String)
		d.Required("code")
	})
	objectPayload := d.Type("ObjectPayload", func() {
		d.Attribute("name", d.String)
		d.Attribute("part", part)
		d.Attribute("site", d.String)
		d.Attribute("count", d.Int)
		d.Attribute("token", d.String)
		d.Required("name", "part", "site", "count")
	})
	d.Service("upload", func() {
		d.Method("Object", func() {
			d.Payload(objectPayload)
			d.HTTP(func() {
				d.POST("/objects/{site}")
				d.Param("count")
				d.Header("token:X-Token")
				d.MultipartRequest()
			})
		})
		d.Method("Array", func() {
			d.Payload(d.ArrayOf(part))
			d.HTTP(func() {
				d.POST("/array")
				d.MultipartRequest()
			})
		})
		d.Method("Map", func() {
			d.Payload(d.MapOf(d.String, d.Int))
			d.HTTP(func() {
				d.POST("/map")
				d.MultipartRequest()
			})
		})
	})
}

// writeMultipartValidationRuntimeTest adds assertions against the generated
// request decoder so validation order and payload construction are exercised.
func writeMultipartValidationRuntimeTest(t *testing.T, directory string) {
	t.Helper()
	const source = `package multiparttest_test

import (
	"errors"
	"net/http"
	"testing"

	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
	genserver "generated.local/gen/http/upload/server"
)

type mux struct{}

func (mux) Handle(string, string, http.HandlerFunc) {}
func (mux) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (mux) Vars(*http.Request) map[string]string { return map[string]string{"site": "west"} }

func TestObjectValidationRunsBeforeConstruction(t *testing.T) {
	code := "ready"
	request, err := http.NewRequest(http.MethodPost, "/objects/west?count=2", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("X-Token", "secret")

	missingName := genserver.DecodeObjectRequest(mux{}, bodyDecoder(func(body *genserver.ObjectRequestBody) {
		body.Part = &genserver.PartRequestBody{Code: &code}
	}))
	_, err = missingName(request)
	assertMissingField(t, err, "name", "body")

	name := "report"
	missingCode := genserver.DecodeObjectRequest(mux{}, bodyDecoder(func(body *genserver.ObjectRequestBody) {
		body.Name = &name
		body.Part = &genserver.PartRequestBody{}
	}))
	_, err = missingCode(request)
	assertMissingField(t, err, "code", "body.part")

	valid := genserver.DecodeObjectRequest(mux{}, bodyDecoder(func(body *genserver.ObjectRequestBody) {
		body.Name = &name
		body.Part = &genserver.PartRequestBody{Code: &code}
	}))
	payload, err := valid(request)
	if err != nil {
		t.Fatalf("valid multipart body failed: %v", err)
	}
	if payload.Name != name || payload.Part.Code != code || payload.Site != "west" || payload.Count != 2 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.Token == nil || *payload.Token != "secret" {
		t.Fatalf("mapped header was not preserved: %#v", payload.Token)
	}
}

func TestArrayAndMapBodiesConstructPayloads(t *testing.T) {
	code := "ready"
	request, err := http.NewRequest(http.MethodPost, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	decodeArray := genserver.DecodeArrayRequest(mux{}, func(*http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(value any) error {
			body := value.(*[]*genserver.PartRequestBody)
			*body = []*genserver.PartRequestBody{{Code: &code}}
			return nil
		})
	})
	array, err := decodeArray(request)
	if err != nil || len(array) != 1 || array[0].Code != code {
		t.Fatalf("unexpected array payload: %#v, %v", array, err)
	}

	decodeMap := genserver.DecodeMapRequest(mux{}, func(*http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(value any) error {
			body := value.(*map[string]int)
			*body = map[string]int{"count": 2}
			return nil
		})
	})
	values, err := decodeMap(request)
	if err != nil || values["count"] != 2 {
		t.Fatalf("unexpected map payload: %#v, %v", values, err)
	}
}

func bodyDecoder(fill func(*genserver.ObjectRequestBody)) func(*http.Request) goahttp.Decoder {
	return func(*http.Request) goahttp.Decoder {
		return goahttp.EncodingFunc(func(value any) error {
			fill(value.(*genserver.ObjectRequestBody))
			return nil
		})
	}
}

func assertMissingField(t *testing.T, err error, field, location string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected missing field %q", field)
	}
	var serviceError *goa.ServiceError
	if !errors.As(err, &serviceError) {
		t.Fatalf("expected Goa service error, got %T: %v", err, err)
	}
	if serviceError.Name != goa.MissingField || serviceError.Field == nil || *serviceError.Field != field {
		t.Fatalf("unexpected missing field error: %#v", serviceError)
	}
	if serviceError.Message != "\""+field+"\" is missing from "+location {
		t.Fatalf("unexpected message: %q", serviceError.Message)
	}
}
`
	dir := filepath.Join(directory, "multiparttest")
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "multipart_validation_test.go"), []byte(source), 0o600))
}
