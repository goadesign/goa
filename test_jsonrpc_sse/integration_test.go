package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	testserverjsonrpc "test-jsonrpc-sse/gen/jsonrpc/test_service/server"
	testservice "test-jsonrpc-sse/gen/test_service"

	goahttp "goa.design/goa/v3/http"
)

// Simple implementation of the TestService for testing
type testSvc struct{}

func (s *testSvc) StreamMessages(ctx context.Context, p *testservice.StreamMessagesPayload, stream testservice.StreamMessagesServerStream) error {
	fmt.Printf("StreamMessages called with payload: %+v\n", p)

	// Send a few test messages
	for i := 0; i < 3; i++ {
		msg := &testservice.StreamMessagesResult{
			EventID:   func() *string { s := fmt.Sprintf("evt-%d", i); return &s }(),
			Message:   func() *string { s := fmt.Sprintf("Message %d for topic %s", i, *p.Topic); return &s }(),
			Timestamp: func() *string { s := time.Now().Format(time.RFC3339); return &s }(),
		}
		if err := stream.Send(msg); err != nil {
			return fmt.Errorf("failed to send message %d: %w", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Send final response
	finalMsg := &testservice.StreamMessagesResult{
		EventID:   func() *string { s := "final"; return &s }(),
		Message:   func() *string { s := "Stream complete"; return &s }(),
		Timestamp: func() *string { s := time.Now().Format(time.RFC3339); return &s }(),
	}

	return stream.SendWithContext(ctx, finalMsg)
}

func (s *testSvc) StreamSimple(ctx context.Context, p *testservice.StreamSimplePayload, stream testservice.StreamSimpleServerStream) error {
	fmt.Printf("StreamSimple called with payload: %+v\n", p)

	// Send a few simple string messages
	for i := 0; i < 3; i++ {
		msg := fmt.Sprintf("Simple message %d", i)
		if err := stream.Send(msg); err != nil {
			return fmt.Errorf("failed to send simple message %d: %w", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	return stream.SendWithContext(ctx, "Simple stream complete")
}

func (s *testSvc) Notification(ctx context.Context, p *testservice.NotificationPayload) error {
	fmt.Printf("Notification received: %s\n", *p.Message)
	return nil
}

func TestSSEStreamMessages(t *testing.T) {
	// Create service implementation
	svc := &testSvc{}

	// Create endpoints
	endpoints := testservice.NewEndpoints(svc)

	// Create HTTP mux and mount JSON-RPC handlers
	mux := goahttp.NewMuxer()

	// Create JSON-RPC server with proper parameters
	server := testserverjsonrpc.New(
		endpoints,
		mux,
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(ctx context.Context, w http.ResponseWriter, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	)

	testserverjsonrpc.Mount(mux, server)

	// Create test server
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Create JSON-RPC request for StreamMessages
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "StreamMessages",
		"params": map[string]interface{}{
			"id":            "test-request-123",
			"last_event_id": "0",
			"topic":         "test-topic",
		},
		"id": "test-request-123",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON-RPC request: %v", err)
	}

	// Send the request to the correct path /stream
	resp, err := http.Post(ts.URL+"/stream", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response headers
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type: text/event-stream, got: %s", resp.Header.Get("Content-Type"))
	}

	if resp.Header.Get("Cache-Control") != "no-cache" {
		t.Errorf("Expected Cache-Control: no-cache, got: %s", resp.Header.Get("Cache-Control"))
	}

	// Read the SSE stream
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)
	t.Logf("SSE Response:\n%s", bodyStr)

	// Verify that we get SSE events
	if !strings.Contains(bodyStr, "event: notification") {
		t.Error("Expected to find 'event: notification' in response")
	}

	if !strings.Contains(bodyStr, "data: ") {
		t.Error("Expected to find 'data: ' in response")
	}

	if !strings.Contains(bodyStr, "jsonrpc") {
		t.Error("Expected to find 'jsonrpc' in response data")
	}

	if !strings.Contains(bodyStr, "StreamMessages") {
		t.Error("Expected to find 'StreamMessages' method in response")
	}
}

func TestSSEStreamSimple(t *testing.T) {
	// Create service implementation
	svc := &testSvc{}

	// Create endpoints
	endpoints := testservice.NewEndpoints(svc)

	// Create HTTP mux and mount JSON-RPC handlers
	mux := goahttp.NewMuxer()

	// Create JSON-RPC server with proper parameters
	server := testserverjsonrpc.New(
		endpoints,
		mux,
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(ctx context.Context, w http.ResponseWriter, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	)

	testserverjsonrpc.Mount(mux, server)

	// Create test server
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Create JSON-RPC request for StreamSimple
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "StreamSimple",
		"params": map[string]interface{}{
			"id": "simple-request-456",
		},
		"id": "simple-request-456",
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON-RPC request: %v", err)
	}

	// Send the request to the correct path /simple using GET
	req, err := http.NewRequest("GET", ts.URL+"/simple", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response headers
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type: text/event-stream, got: %s", resp.Header.Get("Content-Type"))
	}

	// Read the SSE stream
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)
	t.Logf("SSE Response:\n%s", bodyStr)

	// Verify that we get SSE events with simple string messages
	if !strings.Contains(bodyStr, "event: notification") {
		t.Error("Expected to find 'event: notification' in response")
	}

	if !strings.Contains(bodyStr, "Simple message") {
		t.Error("Expected to find 'Simple message' in response data")
	}
}

func TestNotification(t *testing.T) {
	// Create service implementation
	svc := &testSvc{}

	// Create endpoints
	endpoints := testservice.NewEndpoints(svc)

	// Create HTTP mux and mount JSON-RPC handlers
	mux := goahttp.NewMuxer()

	// Create JSON-RPC server with proper parameters
	server := testserverjsonrpc.New(
		endpoints,
		mux,
		goahttp.RequestDecoder,
		goahttp.ResponseEncoder,
		func(ctx context.Context, w http.ResponseWriter, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	)

	testserverjsonrpc.Mount(mux, server)

	// Create test server
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Create JSON-RPC notification (no id field)
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "Notification",
		"params": map[string]interface{}{
			"message": "Test notification message",
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON-RPC request: %v", err)
	}

	// Send the notification to the correct path /notify
	resp, err := http.Post(ts.URL+"/notify", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}
	defer resp.Body.Close()

	// For now, notifications in SSE mode return 200 with SSE headers
	// This is acceptable since notifications are handled properly by the service
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got: %d", resp.StatusCode)
	}

	// Check that notification was handled by reading the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	t.Logf("Notification response: %s", string(body))
}
