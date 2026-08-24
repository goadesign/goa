// This file renders an HTTP SSE service into a temporary module. The generated
// server and client tests check the exact text used for primitive fields and
// the JSON used for structured fields.
package codegen

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedSSEFieldWireFormat checks both sides of the generated SSE
// connection for primitive, declared primitive, object, and array fields.
func TestGeneratedSSEFieldWireFormat(t *testing.T) {
	root := expr.RunDSL(t, sseFieldWireDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())

	serviceFiles, err := service.Files(servicePlan)
	require.NoError(t, err)
	files := slices.Clone(serviceFiles)
	files = append(files, httpPlans[0].ServerFiles()...)
	files = append(files, httpPlans[0].ClientFiles()...)
	files = append(files, httpPlans[0].ServerTypeFiles()...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	runGeneratedSSEFieldWireTests(t, files)
}

// sseFieldWireDSL maps each representative field type to the SSE data line.
func sseFieldWireDSL() {
	eventText := dsl.Type("EventText", dsl.String)
	requiredText := dsl.Type("RequiredText", func() {
		dsl.Attribute("value", dsl.String)
		dsl.Required("value")
	})
	aliasText := dsl.Type("AliasText", func() {
		dsl.Attribute("value", eventText)
		dsl.Required("value")
	})
	optionalText := dsl.Type("OptionalText", func() {
		dsl.Attribute("value", dsl.String)
	})
	wireObject := dsl.Type("WireObject", func() {
		dsl.Attribute("label", dsl.String)
		dsl.Required("label")
	})
	structured := dsl.Type("Structured", func() {
		dsl.Attribute("object", wireObject)
		dsl.Attribute("values", dsl.ArrayOf(dsl.String))
		dsl.Required("object", "values")
	})

	dsl.Service("SSE Wire", func() {
		dsl.Method("required", func() {
			dsl.StreamingResult(requiredText)
			dsl.HTTP(func() {
				dsl.GET("/required")
				dsl.ServerSentEvents("value")
			})
		})
		dsl.Method("alias", func() {
			dsl.StreamingResult(aliasText)
			dsl.HTTP(func() {
				dsl.GET("/alias")
				dsl.ServerSentEvents("value")
			})
		})
		dsl.Method("optional", func() {
			dsl.StreamingResult(optionalText)
			dsl.HTTP(func() {
				dsl.GET("/optional")
				dsl.ServerSentEvents("value")
			})
		})
		dsl.Method("object", func() {
			dsl.StreamingResult(structured)
			dsl.HTTP(func() {
				dsl.GET("/object")
				dsl.ServerSentEvents("object")
			})
		})
		dsl.Method("array", func() {
			dsl.StreamingResult(structured)
			dsl.HTTP(func() {
				dsl.GET("/array")
				dsl.ServerSentEvents("values")
			})
		})
	})
}

// runGeneratedSSEFieldWireTests writes the generated packages and executes the
// server and client tests inside them.
func runGeneratedSSEFieldWireTests(t *testing.T, files []*codegen.File) {
	t.Helper()
	directory := t.TempDir()
	repository, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	module := "module generated.local\n\ngo 1.25\n\n" +
		"require goa.design/goa/v3 v3.0.0\n\n" +
		"replace goa.design/goa/v3 => " + filepath.ToSlash(repository) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0o600))
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}

	serverTest := filepath.Join(directory, "gen", "http", "sse_wire", "server", "wire_test.go")
	clientTest := filepath.Join(directory, "gen", "http", "sse_wire", "client", "wire_test.go")
	require.NoError(t, os.WriteFile(serverTest, []byte(generatedSSEServerWireTest), 0o600))
	require.NoError(t, os.WriteFile(clientTest, []byte(generatedSSEClientWireTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/http/sse_wire/server", "./gen/http/sse_wire/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated SSE wire tests:\n%s", output)
}

const generatedSSEServerWireTest = `package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/sse_wire"
)

func TestPrimitiveFieldsUseRawSSEData(t *testing.T) {
	tests := []struct {
		name string
		send func() string
		want string
	}{
		{
			name: "string",
			send: func() string {
				recorder := httptest.NewRecorder()
				stream := &RequiredServerStream{w: recorder}
				require.NoError(t, stream.Send(&service.RequiredText{Value: "event"}))
				return recorder.Body.String()
			},
			want: "data: event\n\n",
		},
		{
			name: "string alias",
			send: func() string {
				recorder := httptest.NewRecorder()
				stream := &AliasServerStream{w: recorder}
				require.NoError(t, stream.Send(&service.AliasText{Value: service.EventText("event")}))
				return recorder.Body.String()
			},
			want: "data: event\n\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, test.send())
		})
	}
}

func TestOptionalPrimitiveFieldPreservesPresence(t *testing.T) {
	tests := []struct {
		name string
		value *string
		want string
	}{
		{name: "value", value: stringPointer("event"), want: "data: event\n\n"},
		{name: "empty", value: stringPointer(""), want: "data: \n\n"},
		{name: "absent", want: "\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			stream := &OptionalServerStream{w: recorder}
			require.NoError(t, stream.Send(&service.OptionalText{Value: test.value}))
			require.Equal(t, test.want, recorder.Body.String())
		})
	}
}

func TestStructuredFieldsUseJSON(t *testing.T) {
	objectRecorder := httptest.NewRecorder()
	objectStream := &ObjectServerStream{w: objectRecorder}
	value := &service.WireObject{Label: "event"}
	require.NoError(t, objectStream.Send(&service.Structured{Object: value, Values: []string{"one", "two"}}))
	require.Equal(t, "data: {\"label\":\"event\"}\n\n", objectRecorder.Body.String())

	arrayRecorder := httptest.NewRecorder()
	arrayStream := &ArrayServerStream{w: arrayRecorder}
	require.NoError(t, arrayStream.Send(&service.Structured{Object: value, Values: []string{"one", "two"}}))
	require.Equal(t, "data: [\"one\",\"two\"]\n\n", arrayRecorder.Body.String())
}

func stringPointer(value string) *string {
	return &value
}
`

const generatedSSEClientWireTest = `package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptionalPrimitiveDataAllocatesOnlyWhenPresent(t *testing.T) {
	stream := &OptionalStreamImpl{}

	value, err := stream.processEvent([]byte("data: event\n\n"))
	require.NoError(t, err)
	require.NotNil(t, value.Value)
	require.Equal(t, "event", *value.Value)

	empty, err := stream.processEvent([]byte("data: \n\n"))
	require.NoError(t, err)
	require.NotNil(t, empty.Value)
	require.Empty(t, *empty.Value)

	absent, err := stream.processEvent([]byte("\n\n"))
	require.NoError(t, err)
	require.Nil(t, absent.Value)
}
`
