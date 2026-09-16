// This file renders ordinary HTTP clients and servers into a temporary module.
// The generated tests verify that an optional selected body stays absent without
// dropping query values, while present bodies still decode and validate normally.
package codegen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedOptionalSelectedBodyPreservesOtherRequestValues checks both
// sides of the ordinary HTTP transport with absent, null, empty, and valid bodies.
func TestGeneratedOptionalSelectedBodyPreservesOtherRequestValues(t *testing.T) {
	root := expr.RunDSL(t, optionalSelectedBodyDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	serviceData := httpPlans[0].services.Get("optional_body")
	require.NotNil(t, serviceData)
	for _, endpoint := range serviceData.Endpoints {
		require.False(t, endpoint.Payload.Request.PayloadInit.ServerArgs[0].Required, endpoint.Method.Name)
	}

	files, err := service.Files(servicePlan)
	require.NoError(t, err)
	files = append(files, httpPlans[0].ClientFiles()...)
	files = append(files, httpPlans[0].ClientCLIFiles()...)
	files = append(files, httpPlans[0].ServerFiles()...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].ServerTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	runOptionalSelectedBodyRuntimeTest(t, files)
}

// optionalSelectedBodyDSL selects optional object and primitive fields as the
// request body while a required query field remains outside the body.
func optionalSelectedBodyDSL() {
	empty := dsl.Type("Empty", func() {})
	mode := dsl.Type("Mode", dsl.String)
	values := dsl.Type("Values", dsl.ArrayOf(dsl.String))
	labels := dsl.Type("Labels", dsl.MapOf(dsl.String, dsl.String))
	details := dsl.Type("Details", func() {
		dsl.Attribute("name", dsl.String)
		dsl.Required("name")
	})
	dsl.Service("optional_body", func() {
		dsl.Method("object", func() {
			dsl.Payload(func() {
				dsl.Attribute("data", details)
				dsl.Attribute("path", dsl.String)
				dsl.Attribute("tag", dsl.String)
				dsl.Attribute("header", dsl.String)
				dsl.Attribute("session", dsl.String)
				dsl.Required("path", "tag", "header", "session")
			})
			dsl.HTTP(func() {
				dsl.POST("/object/{path}")
				dsl.Param("path")
				dsl.Param("tag")
				dsl.Header("header:X-Test")
				dsl.Cookie("session:SID")
				dsl.Body("data")
			})
		})
		dsl.Method("text", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", dsl.String, func() {
					dsl.Pattern("^[a-z]*$")
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/text")
				dsl.Param("tag")
				dsl.Body("value")
			})
		})
		dsl.Method("inline", func() {
			dsl.Payload(func() {
				dsl.Attribute("data", func() {
					dsl.Attribute("name", dsl.String)
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/inline")
				dsl.Param("tag")
				dsl.Body("data")
			})
		})
		dsl.Method("default_text", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", mode, func() {
					dsl.Default("safe")
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/default-text")
				dsl.Param("tag")
				dsl.Body("value")
			})
		})
		dsl.Method("default_array", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", values, func() {
					dsl.Default([]string{"safe"})
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/default-array")
				dsl.Param("tag")
				dsl.Body("value")
			})
		})
		dsl.Method("default_object", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", details, func() {
					dsl.Default(map[string]any{"name": "safe"})
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/default-object")
				dsl.Param("tag")
				dsl.Body("value")
			})
		})
		dsl.Method("default_map", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", labels, func() {
					dsl.Default(map[string]string{"mode": "safe"})
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/default-map")
				dsl.Param("tag")
				dsl.Body("value")
			})
		})
		dsl.Method("default_bytes", func() {
			dsl.Payload(func() {
				dsl.Attribute("value", dsl.Bytes, func() {
					dsl.Default("safe")
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/default-bytes")
				dsl.Param("tag")
				dsl.Body("value")
			})
		})
		dsl.Method("default_union", func() {
			dsl.Payload(func() {
				dsl.OneOf("value", func() {
					dsl.TypeName("DefaultBodyValue")
					dsl.Field(1, "name", dsl.String)
					dsl.Field(2, "inactive", empty)
					dsl.Default(map[string]any{"type": "inactive", "value": map[string]any{}})
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/default-union")
				dsl.Param("tag")
				dsl.Body("value")
			})
		})
		for _, method := range []struct {
			name  string
			type_ any
		}{
			{name: "array", type_: values},
			{name: "map", type_: labels},
		} {
			dsl.Method(method.name, func() {
				dsl.Payload(func() {
					dsl.Attribute("value", method.type_)
					dsl.Attribute("tag", dsl.String)
					dsl.Required("tag")
				})
				dsl.HTTP(func() {
					dsl.POST("/" + method.name)
					dsl.Param("tag")
					dsl.Body("value")
				})
			})
		}
		dsl.Method("union", func() {
			dsl.Payload(func() {
				dsl.OneOf("value", func() {
					dsl.TypeName("OptionalBodyValue")
					dsl.Field(1, "name", dsl.String)
					dsl.Field(2, "inactive", empty)
				})
				dsl.Attribute("tag", dsl.String)
				dsl.Required("tag")
			})
			dsl.HTTP(func() {
				dsl.POST("/union")
				dsl.Param("tag")
				dsl.Body("value")
			})
		})
	})
}

// runOptionalSelectedBodyRuntimeTest writes the generated packages and runs
// tests beside the generated client and server code.
func runOptionalSelectedBodyRuntimeTest(t *testing.T, files []*goacodegen.File) {
	t.Helper()
	directory := t.TempDir()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	module := fmt.Sprintf(
		"module generated.local\n\ngo 1.25\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n",
		filepath.ToSlash(repository),
	)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	clientTest := filepath.Join(directory, "gen", "http", "optional_body", "client", "optional_body_test.go")
	serverTest := filepath.Join(directory, "gen", "http", "optional_body", "server", "optional_body_test.go")
	require.NoError(t, os.WriteFile(clientTest, []byte(generatedOptionalBodyClientTest), 0o600))
	require.NoError(t, os.WriteFile(serverTest, []byte(generatedOptionalBodyServerTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/http/optional_body/...")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated optional body tests:\n%s", output)
}

const generatedOptionalBodyClientTest = `package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/optional_body"
	goahttp "goa.design/goa/v3/http"
)

func TestOptionalSelectedBodiesEncodeOnlyWhenPresent(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/object/path-value", nil)
	err := EncodeObjectRequest(goahttp.RequestEncoder)(request, &service.ObjectPayload{
		Path: "path-value", Tag: "kept", Header: "header-value", Session: "session-value",
	})
	require.NoError(t, err)
	require.Equal(t, "/object/path-value", request.URL.Path)
	require.Equal(t, "kept", request.URL.Query().Get("tag"))
	require.Equal(t, "header-value", request.Header.Get("X-Test"))
	cookie, err := request.Cookie("SID")
	require.NoError(t, err)
	require.Equal(t, "session-value", cookie.Value)
	body, err := io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	request = httptest.NewRequest(http.MethodPost, "/object/path-value", nil)
	err = EncodeObjectRequest(goahttp.RequestEncoder)(request, &service.ObjectPayload{
		Data: &service.Details{Name: "value"}, Path: "path-value", Tag: "kept",
		Header: "header-value", Session: "session-value",
	})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `{"name":"value"}` + "`" + `, string(body))

	request = httptest.NewRequest(http.MethodPost, "/text", nil)
	err = EncodeTextRequest(goahttp.RequestEncoder)(request, &service.TextPayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	empty := ""
	request = httptest.NewRequest(http.MethodPost, "/text", nil)
	err = EncodeTextRequest(goahttp.RequestEncoder)(request, &service.TextPayload{Value: &empty, Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `""` + "`" + `, string(body))

	request = httptest.NewRequest(http.MethodPost, "/inline", nil)
	err = EncodeInlineRequest(goahttp.RequestEncoder)(request, &service.InlinePayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	request = httptest.NewRequest(http.MethodPost, "/array", nil)
	err = EncodeArrayRequest(goahttp.RequestEncoder)(request, &service.ArrayPayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	request = httptest.NewRequest(http.MethodPost, "/array", nil)
	err = EncodeArrayRequest(goahttp.RequestEncoder)(request, &service.ArrayPayload{
		Value: service.Values{},
		Tag: "kept",
	})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `[]` + "`" + `, string(body))

	request = httptest.NewRequest(http.MethodPost, "/map", nil)
	err = EncodeMapRequest(goahttp.RequestEncoder)(request, &service.MapPayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	request = httptest.NewRequest(http.MethodPost, "/map", nil)
	err = EncodeMapRequest(goahttp.RequestEncoder)(request, &service.MapPayload{
		Value: service.Labels{},
		Tag: "kept",
	})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `{}` + "`" + `, string(body))

	union := &service.UnionPayload{Tag: "kept"}
	union.Value.SetInactive(&service.Empty{})
	request = httptest.NewRequest(http.MethodPost, "/union", nil)
	err = EncodeUnionRequest(goahttp.RequestEncoder)(request, union)
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `{"type":"inactive","value":{}}` + "`" + `, string(body))

	request = httptest.NewRequest(http.MethodPost, "/default-text", nil)
	err = EncodeDefaultTextRequest(goahttp.RequestEncoder)(request, &service.DefaultTextPayload{
		Value: service.Mode(""),
		Tag: "kept",
	})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `""` + "`" + `, string(body))

	request = httptest.NewRequest(http.MethodPost, "/default-array", nil)
	err = EncodeDefaultArrayRequest(goahttp.RequestEncoder)(request, &service.DefaultArrayPayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	request = httptest.NewRequest(http.MethodPost, "/default-array", nil)
	err = EncodeDefaultArrayRequest(goahttp.RequestEncoder)(request, &service.DefaultArrayPayload{
		Value: service.Values{},
		Tag: "kept",
	})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.JSONEq(t, ` + "`" + `[]` + "`" + `, string(body))

	request = httptest.NewRequest(http.MethodPost, "/default-object", nil)
	err = EncodeDefaultObjectRequest(goahttp.RequestEncoder)(request, &service.DefaultObjectPayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	request = httptest.NewRequest(http.MethodPost, "/default-map", nil)
	err = EncodeDefaultMapRequest(goahttp.RequestEncoder)(request, &service.DefaultMapPayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	request = httptest.NewRequest(http.MethodPost, "/default-bytes", nil)
	err = EncodeDefaultBytesRequest(goahttp.RequestEncoder)(request, &service.DefaultBytesPayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	request = httptest.NewRequest(http.MethodPost, "/default-union", nil)
	err = EncodeDefaultUnionRequest(goahttp.RequestEncoder)(request, &service.DefaultUnionPayload{Tag: "kept"})
	require.NoError(t, err)
	body, err = io.ReadAll(request.Body)
	require.NoError(t, err)
	require.Empty(t, body)
}

func TestOptionalSelectedBodyCLIBuildersPreserveFlagPresence(t *testing.T) {
	tag := "kept"

	text, err := BuildTextPayload(nil, &tag)
	require.NoError(t, err)
	require.Nil(t, text.Value)
	require.Equal(t, "kept", text.Tag)
	empty := ""
	text, err = BuildTextPayload(&empty, &tag)
	require.NoError(t, err)
	require.NotNil(t, text.Value)
	require.Empty(t, *text.Value)

	path, header, session := "path-value", "header-value", "session-value"
	object, err := BuildObjectPayload(nil, &path, &tag, &header, &session)
	require.NoError(t, err)
	require.Nil(t, object.Data)
	require.Equal(t, "kept", object.Tag)

	inline, err := BuildInlinePayload(nil, &tag)
	require.NoError(t, err)
	require.Nil(t, inline.Data)
	require.Equal(t, "kept", inline.Tag)

	union, err := BuildUnionPayload(nil, &tag)
	require.NoError(t, err)
	require.Empty(t, union.Value.Kind())
	require.Equal(t, "kept", union.Tag)
}
`

const generatedOptionalBodyServerTest = `package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/optional_body"
	goahttp "goa.design/goa/v3/http"
)

type pathMux struct{}

func (pathMux) Handle(string, string, http.HandlerFunc) {}

func (pathMux) ServeHTTP(http.ResponseWriter, *http.Request) {}

func (pathMux) Vars(request *http.Request) map[string]string {
	return map[string]string{"path": request.PathValue("path")}
}

func TestOptionalSelectedBodiesDecodeWithoutDroppingQueryValues(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/object/path-value?tag=kept", nil)
	request.SetPathValue("path", "path-value")
	request.Header.Set("X-Test", "header-value")
	request.AddCookie(&http.Cookie{Name: "SID", Value: "session-value"})
	payload, err := DecodeObjectRequest(pathMux{}, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "path-value", payload.Path)
	require.Equal(t, "kept", payload.Tag)
	require.Equal(t, "header-value", payload.Header)
	require.Equal(t, "session-value", payload.Session)
	require.Nil(t, payload.Data)

	request = httptest.NewRequest(http.MethodPost, "/object/path-value?tag=kept", strings.NewReader(` + "`" + `{"name":"value"}` + "`" + `))
	request.SetPathValue("path", "path-value")
	request.Header.Set("X-Test", "header-value")
	request.AddCookie(&http.Cookie{Name: "SID", Value: "session-value"})
	payload, err = DecodeObjectRequest(pathMux{}, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", payload.Tag)
	require.Equal(t, "value", payload.Data.Name)

	request = httptest.NewRequest(http.MethodPost, "/object/path-value?tag=kept", strings.NewReader(` + "`" + `{}` + "`" + `))
	request.SetPathValue("path", "path-value")
	request.Header.Set("X-Test", "header-value")
	request.AddCookie(&http.Cookie{Name: "SID", Value: "session-value"})
	_, err = DecodeObjectRequest(pathMux{}, goahttp.RequestDecoder)(request)
	require.ErrorContains(t, err, ` + "`" + `"name" is missing` + "`" + `)

	request = httptest.NewRequest(http.MethodPost, "/object/path-value?tag=kept", strings.NewReader("null"))
	request.SetPathValue("path", "path-value")
	request.Header.Set("X-Test", "header-value")
	request.AddCookie(&http.Cookie{Name: "SID", Value: "session-value"})
	payload, err = DecodeObjectRequest(pathMux{}, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", payload.Tag)
	require.Nil(t, payload.Data)

	request = httptest.NewRequest(http.MethodPost, "/text?tag=kept", strings.NewReader("null"))
	text, err := DecodeTextRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", text.Tag)
	require.Nil(t, text.Value)

	request = httptest.NewRequest(http.MethodPost, "/text?tag=kept", strings.NewReader(` + "`" + `""` + "`" + `))
	text, err = DecodeTextRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", text.Tag)
	require.NotNil(t, text.Value)
	require.Empty(t, *text.Value)

	request = httptest.NewRequest(http.MethodPost, "/inline?tag=kept", nil)
	inline, err := DecodeInlineRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", inline.Tag)
	require.Nil(t, inline.Data)

	request = httptest.NewRequest(http.MethodPost, "/inline?tag=kept", strings.NewReader(` + "`" + `{"name":"value"}` + "`" + `))
	inline, err = DecodeInlineRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.NotNil(t, inline.Data.Name)
	require.Equal(t, "value", *inline.Data.Name)

	request = httptest.NewRequest(http.MethodPost, "/text?tag=kept", strings.NewReader(` + "`" + `"BAD"` + "`" + `))
	_, err = DecodeTextRequest(nil, goahttp.RequestDecoder)(request)
	require.ErrorContains(t, err, "must match")

	request = httptest.NewRequest(http.MethodPost, "/array?tag=kept", strings.NewReader("[]"))
	array, err := DecodeArrayRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", array.Tag)
	require.NotNil(t, array.Value)
	require.Empty(t, array.Value)

	request = httptest.NewRequest(http.MethodPost, "/map?tag=kept", strings.NewReader("{}"))
	values, err := DecodeMapRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", values.Tag)
	require.NotNil(t, values.Value)
	require.Empty(t, values.Value)

	request = httptest.NewRequest(http.MethodPost, "/union?tag=kept", strings.NewReader(` + "`" + `{"type":"inactive","value":{}}` + "`" + `))
	union, err := DecodeUnionRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", union.Tag)
	inactive, ok := union.Value.AsInactive()
	require.True(t, ok)
	require.NotNil(t, inactive)

	request = httptest.NewRequest(http.MethodPost, "/array?tag=kept", nil)
	array, err = DecodeArrayRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", array.Tag)
	require.Nil(t, array.Value)

	request = httptest.NewRequest(http.MethodPost, "/map?tag=kept", strings.NewReader("null"))
	values, err = DecodeMapRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", values.Tag)
	require.Nil(t, values.Value)

	request = httptest.NewRequest(http.MethodPost, "/union?tag=kept", nil)
	union, err = DecodeUnionRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", union.Tag)
	require.Empty(t, union.Value.Kind())

	request = httptest.NewRequest(http.MethodPost, "/default-text?tag=kept", nil)
	defaultText, err := DecodeDefaultTextRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", defaultText.Tag)
	require.Equal(t, "safe", string(defaultText.Value))

	request = httptest.NewRequest(http.MethodPost, "/default-text?tag=kept", strings.NewReader(` + "`" + `""` + "`" + `))
	defaultText, err = DecodeDefaultTextRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Empty(t, defaultText.Value)

	request = httptest.NewRequest(http.MethodPost, "/default-array?tag=kept", strings.NewReader("null"))
	defaultArray, err := DecodeDefaultArrayRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", defaultArray.Tag)
	require.Equal(t, service.Values{"safe"}, defaultArray.Value)

	request = httptest.NewRequest(http.MethodPost, "/default-array?tag=kept", strings.NewReader("[]"))
	defaultArray, err = DecodeDefaultArrayRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.NotNil(t, defaultArray.Value)
	require.Empty(t, defaultArray.Value)

	request = httptest.NewRequest(http.MethodPost, "/default-object?tag=kept", nil)
	defaultObject, err := DecodeDefaultObjectRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", defaultObject.Tag)
	require.NotNil(t, defaultObject.Value)
	require.Equal(t, "safe", defaultObject.Value.Name)

	request = httptest.NewRequest(http.MethodPost, "/default-map?tag=kept", nil)
	defaultMap, err := DecodeDefaultMapRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", defaultMap.Tag)
	require.Equal(t, service.Labels{"mode": "safe"}, defaultMap.Value)

	request = httptest.NewRequest(http.MethodPost, "/default-bytes?tag=kept", nil)
	defaultBytes, err := DecodeDefaultBytesRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", defaultBytes.Tag)
	require.Equal(t, []byte("safe"), defaultBytes.Value)

	request = httptest.NewRequest(http.MethodPost, "/default-union?tag=kept", nil)
	defaultUnion, err := DecodeDefaultUnionRequest(nil, goahttp.RequestDecoder)(request)
	require.NoError(t, err)
	require.Equal(t, "kept", defaultUnion.Tag)
	inactive, ok = defaultUnion.Value.AsInactive()
	require.True(t, ok)
	require.NotNil(t, inactive)
}
`
