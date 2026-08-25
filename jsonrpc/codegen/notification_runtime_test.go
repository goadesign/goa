// This file renders a small JSON-RPC service into a temporary Go module.
// The generated client and server tests send real HTTP requests and check how
// notifications and request IDs control payloads, dispatch, and responses.
package codegen_test

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
	httpcodegen "goa.design/goa/v3/http/codegen"
	jsonrpccodegen "goa.design/goa/v3/jsonrpc/codegen"
)

// TestGeneratedNotificationAndRequestIDs renders and runs the generated
// client and server so the test covers complete request and response paths.
func TestGeneratedNotificationAndRequestIDs(t *testing.T) {
	dir := renderNotificationRuntimeModule(t)
	serverTestDir := filepath.Join(dir, "jsonrpc", "request_contract", "server")
	require.NoError(t, os.MkdirAll(serverTestDir, 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(serverTestDir, "notification_runtime_test.go"),
		[]byte(notificationRuntimeTest),
		0o600,
	))
	clientTestDir := filepath.Join(dir, "jsonrpc", "request_contract", "client")
	require.NoError(t, os.MkdirAll(clientTestDir, 0o750))
	require.NoError(t, os.WriteFile(
		filepath.Join(clientTestDir, "notification_runtime_test.go"),
		[]byte(notificationClientRuntimeTest),
		0o600,
	))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		"go",
		"test",
		"-mod=mod",
		"./jsonrpc/request_contract/client",
		"./jsonrpc/request_contract/server",
	)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
}

// renderNotificationRuntimeModule writes generated service, client, and server
// packages without changing any generated directory in this repository.
func renderNotificationRuntimeModule(t *testing.T) string {
	t.Helper()
	root := expr.RunDSL(t, notificationRuntimeDSL)
	generation, err := goacodegen.NewGeneration("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	servicePlan, err := service.NewPlan(
		root,
		generation,
		expr.NewExampleGenerator(root.API.RandomizerFactory),
	)
	require.NoError(t, err)
	httpPlans, err := httpcodegen.NewJSONRPCPlans(generation, httpcodegen.PlanInput{
		Root:    root,
		Service: servicePlan,
	})
	require.NoError(t, err)
	jsonPlans, err := jsonrpccodegen.NewPlans(generation, jsonrpccodegen.PlanInput{
		Root:    root,
		Service: servicePlan,
		HTTP:    httpPlans[0],
	})
	require.NoError(t, err)
	require.NoError(t, generation.Freeze())
	require.NoError(t, servicePlan.Link())
	require.NoError(t, httpPlans[0].Link())
	require.NoError(t, jsonPlans[0].Link())

	files, err := service.Files(servicePlan)
	require.NoError(t, err)
	files = append(files, jsonPlans[0].ClientFiles()...)
	files = append(files, jsonPlans[0].ServerFiles()...)
	files = append(files, jsonPlans[0].ClientTypeFiles()...)
	files = append(files, jsonPlans[0].ServerTypeFiles()...)
	files = append(files, jsonPlans[0].PathFiles()...)
	base := t.TempDir()
	for _, file := range files {
		_, err := file.Render(base)
		require.NoError(t, err)
	}

	moduleDir := filepath.Join(base, goacodegen.Gendir)
	workingDir, err := os.Getwd()
	require.NoError(t, err)
	repository := filepath.Clean(filepath.Join(workingDir, "..", ".."))
	module := fmt.Sprintf(
		"module generated.local/gen\n\ngo 1.25\n\nrequire goa.design/goa/v3 v3.0.0\n\nreplace goa.design/goa/v3 => %s\n",
		repository,
	)
	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(module), 0o600))
	return moduleDir
}

// notificationRuntimeDSL declares the request shapes exercised by the
// generated server test.
func notificationRuntimeDSL() {
	requestID := dsl.Type("RequestID", dsl.String)
	validatedRequestID := dsl.Type("ValidatedRequestID", dsl.String, func() {
		dsl.Pattern(`^req-[0-9]+$`)
	})
	dsl.Service("Request Contract", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("notify", func() {
			dsl.Payload(func() {
				dsl.Attribute("message", dsl.String)
				dsl.Required("message")
			})
			dsl.JSONRPC(func() {
				dsl.Notification()
			})
		})
		dsl.Method("ping", func() {
			dsl.JSONRPC(func() {})
		})
		dsl.Method("required_id", func() {
			dsl.Payload(func() {
				dsl.ID("id", requestID)
				dsl.Attribute("message", dsl.String)
				dsl.Required("id", "message")
			})
			dsl.JSONRPC(func() {})
		})
		dsl.Method("optional_id", func() {
			dsl.Payload(func() {
				dsl.ID("id", requestID)
			})
			dsl.JSONRPC(func() {})
		})
		dsl.Method("defaulted_id", func() {
			dsl.Payload(func() {
				dsl.ID("id", requestID, func() {
					dsl.Default("default-id")
				})
				dsl.Required("id")
			})
			dsl.JSONRPC(func() {})
		})
		dsl.Method("validated_id", func() {
			dsl.Payload(func() {
				dsl.ID("id", validatedRequestID)
				dsl.Required("id")
			})
			dsl.JSONRPC(func() {})
		})
	})
}

