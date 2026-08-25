// This file renders an ordinary HTTP SSE client into a temporary module. The
// generated test proves required JSON fields are checked before an event is
// returned through the service interface.
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

// TestGeneratedSSEClientValidatesResponseBody checks that a required zero-valued
// field is accepted when present and rejected when missing.
func TestGeneratedSSEClientValidatesResponseBody(t *testing.T) {
	root := expr.RunDSL(t, sseRequiredPrimitiveDSL)
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
	runGeneratedSSEClientValidationTest(t, files)
}

// sseRequiredPrimitiveDSL defines a required primitive whose zero value must
// remain different from an absent JSON property.
func sseRequiredPrimitiveDSL() {
	viewed := dsl.ResultType("application/vnd.viewed-sse-validation", func() {
		dsl.TypeName("ViewedEvent")
		dsl.Attribute("id", dsl.String)
		dsl.Attribute("kind", dsl.String, func() {
			dsl.Default("default-kind")
		})
		dsl.Attribute("retry", dsl.Int, func() {
			dsl.Default(5)
		})
		dsl.Attribute("data", dsl.String)
		dsl.Required("id", "data")
		dsl.View("default", func() {
			dsl.Attribute("id")
			dsl.Attribute("kind")
			dsl.Attribute("retry")
			dsl.Attribute("data")
		})
	})
	dsl.Service("SSE Validation", func() {
		dsl.Method("watch", func() {
			dsl.StreamingResult(func() {
				dsl.Attribute("count", dsl.Int)
				dsl.Required("count")
			})
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
	})
	dsl.Service("Viewed SSE Validation", func() {
		dsl.Method("watch", func() {
			dsl.StreamingResult(viewed)
			dsl.HTTP(func() {
				dsl.GET("/viewed")
				dsl.ServerSentEvents("data", func() {
					dsl.SSEEventID("id")
					dsl.SSEEventType("kind")
					dsl.SSEEventRetry("retry")
				})
			})
		})
	})
}

// runGeneratedSSEClientValidationTest renders the generated packages and runs
// the generated client's event parser against valid and incomplete JSON.
func runGeneratedSSEClientValidationTest(t *testing.T, files []*codegen.File) {
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

	testPath := filepath.Join(directory, "gen", "http", "sse_validation", "client", "sse_validation_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(generatedSSEClientValidationTest), 0o600))
	viewedTestPath := filepath.Join(directory, "gen", "http", "viewed_sse_validation", "client", "viewed_sse_validation_test.go")
	require.NoError(t, os.WriteFile(viewedTestPath, []byte(generatedViewedSSEClientValidationTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/http/sse_validation/client", "./gen/http/viewed_sse_validation/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated SSE client test:\n%s", output)
}

const generatedSSEClientValidationTest = `package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	goahttp "goa.design/goa/v3/http"
)

func TestRequiredPrimitivePresence(t *testing.T) {
	stream := &WatchStreamImpl{decoder: goahttp.ResponseDecoder}

	event, hasData, err := stream.processEvent([]byte("data: {\"count\":0}\n\n"))
	require.NoError(t, err)
	require.True(t, hasData)
	require.Equal(t, 0, event.Count)

	_, _, err = stream.processEvent([]byte("data: {}\n\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "count")
}
`

const generatedViewedSSEClientValidationTest = `package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOuterFieldsAreValidatedWithTheBody(t *testing.T) {
	stream := &WatchStreamImpl{}

	event, hasData, err := stream.processEvent([]byte("id:\nevent:\nretry: 0\ndata: ready\n\n"))
	require.NoError(t, err)
	require.True(t, hasData)
	require.Empty(t, event.ID)
	require.Empty(t, event.Kind)
	require.Zero(t, event.Retry)
	require.Equal(t, "ready", event.Data)

	event, hasData, err = stream.processEvent([]byte("id: event-1\ndata: ready\n\n"))
	require.NoError(t, err)
	require.True(t, hasData)
	require.Equal(t, "event-1", event.ID)
	require.Equal(t, "default-kind", event.Kind)
	require.Equal(t, 5, event.Retry)

	stream = &WatchStreamImpl{}
	_, _, err = stream.processEvent([]byte("data: ready\n\n"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "id")
}
`
