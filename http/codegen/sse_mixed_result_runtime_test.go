// This file renders a mixed HTTP/SSE client into a temporary module. The
// generated test proves mapped SSE fields are validated before the wire event
// becomes the service event returned by Recv.
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

// TestGeneratedMixedSSEClientValidatesMappedWireBody catches clients decoding
// event data directly into a service value and skipping transport validation.
func TestGeneratedMixedSSEClientValidatesMappedWireBody(t *testing.T) {
	root := expr.RunDSL(t, mixedSSEMappedFieldDSL)
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
	files = append(files, httpPlans[0].ClientFiles()...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	runGeneratedMixedSSEClientTest(t, files)
}

// TestGeneratedMixedSSEResultShapesCompile verifies each generation-time
// conversion branch produces complete client and server packages: direct
// primitive values, direct primitive collections, converted anonymous objects,
// and empty events.
func TestGeneratedMixedSSEResultShapesCompile(t *testing.T) {
	root := expr.RunDSL(t, mixedSSEResultShapesDSL)
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
	files = append(files, httpPlans[0].ClientFiles()...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].ServerFiles()...)
	files = append(files, httpPlans[0].ServerTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	runGeneratedMixedSSECompile(t, files, "./gen/...")
}

// mixedSSEResultShapesDSL puts every optional-transform shape in one generated
// client package so one compile checks their declarations and return paths.
func mixedSSEResultShapesDSL() {
	dsl.Service("Mixed SSE Shapes", func() {
		shapes := []struct {
			name   string
			result any
		}{
			{"int", dsl.Int},
			{"ints", dsl.ArrayOf(dsl.Int)},
			{"inline", func() { dsl.Attribute("value", dsl.Int) }},
			{"empty", func() {}},
		}
		for _, shape := range shapes {
			dsl.Method("watch_"+shape.name, func() {
				dsl.Result(dsl.String)
				dsl.StreamingResult(shape.result)
				dsl.HTTP(func() {
					dsl.GET("/" + shape.name)
					dsl.ServerSentEvents()
				})
			})
		}
	})
}

// mixedSSEMappedFieldDSL makes event_id required and maps it to the SSE id
// line while message is carried by the data line.
func mixedSSEMappedFieldDSL() {
	result := dsl.Type("Result", func() {
		dsl.Attribute("event_id", dsl.String)
		dsl.Attribute("message", dsl.String)
		dsl.Required("event_id", "message")
	})
	event := dsl.Type("Event", func() {
		dsl.Attribute("event_id", dsl.String)
		dsl.Attribute("message", dsl.String)
		dsl.Required("event_id", "message")
	})
	dsl.Service("Mixed SSE Wire", func() {
		dsl.Method("watch", func() {
			dsl.Result(result)
			dsl.StreamingResult(event)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents("message", func() {
					dsl.SSEEventID("event_id")
				})
			})
		})
	})
}

// runGeneratedMixedSSEClientTest writes generated packages and runs the
// generated client's private event parser against complete and incomplete
// frames.
func runGeneratedMixedSSEClientTest(t *testing.T, files []*codegen.File) {
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

	testPath := filepath.Join(directory, "gen", "http", "mixed_sse_wire", "client", "mixed_sse_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(generatedMixedSSEClientTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/http/mixed_sse_wire/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated mixed SSE client test:\n%s", output)
}

// runGeneratedMixedSSECompile renders files in an isolated module and compiles
// the requested generated package.
func runGeneratedMixedSSECompile(t *testing.T, files []*codegen.File, packagePath string) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", packagePath)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "compile generated mixed SSE clients:\n%s", output)
}

const generatedMixedSSEClientTest = `package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMappedEventBodyIsValidated(t *testing.T) {
	stream := &WatchStreamImpl{}

	event, err := stream.processEvent([]byte("id: event-1\ndata: ready\n\n"))
	require.NoError(t, err)
	require.Equal(t, "event-1", event.EventID)
	require.Equal(t, "ready", event.Message)

	_, err = stream.processEvent([]byte("data: ready\n\n"))
	require.Error(t, err)
}
`