// notificationClientRuntimeTest checks the generated notification endpoint
// against controlled HTTP acknowledgements.
const notificationClientRuntimeTest = `package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/request_contract"
	goahttp "goa.design/goa/v3/http"
)

type doerFunc func(*http.Request) (*http.Response, error)

type notificationResponseBody struct {
	reader io.Reader
	reads  int
	closes int
}

type notificationEnvelope struct {
	JSONRPC string          ` + "`json:\"jsonrpc\"`" + `
	Method  string          ` + "`json:\"method\"`" + `
	ID      json.RawMessage ` + "`json:\"id\"`" + `
	Params  struct {
		Message string          ` + "`json:\"message\"`" + `
		ID      json.RawMessage ` + "`json:\"id\"`" + `
	} ` + "`json:\"params\"`" + `
}

func (do doerFunc) Do(request *http.Request) (*http.Response, error) {
	return do(request)
}

func (body *notificationResponseBody) Read(buffer []byte) (int, error) {
	body.reads++
	return body.reader.Read(buffer)
}

func (body *notificationResponseBody) Close() error {
	body.closes++
	return nil
}

func TestNotificationClientAcceptsSuccessfulAcknowledgementWithoutDecoding(t *testing.T) {
	body := &notificationResponseBody{reader: strings.NewReader("not JSON-RPC")}
	decoderCalls := 0
	client := NewClient(
		"http",
		"example.test",
		doerFunc(func(request *http.Request) (*http.Response, error) {
			assertNotificationRequest(t, request)
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
				Body:       body,
			}, nil
		}),
		goahttp.RequestEncoder,
		func(response *http.Response) goahttp.Decoder {
			decoderCalls++
			return goahttp.ResponseDecoder(response)
		},
		false,
	)

	_, err := client.Notify()(context.Background(), &service.NotifyPayload{Message: "ready"})

	require.NoError(t, err)
	require.Zero(t, decoderCalls)
	require.Zero(t, body.reads)
	require.Equal(t, 1, body.closes)
}

func TestNotificationClientReturnsNonSuccessAcknowledgementError(t *testing.T) {
	body := &notificationResponseBody{reader: strings.NewReader("denied")}
	decoderCalls := 0
	client := NewClient(
		"http",
		"example.test",
		doerFunc(func(request *http.Request) (*http.Response, error) {
			assertNotificationRequest(t, request)
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     make(http.Header),
				Body:       body,
			}, nil
		}),
		goahttp.RequestEncoder,
		func(response *http.Response) goahttp.Decoder {
			decoderCalls++
			return goahttp.ResponseDecoder(response)
		},
		false,
	)

	_, err := client.Notify()(context.Background(), &service.NotifyPayload{Message: "ready"})

	var clientErr *goahttp.ClientError
	require.ErrorAs(t, err, &clientErr)
	require.Equal(t, "invalid_response", clientErr.Name)
	require.Equal(t, "Request Contract", clientErr.Service)
	require.Equal(t, "notify", clientErr.Method)
	require.ErrorContains(t, err, "invalid response code 502, body: denied")
	require.Zero(t, decoderCalls)
	require.Greater(t, body.reads, 0)
	require.Equal(t, 1, body.closes)
}

// assertNotificationRequest checks the exact envelope written by the client.
func assertNotificationRequest(t *testing.T, request *http.Request) {
	t.Helper()
	var envelope notificationEnvelope
	require.NoError(t, json.NewDecoder(request.Body).Decode(&envelope))
	require.Equal(t, "2.0", envelope.JSONRPC)
	require.Equal(t, "notify", envelope.Method)
	require.Nil(t, envelope.ID)
	require.Equal(t, "ready", envelope.Params.Message)
	require.Nil(t, envelope.Params.ID)
}
`

