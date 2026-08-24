// This file renders an HTTP client into a temporary module and calls its
// generated endpoints and response decoders. The response bodies can fail
// while reading or closing so the tests can check every returned error.
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
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

// TestGeneratedClientResponseBodyLifecycle checks that generated decoders
// close bodies they consume, preserve bodies requested by callers, and return
// every read, decode, and close error.
func TestGeneratedClientResponseBodyLifecycle(t *testing.T) {
	root := expr.RunDSL(t, responseBodyLifecycleDSL)
	generation, err := codegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(root, generation, expr.NewExampleGenerator(root.API.RandomizerFactory))
	require.NoError(t, err)
	httpPlans, err := NewPlans(generation, PlanInput{Root: root, Service: servicePlan})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())

	clientFiles := httpPlans[0].ClientFiles()
	endpointCode := codegen.SectionsCode(t, clientFiles[0].Section("client-endpoint-init"))
	testutil.AssertGo(t, "testdata/golden/client_endpoint_response_body_lifecycle.go.golden", endpointCode)

	serviceFiles, err := service.Files(servicePlan)
	require.NoError(t, err)
	files := slices.Clone(serviceFiles)
	files = append(files, clientFiles...)
	files = append(files, httpPlans[0].ClientTypeFiles()...)
	files = append(files, httpPlans[0].PathFiles()...)
	runGeneratedResponseBodyLifecycleTest(t, files)
}

// responseBodyLifecycleDSL defines an ordinary response, a response whose
// bytes are returned to the caller, and a server-sent event stream.
func responseBodyLifecycleDSL() {
	dsl.Service("body_lifecycle", func() {
		dsl.Method("read", func() {
			dsl.Result(func() {
				dsl.Attribute("value", dsl.String)
				dsl.Required("value")
			})
			dsl.HTTP(func() {
				dsl.GET("/read")
			})
		})
		dsl.Method("raw", func() {
			dsl.Error("bad", func() {
				dsl.Attribute("message", dsl.String)
				dsl.Required("message")
			})
			dsl.HTTP(func() {
				dsl.GET("/raw")
				dsl.SkipResponseBodyEncodeDecode()
				dsl.Response(dsl.StatusOK)
				dsl.Response("bad", dsl.StatusBadRequest)
			})
		})
		dsl.Method("watch", func() {
			dsl.StreamingResult(dsl.String)
			dsl.HTTP(func() {
				dsl.GET("/watch")
				dsl.ServerSentEvents()
			})
		})
	})
}

