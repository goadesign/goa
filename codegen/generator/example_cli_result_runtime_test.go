// This file runs the result writers emitted into generated example clients.
// The test covers values received from a server stream and the errors returned
// when receiving or writing fails.
package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
)

// TestExampleCLIResultWritersRun verifies the generated helpers against real
// endpoint functions, stream receive functions, and writers.
func TestExampleCLIResultWritersRun(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.API("example stream", func() {
			dsl.Server("stream", func() {
				dsl.Services("events")
				dsl.Host("local", func() {
					dsl.URI("http://localhost:8080")
				})
			})
		})
		dsl.Service("events", func() {
			dsl.Method("watch", func() {
				dsl.StreamingResult(dsl.String)
				dsl.HTTP(func() {
					dsl.GET("/events")
				})
			})
			dsl.Method("upload", func() {
				dsl.StreamingPayload(dsl.String)
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/events")
				})
			})
			dsl.Method("create", func() {
				dsl.Result(dsl.String)
				dsl.StreamingResult(dsl.Int)
				dsl.HTTP(func() {
					dsl.POST("/create")
					dsl.ServerSentEvents()
				})
			})
		})
	})
	plan := mustTestPlan(t, "generated.local/gen", []eval.Root{root}, planExampleData)
	files, err := testServiceFiles(plan)
	require.NoError(t, err)
	transportFiles, err := testTransportFiles(plan)
	require.NoError(t, err)
	files = append(files, transportFiles...)
	exampleFiles, err := assembleExampleFilesForTest(plan)
	require.NoError(t, err)
	for _, file := range exampleFiles {
		if strings.HasPrefix(file.Path, filepath.Join("cmd", "stream-cli")) {
			files = append(files, file)
		}
	}
	files, err = mergeFilesByPath(files)
	require.NoError(t, err)

	directory := t.TempDir()
	writeGeneratedModule(t, directory, "generated.local")
	for _, file := range files {
		_, err := file.Render(directory)
		require.NoError(t, err)
	}
	testPath := filepath.Join(directory, "cmd", "stream-cli", "result_writer_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(exampleCLIResultWriterTest), 0o600))
	runGeneratedTests(t, directory)
}

const exampleCLIResultWriterTest = `package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	goa "goa.design/goa/v3/pkg"
)

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestWriteEndpointResult(t *testing.T) {
	endpoint := goa.Endpoint(func(context.Context, any) (any, error) {
		return map[string]string{"message": "hello"}, nil
	})
	var output bytes.Buffer
	if err := writeEndpointResult(context.Background(), &output, endpoint, nil); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "{\n    \"message\": \"hello\"\n}\n"; got != want {
		t.Fatalf("unexpected output:\n%s", got)
	}
}

func TestWriteEndpointResultReturnsEncodingError(t *testing.T) {
	endpoint := goa.Endpoint(func(context.Context, any) (any, error) {
		return func() {}, nil
	})
	err := writeEndpointResult(context.Background(), io.Discard, endpoint, nil)
	if err == nil {
		t.Fatal("expected JSON encoding error")
	}
}

func TestWriteStreamResults(t *testing.T) {
	values := []string{"first", "second"}
	next := 0
	recv := func(context.Context) (string, error) {
		if next == len(values) {
			return "", io.EOF
		}
		value := values[next]
		next++
		return value, nil
	}
	var output bytes.Buffer
	if err := writeStreamResults(context.Background(), &output, recv); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "\"first\"\n\"second\"\n"; got != want {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestWriteStreamResultsReturnsReceiveError(t *testing.T) {
	want := errors.New("receive failed")
	recv := func(context.Context) (string, error) {
		return "", want
	}
	err := writeStreamResults(context.Background(), io.Discard, recv)
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped receive error", err)
	}
}

func TestWriteStreamResultsDoesNotHideFailureJoinedWithEOF(t *testing.T) {
	want := errors.New("close failed")
	recv := func(context.Context) (string, error) {
		return "", errors.Join(io.EOF, want)
	}
	err := writeStreamResults(context.Background(), io.Discard, recv)
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped close error", err)
	}
}

func TestWriteStreamResultsReturnsOutputError(t *testing.T) {
	want := errors.New("write failed")
	called := false
	recv := func(context.Context) (string, error) {
		if called {
			return "", io.EOF
		}
		called = true
		return "value", nil
	}
	err := writeStreamResults(context.Background(), failingWriter{err: want}, recv)
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped output error", err)
	}
}

func TestInputStreamIsRejectedBeforeCallingEndpoint(t *testing.T) {
	args := os.Args
	commandLine := flag.CommandLine
	defer func() {
		os.Args = args
		flag.CommandLine = commandLine
	}()
	os.Args = []string{"stream-cli", "events", "upload"}
	flag.CommandLine = flag.NewFlagSet("stream-cli", flag.ContinueOnError)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatal(err)
	}

	err := doHTTP(context.Background(), "http", "127.0.0.1:1", 1, false, io.Discard)
	want := "example client does not support streamed input for service \"events\" method \"upload\""
	if err == nil || err.Error() != want {
		t.Fatalf("got %v, want %q", err, want)
	}
}

func TestMixedHTTPResultUsesNormalResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := io.WriteString(w, "\"created\""); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()

	args := os.Args
	commandLine := flag.CommandLine
	defer func() {
		os.Args = args
		flag.CommandLine = commandLine
	}()
	os.Args = []string{"stream-cli", "events", "create"}
	flag.CommandLine = flag.NewFlagSet("stream-cli", flag.ContinueOnError)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	err := doHTTP(context.Background(), "http", strings.TrimPrefix(server.URL, "http://"), 1, false, &output)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "\"created\"\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
`
