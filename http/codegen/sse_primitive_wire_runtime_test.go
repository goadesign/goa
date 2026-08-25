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

// TestSSEClientFieldPlan checks the HTTP body type and presence shape chosen
// before templates render mapped SSE values.
func TestSSEClientFieldPlan(t *testing.T) {
	eventText := &expr.UserTypeExpr{
		AttributeExpr: &expr.AttributeExpr{Type: expr.String},
		TypeName:      "EventText",
		UID:           "EventText",
	}
	tests := []struct {
		name        string
		field       *expr.AttributeExpr
		required    bool
		typeRef     string
		wantTypeRef string
		wantPointer bool
	}{
		{name: "required alias", field: &expr.AttributeExpr{Type: eventText}, required: true, typeRef: "service.EventText", wantTypeRef: "string", wantPointer: true},
		{name: "optional alias", field: &expr.AttributeExpr{Type: eventText}, typeRef: "service.EventText", wantTypeRef: "string", wantPointer: true},
		{name: "bytes", field: &expr.AttributeExpr{Type: expr.Bytes}, typeRef: "[]byte", wantTypeRef: "[]byte"},
		{name: "any", field: &expr.AttributeExpr{Type: expr.Any}, typeRef: "any", wantTypeRef: "any"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object := expr.Object{&expr.NamedAttributeExpr{Name: "value", Attribute: test.field}}
			event := &expr.AttributeExpr{Type: &object}
			if test.required {
				event.Validation = &expr.ValidationExpr{}
				event.Validation.AddRequired("value")
			}
			value := SSEValueData{TypeRef: test.typeRef, ClientTypeRef: test.typeRef}

			setSSEClientField(&value, &object, "value")

			require.Equal(t, test.required, event.IsRequired("value"))
			require.Equal(t, test.wantTypeRef, value.ClientTypeRef)
			require.Equal(t, test.wantPointer, value.ClientPointer)
		})
	}
}

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
	framingEvent := dsl.Type("FramingEvent", func() {
		dsl.Attribute("id", dsl.String)
		dsl.Attribute("event", dsl.String)
		dsl.Attribute("retry", dsl.Int)
		dsl.Attribute("data", dsl.Bytes)
		dsl.Required("id", "event", "retry", "data")
	})
	presenceEvent := dsl.Type("PresenceEvent", func() {
		dsl.Attribute("id", dsl.String)
		dsl.Attribute("event", dsl.String)
		dsl.Attribute("retry", dsl.Int)
		dsl.Attribute("data", dsl.Bytes)
		dsl.Required("data")
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
		dsl.Method("framing", func() {
			dsl.StreamingResult(framingEvent)
			dsl.HTTP(func() {
				dsl.GET("/framing")
				dsl.ServerSentEvents("data", func() {
					dsl.SSEEventID("id")
					dsl.SSEEventType("event")
					dsl.SSEEventRetry("retry")
				})
			})
		})
		dsl.Method("presence", func() {
			dsl.StreamingResult(presenceEvent)
			dsl.HTTP(func() {
				dsl.GET("/presence")
				dsl.ServerSentEvents("data", func() {
					dsl.SSEEventID("id")
					dsl.SSEEventType("event")
					dsl.SSEEventRetry("retry")
				})
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

func TestMappedFieldsCannotChangeFrame(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		event string
	}{
		{name: "id newline", id: "one\ntwo", event: "update"},
		{name: "id carriage return", id: "one\rtwo", event: "update"},
		{name: "id null", id: "one\x00two", event: "update"},
		{name: "event newline", id: "event-1", event: "one\ntwo"},
		{name: "event carriage return", id: "event-1", event: "one\rtwo"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			stream := &FramingServerStream{w: recorder}
			err := stream.Send(&service.FramingEvent{
				ID:    test.id,
				Event: test.event,
				Retry: 1,
				Data:  []byte("ready"),
			})
			require.Error(t, err)
			require.Empty(t, recorder.Body.String())
		})
	}
}

func TestEveryPhysicalDataLineHasPrefix(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := &FramingServerStream{w: recorder}
	err := stream.Send(&service.FramingEvent{
		ID:    "event-1",
		Event: "update",
		Retry: 1,
		Data:  []byte("one\r\ntwo\rthree\nfour"),
	})
	require.NoError(t, err)
	require.Equal(t, "id: event-1\nevent: update\nretry: 1\ndata: one\ndata: two\ndata: three\ndata: four\n\n", recorder.Body.String())
}

func TestMappedFieldPresenceIsPreserved(t *testing.T) {
	tests := []struct {
		name  string
		value *string
		retry *int
		want  string
	}{
		{name: "absent", want: "data: ready\n\n"},
		{
			name:  "selected zero values",
			value: stringPointer(""),
			retry: intPointer(0),
			want:  "id: \nevent: \nretry: 0\ndata: ready\n\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			stream := &PresenceServerStream{w: recorder}
			err := stream.Send(&service.PresenceEvent{
				ID:    test.value,
				Event: test.value,
				Retry: test.retry,
				Data:  []byte("ready"),
			})
			require.NoError(t, err)
			require.Equal(t, test.want, recorder.Body.String())
		})
	}
}

func TestNegativeRetryIsRejectedBeforeHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	stream := &FramingServerStream{w: recorder}
	err := stream.Send(&service.FramingEvent{
		ID:    "event-1",
		Event: "update",
		Retry: -1,
		Data:  []byte("ready"),
	})
	require.Error(t, err)
	require.Empty(t, recorder.Body.String())
}

func stringPointer(value string) *string {
	return &value
}

func intPointer(value int) *int {
	return &value
}
`

const generatedSSEClientWireTest = `package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptionalPrimitiveDataAllocatesOnlyWhenPresent(t *testing.T) {
	stream := &OptionalStreamImpl{}

	value, hasData, err := stream.processEvent([]byte("data: event\n\n"))
	require.NoError(t, err)
	require.True(t, hasData)
	require.NotNil(t, value.Value)
	require.Equal(t, "event", *value.Value)

	empty, hasData, err := stream.processEvent([]byte("data: \n\n"))
	require.NoError(t, err)
	require.True(t, hasData)
	require.NotNil(t, empty.Value)
	require.Empty(t, *empty.Value)

	absent, hasData, err := stream.processEvent([]byte("\n\n"))
	require.NoError(t, err)
	require.False(t, hasData)
	require.Nil(t, absent)
}

func TestUnterminatedEventIsDiscarded(t *testing.T) {
	for _, frame := range []string{"data: event\n", "data: event\r"} {
		stream := &OptionalStreamImpl{
			resp: &http.Response{Body: io.NopCloser(strings.NewReader(frame))},
		}
		event, err := stream.readEvent(context.Background())
		require.ErrorIs(t, err, io.EOF)
		require.Nil(t, event)
	}
}

func TestEverySSELineEndingIsAccepted(t *testing.T) {
	for _, ending := range []string{"\n\n", "\r\r", "\r\n\r\n", "\r\n\n"} {
		stream := &OptionalStreamImpl{
			resp: &http.Response{Body: io.NopCloser(strings.NewReader("data: event" + ending))},
		}
		frame, err := stream.readEvent(context.Background())
		require.NoError(t, err)
		event, hasData, err := stream.processEvent(frame)
		require.NoError(t, err)
		require.True(t, hasData)
		require.NotNil(t, event.Value)
		require.Equal(t, "event", *event.Value)
	}
}

func TestCRLFMaySpanReads(t *testing.T) {
	stream := &OptionalStreamImpl{
		resp: &http.Response{Body: io.NopCloser(&chunkReader{chunks: [][]byte{
			[]byte("data: event\r"),
			[]byte("\n\r"),
			[]byte("\n"),
		}})},
	}
	frame, err := stream.readEvent(context.Background())
	require.NoError(t, err)
	require.Equal(t, "data: event\r\n\r\n", string(frame))
	event, hasData, err := stream.processEvent(frame)
	require.NoError(t, err)
	require.True(t, hasData)
	require.NotNil(t, event.Value)
	require.Equal(t, "event", *event.Value)
}

func TestDataWithoutColonSelectsAnEmptyValue(t *testing.T) {
	stream := &OptionalStreamImpl{}
	event, hasData, err := stream.processEvent([]byte("data\n\n"))
	require.NoError(t, err)
	require.True(t, hasData)
	require.NotNil(t, event.Value)
	require.Empty(t, *event.Value)
}

func TestRecvSkipsEventsWithoutData(t *testing.T) {
	stream := &OptionalStreamImpl{
		resp: &http.Response{Body: io.NopCloser(strings.NewReader("event: ignored\n\ndata: ready\n\n"))},
	}
	event, err := stream.Recv()
	require.NoError(t, err)
	require.NotNil(t, event.Value)
	require.Equal(t, "ready", *event.Value)
}

func TestRetryUsesOnlyValidDigitFields(t *testing.T) {
	valid := &FramingStreamImpl{}
	event, hasData, err := valid.processEvent([]byte("id: event-1\nevent: update\nretry: 0\ndata: ready\n\n"))
	require.NoError(t, err)
	require.True(t, hasData)
	require.Equal(t, 0, event.Retry)

	for _, retry := range []string{"", "+1", "-0", "١"} {
		stream := &PresenceStreamImpl{}
		event, hasData, err := stream.processEvent([]byte("retry: " + retry + "\ndata: ready\n\n"))
		require.NoError(t, err)
		require.True(t, hasData)
		require.Nil(t, event.Retry)
		require.Equal(t, []byte("ready"), event.Data)
	}

	for _, fields := range []string{
		"retry: later\nretry: 7\n",
		"retry: 7\nretry: later\n",
	} {
		stream := &FramingStreamImpl{}
		event, hasData, err := stream.processEvent([]byte("id: event-1\nevent: update\n" + fields + "data: ready\n\n"))
		require.NoError(t, err)
		require.True(t, hasData)
		require.Equal(t, 7, event.Retry)
	}
}

func TestOnlyTheFirstStreamLineMayStartWithBOM(t *testing.T) {
	stream := &OptionalStreamImpl{}
	event, hasData, err := stream.processEvent([]byte("\uFEFFdata: first\n\n"))
	require.NoError(t, err)
	require.True(t, hasData)
	require.Equal(t, "first", *event.Value)

	_, hasData, err = stream.processEvent([]byte("\uFEFFdata: second\n\n"))
	require.NoError(t, err)
	require.False(t, hasData)
}

func TestEventIDPersistsUntilAnEmptyValidIDResetsIt(t *testing.T) {
	body := strings.Join([]string{
		"\uFEFFid: item-1\ndata: first\n\n",
		"data: second\n\n",
		"\uFEFFid: later\ndata: third\n\n",
		"id: invalid\x00id\ndata: fourth\n\n",
		"id:\ndata: fifth\n\n",
		"data: sixth\n\n",
	}, "")
	stream := &PresenceStreamImpl{
		resp: &http.Response{Body: io.NopCloser(strings.NewReader(body))},
	}

	tests := []struct {
		id   string
		data string
	}{
		{id: "item-1", data: "first"},
		{id: "item-1", data: "second"},
		{id: "item-1", data: "third"},
		{id: "item-1", data: "fourth"},
		{id: "", data: "fifth"},
		{id: "", data: "sixth"},
	}
	for _, test := range tests {
		event, err := stream.Recv()
		require.NoError(t, err)
		require.NotNil(t, event.ID)
		require.Equal(t, test.id, *event.ID)
		require.Equal(t, []byte(test.data), event.Data)
	}
}

type chunkReader struct {
	chunks [][]byte
}

func (r *chunkReader) Read(buffer []byte) (int, error) {
	if len(r.chunks) == 0 {
		return 0, io.EOF
	}
	n := copy(buffer, r.chunks[0])
	r.chunks = r.chunks[1:]
	return n, nil
}
`