// runGeneratedResponseBodyLifecycleTest writes generated code and its runtime
// test into an isolated module, then runs only the generated client package.
func runGeneratedResponseBodyLifecycleTest(t *testing.T, files []*codegen.File) {
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

	testPath := filepath.Join(directory, "gen", "http", "body_lifecycle", "client", "response_body_test.go")
	require.NoError(t, os.WriteFile(testPath, []byte(generatedResponseBodyLifecycleTest), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./gen/http/body_lifecycle/client")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "run generated response body test:\n%s", output)
}

const generatedResponseBodyLifecycleTest = `package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	goahttp "goa.design/goa/v3/http"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (doer doerFunc) Do(request *http.Request) (*http.Response, error) {
	return doer(request)
}

type controlledBody struct {
	reader     io.Reader
	readErr    error
	closeErr   error
	closeCalls int
}

func (body *controlledBody) Read(buffer []byte) (int, error) {
	if body.readErr != nil {
		return 0, body.readErr
	}
	return body.reader.Read(buffer)
}

func (body *controlledBody) Close() error {
	body.closeCalls++
	return body.closeErr
}

func TestRestoreBodyReturnsReadFailureAndClosesOriginal(t *testing.T) {
	readErr := errors.New("read failed")
	body := &controlledBody{reader: strings.NewReader("ignored"), readErr: readErr}
	response := response(http.StatusOK, body)

	_, err := DecodeReadResponse(goahttp.ResponseDecoder, true)(response)

	assertDecodingError(t, err)
	require.ErrorIs(t, err, readErr)
	require.Equal(t, 1, body.closeCalls)
}

func TestDecoderReturnsCloseFailureAfterSuccess(t *testing.T) {
	closeErr := errors.New("close failed")
	body := &controlledBody{reader: strings.NewReader(` + "`" + `{"value":"ready"}` + "`" + `), closeErr: closeErr}
	response := response(http.StatusOK, body)

	result, err := DecodeReadResponse(goahttp.ResponseDecoder, false)(response)

	require.NotNil(t, result)
	assertDecodingError(t, err)
	require.ErrorIs(t, err, closeErr)
	require.Equal(t, 1, body.closeCalls)
}

func TestDecoderReturnsDecodeAndCloseFailures(t *testing.T) {
	decodeErr := errors.New("decode failed")
	closeErr := errors.New("close failed")
	body := &controlledBody{reader: strings.NewReader("ignored"), closeErr: closeErr}
	response := response(http.StatusOK, body)
	decoder := func(*http.Response) goahttp.Decoder {
		return goahttp.EncodingFunc(func(any) error {
			return decodeErr
		})
	}

	_, err := DecodeReadResponse(decoder, false)(response)

	assertDecodingError(t, err)
	require.ErrorIs(t, err, decodeErr)
	require.ErrorIs(t, err, closeErr)
	require.Equal(t, 1, body.closeCalls)
}

func TestRestoreBodyLeavesReadableCopyAndClosesOriginal(t *testing.T) {
	const encoded = ` + "`" + `{"value":"ready"}` + "`" + `
	body := &controlledBody{reader: strings.NewReader(encoded)}
	response := response(http.StatusOK, body)

	result, err := DecodeReadResponse(goahttp.ResponseDecoder, true)(response)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, body.closeCalls)
	restored, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, encoded, string(restored))
}

func TestUnexpectedStatusReturnsReadFailure(t *testing.T) {
	readErr := errors.New("read failed")
	body := &controlledBody{reader: strings.NewReader("ignored"), readErr: readErr}
	response := response(http.StatusTeapot, body)

	_, err := DecodeReadResponse(goahttp.ResponseDecoder, false)(response)

	assertDecodingError(t, err)
	require.ErrorIs(t, err, readErr)
	require.Equal(t, 1, body.closeCalls)
}

func TestRawBodyRemainsCallerOwned(t *testing.T) {
	for _, restoreBody := range []bool{false, true} {
		t.Run(fmt.Sprintf("restoreBody=%t", restoreBody), func(t *testing.T) {
			body := &controlledBody{reader: strings.NewReader("raw bytes")}
			response := response(http.StatusOK, body)

			_, err := DecodeRawResponse(goahttp.ResponseDecoder, restoreBody)(response)

			require.NoError(t, err)
			require.Same(t, body, response.Body)
			require.Zero(t, body.closeCalls)
			content, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, "raw bytes", string(content))
		})
	}
}

func TestStreamContentTypeReturnsCloseFailure(t *testing.T) {
	closeErr := errors.New("close failed")
	body := &controlledBody{reader: strings.NewReader("ignored"), closeErr: closeErr}
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		response := response(http.StatusOK, body)
		response.Header.Set("Content-Type", "application/json")
		return response, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

	_, err := client.Watch()(context.Background(), nil)

	require.ErrorContains(t, err, "unexpected content type")
	require.ErrorIs(t, err, closeErr)
	assertDecodingError(t, err)
	require.Equal(t, 1, body.closeCalls)
}

func TestStreamContentTypeRemainsPlainWhenCloseSucceeds(t *testing.T) {
	body := &controlledBody{reader: strings.NewReader("ignored")}
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		response := response(http.StatusOK, body)
		response.Header.Set("Content-Type", "application/json")
		return response, nil
	})
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, goahttp.ResponseDecoder, false)

	_, err := client.Watch()(context.Background(), nil)

	require.EqualError(t, err, "unexpected content type: application/json (expected text/event-stream)")
	var clientErr *goahttp.ClientError
	require.NotErrorAs(t, err, &clientErr)
	require.Equal(t, 1, body.closeCalls)
}

func TestRawEndpointReturnsDecoderAndCloseFailures(t *testing.T) {
	decodeErr := errors.New("decode failed")
	closeErr := errors.New("close failed")
	body := &controlledBody{reader: strings.NewReader("ignored"), closeErr: closeErr}
	doer := doerFunc(func(*http.Request) (*http.Response, error) {
		return response(http.StatusBadRequest, body), nil
	})
	decoder := func(*http.Response) goahttp.Decoder {
		return goahttp.EncodingFunc(func(any) error {
			return decodeErr
		})
	}
	client := NewClient("http", "example.test", doer, goahttp.RequestEncoder, decoder, false)

	_, err := client.Raw()(context.Background(), nil)

	require.ErrorIs(t, err, decodeErr)
	require.ErrorIs(t, err, closeErr)
	assertDecodingError(t, err)
	require.Equal(t, 1, body.closeCalls)
}

func response(status int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: body}
}

func assertDecodingError(t *testing.T, err error) {
	t.Helper()
	var clientErr *goahttp.ClientError
	require.ErrorAs(t, err, &clientErr)
	require.Equal(t, "decoding_error", clientErr.Name)
}
`
