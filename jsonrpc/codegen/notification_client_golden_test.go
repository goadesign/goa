// This file checks that generated JSON-RPC clients preserve methods declared
// as notifications by sending no request ID.
package codegen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	goacodegen "goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/testutil"
	"goa.design/goa/v3/dsl"
)

// TestNotificationClientGeneratedSource checks notifications with and without
// parameters beside an ordinary method that returns JSON null.
func TestNotificationClientGeneratedSource(t *testing.T) {
	_, plan := linkedJSONRPCPlan(t, notificationClientDSL)
	var encoderSource string
	for _, file := range plan.ClientFiles() {
		if filepath.Base(file.Path) == "encode_decode.go" {
			encoderSource = "package client\n\n" + goacodegen.SectionsCode(t, file.Section("jsonrpc-request-encoder"))
			break
		}
	}
	require.NotEmpty(t, encoderSource)
	require.Contains(t, encoderSource, `Method:  "with_payload"`)
	require.Contains(t, encoderSource, `Method:  "without_payload"`)
	require.Contains(t, encoderSource, `Method:  "ordinary_void"`)
	require.Equal(t, 2, strings.Count(encoderSource, "uuid.New"))
	require.NotContains(t, encoderSource, "ID: p.ID")
	testutil.AssertGo(t, "testdata/golden/notification_client.go.golden", encoderSource)

	var endpointSource string
	for _, file := range plan.ClientFiles() {
		if filepath.Base(file.Path) == "client.go" {
			endpointSource = "package client\n\n" + goacodegen.SectionsCode(t, file.Section("jsonrpc-client-endpoint-init"))
			break
		}
	}
	require.NotEmpty(t, endpointSource)
	require.Equal(t, 5, strings.Count(endpointSource, "decodeResponse ="))
	require.Equal(t, 4, strings.Count(endpointSource, "resp.Body.Close()"))
	testutil.AssertGo(t, "testdata/golden/notification_endpoint.go.golden", endpointSource)

	var serverSource string
	for _, file := range plan.ServerFiles() {
		if filepath.Base(file.Path) == "server.go" {
			serverSource = "package server\n\n" + goacodegen.SectionsCode(t, file.Section("jsonrpc-server-handler-init"))
			break
		}
	}
	require.NotEmpty(t, serverSource)
	require.NotContains(t, serverSource, "MakeSuccessResponse(id,")
	require.Contains(t, serverSource, "MakeSuccessResponse(req.ID,")
	testutil.AssertGo(t, "testdata/golden/notification_server.go.golden", serverSource)

	var decoderSource string
	for _, file := range plan.ServerFiles() {
		if filepath.Base(file.Path) == "encode_decode.go" {
			decoderSource = "package server\n\n" + goacodegen.SectionsCode(t, file.Section("jsonrpc-request-decoder"))
			break
		}
	}
	require.NotEmpty(t, decoderSource)
	require.Contains(t, decoderSource, "if !req.HasID || req.ID == nil")
	require.Contains(t, decoderSource, "jsonrpc.IDToString(req.ID)")
	testutil.AssertGo(t, "testdata/golden/notification_decoder.go.golden", decoderSource)
}

// notificationClientDSL declares two methods that receive no JSON-RPC
// response and therefore must never send a request ID.
func notificationClientDSL() {
	requestID := dsl.Type("RequestID", dsl.String)
	validatedRequestID := dsl.Type("ValidatedRequestID", dsl.String, func() {
		dsl.Pattern(`^req-[0-9]+$`)
	})
	dsl.Service("notifications", func() {
		dsl.JSONRPC(func() {
			dsl.POST("/rpc")
		})
		dsl.Method("with_payload", func() {
			dsl.Payload(func() {
				dsl.Attribute("message", dsl.String)
				dsl.Required("message")
			})
			dsl.JSONRPC(func() {
				dsl.Notification()
			})
		})
		dsl.Method("without_payload", func() {
			dsl.JSONRPC(func() {
				dsl.Notification()
			})
		})
		dsl.Method("ordinary_void", func() {
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
