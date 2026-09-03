// This file renders HTTP clients with caller-selected, fixed, and ordinary
// response bodies. The generated tests prove each client validates the exact
// body shape selected by its design and response headers.
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

// TestGeneratedViewedHTTPClientValidatesSelectedBody renders and runs clients
// whose response views select different required fields.
func TestGeneratedViewedHTTPClientValidatesSelectedBody(t *testing.T) {
	root := expr.RunDSL(t, viewedHTTPClientValidationDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	mixed := httpPlans[0].services.Get("viewed_validation").Endpoint("mixed")
	require.NotNil(t, mixed)
	require.True(t, mixed.Result.Responses[0].SelectClientBodyByView)
	require.NotEmpty(t, mixed.Result.Responses[0].ViewedRepresentations)
	require.NotNil(t, mixed.SSE)
	require.NotNil(t, mixed.SSE.Response)
	require.NotNil(t, mixed.SSE.Response.ClientBody)
	require.False(t, mixed.SSE.Response.SelectClientBodyByView)
	require.Empty(t, mixed.SSE.Response.ViewedRepresentations)

	serviceFiles, err := service.Files(servicePlan)
	require.NoError(t, err)
	files := slices.Clone(serviceFiles)
	files = append(files, httpPlans[0].ClientFiles()...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	runGeneratedViewedHTTPClientValidationTest(t, files)
}

// viewedHTTPClientValidationDSL defines one result whose tiny view omits a
// required nested object, plus fixed-view, selected-field, and ordinary
// methods that keep their existing response rules.
func viewedHTTPClientValidationDSL() {
	details := dsl.Type("Details", func() {
		dsl.Attribute("name", dsl.String)
		dsl.Required("name")
	})
	viewed := dsl.ResultType("application/vnd.viewed-http-validation", func() {
		dsl.TypeName("ViewedResult")
		dsl.Attribute("id", dsl.String, func() {
			dsl.MinLength(3)
		})
		dsl.Attribute("details", details)
		dsl.Required("id", "details")
		dsl.View("default", func() {
			dsl.Attribute("id")
			dsl.Attribute("details")
		})
		dsl.View("tiny", func() {
			dsl.Attribute("id")
		})
	})
	streamBody := dsl.Type("StreamBody", func() {
		dsl.Attribute("id", dsl.String, func() {
			dsl.MinLength(3)
		})
		dsl.Required("id")
	})
	explicitViewed := dsl.ResultType("application/vnd.explicit-viewed-http-validation", func() {
		dsl.TypeName("ExplicitViewedResult")
		dsl.Attribute("id", dsl.String, func() {
			dsl.MinLength(3)
		})
		dsl.Attribute("details", details)
		dsl.Required("id")
		dsl.View("default", func() {
			dsl.Attribute("id")
			dsl.Attribute("details")
		})
		dsl.View("tiny", func() {
			dsl.Attribute("id")
		})
	})

	dsl.Service("viewed_validation", func() {
		dsl.Method("dynamic", func() {
			dsl.Result(viewed)
			dsl.HTTP(func() {
				dsl.GET("/dynamic")
			})
		})
		dsl.Method("dynamic_collection", func() {
			dsl.Result(dsl.CollectionOf(viewed))
			dsl.HTTP(func() {
				dsl.GET("/dynamic-collection")
			})
		})
		dsl.Method("stream_dynamic", func() {
			dsl.StreamingResult(viewed)
			dsl.HTTP(func() {
				dsl.GET("/stream-dynamic")
			})
		})
		dsl.Method("stream_fixed", func() {
			dsl.StreamingResult(viewed, func() {
				dsl.View("tiny")
			})
			dsl.HTTP(func() {
				dsl.GET("/stream-fixed")
			})
		})
		dsl.Method("stream_explicit", func() {
			dsl.StreamingResult(explicitViewed)
			dsl.HTTP(func() {
				dsl.GET("/stream-explicit")
				dsl.Response(func() {
					dsl.Body(streamBody)
				})
			})
		})
		dsl.Method("fixed", func() {
			dsl.Result(viewed, func() {
				dsl.View("tiny")
			})
			dsl.HTTP(func() {
				dsl.GET("/fixed")
			})
		})
		dsl.Method("selected", func() {
			dsl.Result(explicitViewed)
			dsl.HTTP(func() {
				dsl.GET("/selected")
				dsl.Response(func() {
					dsl.Body("id")
				})
			})
		})
		dsl.Method("ordinary", func() {
			dsl.Result(func() {
				dsl.Attribute("id", dsl.String)
				dsl.Attribute("details", details)
				dsl.Required("id", "details")
			})
			dsl.HTTP(func() {
				dsl.GET("/ordinary")
			})
		})
		dsl.Method("mixed", func() {
			dsl.Result(viewed)
			dsl.StreamingResult(func() {
				dsl.Attribute("message", dsl.String)
				dsl.Required("message")
			})
			dsl.HTTP(func() {
				dsl.GET("/mixed")
				dsl.ServerSentEvents()
			})
		})
	})
}

// runGeneratedViewedHTTPClientValidationTest writes the generated client into
// an isolated module and runs its response decoders against representative
// bodies and view headers.
func runGeneratedViewedHTTPClientValidationTest(t *testing.T, files []*codegen.File) {
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

	testPath := filepath.Join(directory, "gen", "http", "viewed_validation", "client", "viewed_validation_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(generatedViewedHTTPClientValidationTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/http/viewed_validation/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated viewed HTTP client test:\n%s", output)
}

const generatedViewedHTTPClientValidationTest = `package client

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	goahttp "goa.design/goa/v3/http"
)

func TestDynamicResponseUsesSelectedView(t *testing.T) {
	tests := []struct {
		name      string
		view      string
		body      string
		wantError string
	}{
		{name: "valid tiny", view: "tiny", body: ` + "`" + `{"id":"valid"}` + "`" + `},
		{name: "valid default", view: "default", body: ` + "`" + `{"id":"valid","details":{"name":"ready"}}` + "`" + `},
		{
			name:      "default requires nested object",
			view:      "default",
			body:      ` + "`" + `{"id":"valid"}` + "`" + `,
			wantError: "details",
		},
		{name: "tiny validates selected field", view: "tiny", body: ` + "`" + `{"id":"x"}` + "`" + `, wantError: "id"},
		{name: "missing view uses default", body: ` + "`" + `{"id":"valid","details":{"name":"ready"}}` + "`" + `},
		{name: "missing view validates default", body: ` + "`" + `{"id":"valid"}` + "`" + `, wantError: "details"},
		{name: "unknown view", view: "unknown", body: "{", wantError: "view"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := DecodeDynamicResponse(goahttp.ResponseDecoder, false)(response(test.view, test.body))
			if test.wantError == "" {
				require.NoError(t, err)
				require.NotNil(t, result)
				return
			}
			require.Nil(t, result)
			require.ErrorContains(t, err, test.wantError)
		})
	}
}

func TestFixedResponseKeepsFixedView(t *testing.T) {
	result, err := DecodeFixedResponse(goahttp.ResponseDecoder, false)(response("", ` + "`" + `{"id":"valid"}` + "`" + `))
	require.NoError(t, err)
	require.NotNil(t, result)

	result, err = DecodeFixedResponse(goahttp.ResponseDecoder, false)(response("", ` + "`" + `{"id":"x"}` + "`" + `))
	require.Nil(t, result)
	require.ErrorContains(t, err, "id")
}

func TestDynamicCollectionUsesSelectedView(t *testing.T) {
	result, err := DecodeDynamicCollectionResponse(goahttp.ResponseDecoder, false)(
		response("tiny", ` + "`" + `[{"id":"valid"}]` + "`" + `),
	)
	require.NoError(t, err)
	require.NotNil(t, result)

	result, err = DecodeDynamicCollectionResponse(goahttp.ResponseDecoder, false)(
		response("default", ` + "`" + `[{"id":"valid"}]` + "`" + `),
	)
	require.Nil(t, result)
	require.ErrorContains(t, err, "details")
}

func TestMixedOrdinaryResponseUsesSelectedView(t *testing.T) {
	tests := []struct {
		name      string
		view      string
		body      string
		wantError string
	}{
		{name: "valid tiny", view: "tiny", body: ` + "`" + `{"id":"valid"}` + "`" + `},
		{name: "valid default", body: ` + "`" + `{"id":"valid","details":{"name":"ready"}}` + "`" + `},
		{name: "default requires nested object", body: ` + "`" + `{"id":"valid"}` + "`" + `, wantError: "details"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := DecodeMixedResponse(goahttp.ResponseDecoder, false)(response(test.view, test.body))
			if test.wantError == "" {
				require.NoError(t, err)
				require.NotNil(t, result)
				return
			}
			require.Nil(t, result)
			require.ErrorContains(t, err, test.wantError)
		})
	}
}

func TestDynamicWebSocketUsesSelectedView(t *testing.T) {
	tests := []struct {
		name      string
		view      string
		body      string
		wantError string
	}{
		{name: "valid tiny", view: "tiny", body: ` + "`" + `{"id":"valid"}` + "`" + `},
		{name: "valid default", body: ` + "`" + `{"id":"valid","details":{"name":"ready"}}` + "`" + `},
		{name: "default requires nested object", body: ` + "`" + `{"id":"valid"}` + "`" + `, wantError: "details"},
		{name: "tiny validates selected field", view: "tiny", body: ` + "`" + `{"id":"x"}` + "`" + `, wantError: "id"},
		{name: "unknown view", view: "unknown", body: "{", wantError: "view"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := &StreamDynamicClientStream{
				conn: webSocketConnection(t, test.body),
				view: test.view,
			}
			result, err := stream.Recv()
			if test.wantError == "" {
				require.NoError(t, err)
				require.NotNil(t, result)
				return
			}
			require.Nil(t, result)
			require.ErrorContains(t, err, test.wantError)
		})
	}
}

func TestFixedWebSocketKeepsFixedView(t *testing.T) {
	stream := &StreamFixedClientStream{conn: webSocketConnection(t, ` + "`" + `{"id":"valid"}` + "`" + `)}
	result, err := stream.Recv()
	require.NoError(t, err)
	require.NotNil(t, result)

	stream = &StreamFixedClientStream{conn: webSocketConnection(t, ` + "`" + `{"id":"x"}` + "`" + `)}
	result, err = stream.Recv()
	require.Nil(t, result)
	require.ErrorContains(t, err, "id")
}

func TestExplicitBodyWebSocketKeepsOneBodyShape(t *testing.T) {
	stream := &StreamExplicitClientStream{
		conn: webSocketConnection(t, ` + "`" + `{"id":"valid"}` + "`" + `),
	}
	result, err := stream.Recv()
	require.NoError(t, err)
	require.NotNil(t, result)

	stream = &StreamExplicitClientStream{
		conn: webSocketConnection(t, ` + "`" + `{"id":"x"}` + "`" + `),
		view: "tiny",
	}
	result, err = stream.Recv()
	require.Nil(t, result)
	require.ErrorContains(t, err, "id")

	stream = &StreamExplicitClientStream{
		conn: webSocketConnection(t, ` + "`" + `{"id":"valid"}` + "`" + `),
		view: "unknown",
	}
	result, err = stream.Recv()
	require.Nil(t, result)
	require.ErrorContains(t, err, "view")
	require.ErrorContains(t, err, "unknown")
}

func TestSelectedFieldResponseKeepsOneBodyShape(t *testing.T) {
	decode := DecodeSelectedResponse(goahttp.ResponseDecoder, false)
	result, err := decode(response("", ` + "`" + `"valid"` + "`" + `))
	require.NoError(t, err)
	require.NotNil(t, result)

	result, err = decode(response("tiny", ` + "`" + `"valid"` + "`" + `))
	require.NoError(t, err)
	require.NotNil(t, result)

	result, err = decode(response("tiny", ` + "`" + `"x"` + "`" + `))
	require.Nil(t, result)
	require.ErrorContains(t, err, "id")

	result, err = decode(response("unknown", ` + "`" + `"valid"` + "`" + `))
	require.Nil(t, result)
	require.ErrorContains(t, err, "view")
	require.ErrorContains(t, err, "unknown")
}

func TestOrdinaryResponseStillValidatesFullBody(t *testing.T) {
	decode := DecodeOrdinaryResponse(goahttp.ResponseDecoder, false)
	result, err := decode(response("", ` + "`" + `{"id":"valid"}` + "`" + `))
	require.Nil(t, result)
	require.ErrorContains(t, err, "details")
}

func response(view, body string) *http.Response {
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	if view != "" {
		response.Header.Set("goa-view", view)
	}
	return response
}

func webSocketConnection(t *testing.T, body string) *websocket.Conn {
	t.Helper()
	serverErr := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		connection, err := (&websocket.Upgrader{}).Upgrade(writer, request, nil)
		if err != nil {
			serverErr <- err
			return
		}
		writeErr := connection.WriteMessage(websocket.TextMessage, []byte(body))
		serverErr <- errors.Join(writeErr, connection.Close())
	}))
	t.Cleanup(server.Close)

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	require.NoError(t, <-serverErr)
	t.Cleanup(func() {
		require.NoError(t, connection.Close())
	})
	return connection
}
`