const notificationRuntimeTest = `package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	service "generated.local/gen/request_contract"
	goahttp "goa.design/goa/v3/http"
)

type requestContractService struct {
	notifyCalls     int
	notifyMessage   string
	pingCalls       int
	requiredIDCalls int
	requiredID      service.RequestID
	optionalIDCalls int
	optionalID      *service.RequestID
	defaultedIDCalls int
	defaultedID      service.RequestID
	validatedIDCalls int
	validatedID      service.ValidatedRequestID
}

func (svc *requestContractService) Notify(_ context.Context, payload *service.NotifyPayload) error {
	svc.notifyCalls++
	svc.notifyMessage = payload.Message
	if payload.Message == "fail" {
		return errors.New("notification failed")
	}
	return nil
}

func (svc *requestContractService) Ping(context.Context) error {
	svc.pingCalls++
	return nil
}

func (svc *requestContractService) RequiredID(_ context.Context, payload *service.RequiredIDPayload) error {
	svc.requiredIDCalls++
	svc.requiredID = payload.ID
	return nil
}

func (svc *requestContractService) OptionalID(_ context.Context, payload *service.OptionalIDPayload) error {
	svc.optionalIDCalls++
	svc.optionalID = payload.ID
	return nil
}

func (svc *requestContractService) DefaultedID(_ context.Context, payload *service.DefaultedIDPayload) error {
	svc.defaultedIDCalls++
	svc.defaultedID = payload.ID
	return nil
}

func (svc *requestContractService) ValidatedID(_ context.Context, payload *service.ValidatedIDPayload) error {
	svc.validatedIDCalls++
	svc.validatedID = payload.ID
	return nil
}

func TestNotificationDispatchesPayloadWithoutResponse(t *testing.T) {
	svc := &requestContractService{}
	response, reported := serveRequestContract(svc, ` + "`" + `{"jsonrpc":"2.0","method":"notify","params":{"message":"ready"}}` + "`" + `)
	require.Empty(t, reported)
	require.Empty(t, response.Body.String())
	require.Equal(t, 1, svc.notifyCalls)
	require.Equal(t, "ready", svc.notifyMessage)
}

func TestNotificationErrorsCannotWriteAResponse(t *testing.T) {
	tests := []struct {
		name    string
		request string
	}{
		{name: "invalid parameters", request: ` + "`" + `{"jsonrpc":"2.0","method":"notify","params":{}}` + "`" + `},
		{name: "service error", request: ` + "`" + `{"jsonrpc":"2.0","method":"notify","params":{"message":"fail"}}` + "`" + `},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := &requestContractService{}
			response, reported := serveRequestContract(svc, test.request)
			require.Len(t, reported, 1)
			require.Error(t, reported[0])
			require.Empty(t, response.Body.String())
		})
	}
}

func TestOrdinaryVoidRequestReturnsNullWithMatchingID(t *testing.T) {
	svc := &requestContractService{}
	response, reported := serveRequestContract(svc, ` + "`" + `{"jsonrpc":"2.0","id":"request-1","method":"ping"}` + "`" + `)
	require.Empty(t, reported)
	require.JSONEq(t, ` + "`" + `{"jsonrpc":"2.0","id":"request-1","result":null}` + "`" + `, response.Body.String())
	require.Equal(t, 1, svc.pingCalls)
}

func TestMissingAndEmptyMethodNamesReturnDifferentProtocolErrors(t *testing.T) {
	tests := []struct {
		name    string
		request string
		code    string
	}{
		{name: "missing", request: ` + "`" + `{"jsonrpc":"2.0","id":"request-1"}` + "`" + `, code: ` + "`" + `"code":-32600` + "`" + `},
		{name: "empty", request: ` + "`" + `{"jsonrpc":"2.0","id":"request-1","method":""}` + "`" + `, code: ` + "`" + `"code":-32601` + "`" + `},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := &requestContractService{}
			response, reported := serveRequestContract(svc, test.request)
			require.Empty(t, reported)
			require.Contains(t, response.Body.String(), test.code)
			require.Zero(t, svc.pingCalls)
		})
	}
}

func TestRequiredIDComesOnlyFromEnvelope(t *testing.T) {
	svc := &requestContractService{}
	response, reported := serveRequestContract(svc, ` + "`" + `{"jsonrpc":"2.0","id":"envelope","method":"required_id","params":{"id":"params","message":"ready"}}` + "`" + `)
	require.Empty(t, reported)
	require.JSONEq(t, ` + "`" + `{"jsonrpc":"2.0","id":"envelope","result":null}` + "`" + `, response.Body.String())
	require.Equal(t, 1, svc.requiredIDCalls)
	require.Equal(t, service.RequestID("envelope"), svc.requiredID)
}

func TestRequiredNumericIDKeepsItsExactDigits(t *testing.T) {
	svc := &requestContractService{}
	response, reported := serveRequestContract(svc, ` + "`" + `{"jsonrpc":"2.0","id":9007199254740993123456789,"method":"required_id","params":{"message":"ready"}}` + "`" + `)
	require.Empty(t, reported)
	require.Contains(t, response.Body.String(), ` + "`" + `"id":9007199254740993123456789` + "`" + `)
	require.Equal(t, 1, svc.requiredIDCalls)
	require.Equal(t, service.RequestID("9007199254740993123456789"), svc.requiredID)
}

func TestRequiredIDRejectsMissingAndNullEnvelopeIDsBeforeDispatch(t *testing.T) {
	tests := []struct {
		name     string
		request  string
		wantBody bool
	}{
		{name: "missing", request: ` + "`" + `{"jsonrpc":"2.0","method":"required_id","params":{"id":"params","message":"ready"}}` + "`" + `},
		{name: "null", request: ` + "`" + `{"jsonrpc":"2.0","id":null,"method":"required_id","params":{"id":"params","message":"ready"}}` + "`" + `, wantBody: true},
	}
	for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				svc := &requestContractService{}
				response, reported := serveRequestContract(svc, test.request)
				require.Zero(t, svc.requiredIDCalls)
				if test.wantBody {
					require.Empty(t, reported)
					require.Contains(t, response.Body.String(), ` + "`" + `"code":-32602` + "`" + `)
					require.Contains(t, response.Body.String(), ` + "`" + `\"id\" is missing from JSON-RPC request` + "`" + `)
					return
				}
				require.Len(t, reported, 1)
				require.ErrorContains(t, reported[0], ` + "`" + `"id" is missing from JSON-RPC request` + "`" + `)
				require.Empty(t, response.Body.String())
			})
		}
}

func TestOptionalMissingIDDispatchesNotificationWithNilPayloadID(t *testing.T) {
	svc := &requestContractService{}
	response, reported := serveRequestContract(svc, ` + "`" + `{"jsonrpc":"2.0","method":"optional_id","params":{"id":"params"}}` + "`" + `)
	require.Empty(t, reported)
	require.Empty(t, response.Body.String())
	require.Equal(t, 1, svc.optionalIDCalls)
	require.Nil(t, svc.optionalID)
}

func TestRequiredDefaultedMissingIDDispatchesNotificationWithAuthoredPayloadID(t *testing.T) {
	svc := &requestContractService{}
	response, reported := serveRequestContract(svc, ` + "`" + `{"jsonrpc":"2.0","method":"defaulted_id"}` + "`" + `)
	require.Empty(t, reported)
	require.Empty(t, response.Body.String())
	require.Equal(t, 1, svc.defaultedIDCalls)
	require.Equal(t, service.RequestID("default-id"), svc.defaultedID)
}

func TestValidatedIDUsesAuthoredValidationBeforeDispatch(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantCalls int
		wantError bool
	}{
		{name: "valid", id: "req-7", wantCalls: 1},
		{name: "invalid", id: "other", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc := &requestContractService{}
			request := ` + "`" + `{"jsonrpc":"2.0","id":"` + "`" + ` + test.id + ` + "`" + `","method":"validated_id"}` + "`" + `
			response, reported := serveRequestContract(svc, request)
			require.Empty(t, reported)
			require.Equal(t, test.wantCalls, svc.validatedIDCalls)
			if test.wantError {
				require.Contains(t, response.Body.String(), ` + "`" + `"code":-32602` + "`" + `)
				require.Contains(t, response.Body.String(), "must match the regexp")
				return
			}
			require.Equal(t, service.ValidatedRequestID(test.id), svc.validatedID)
			require.Contains(t, response.Body.String(), ` + "`" + `"result":null` + "`" + `)
		})
	}
}

func serveRequestContract(svc *requestContractService, body string) (*httptest.ResponseRecorder, []error) {
	reported := make([]error, 0, 1)
	server := New(
		service.NewEndpoints(svc),
		goahttp.NewMuxer(),
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(_ context.Context, writer http.ResponseWriter, err error) {
			reported = append(reported, err)
			if _, writeErr := writer.Write([]byte("error handler wrote a response")); writeErr != nil {
				reported = append(reported, writeErr)
			}
		},
	)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/rpc", strings.NewReader(body))
	server.ServeHTTP(response, request)
	return response, reported
}
`
