package middleware_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	grpcm "goa.design/goa/v3/grpc/middleware"
	"goa.design/goa/v3/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	traceID    = "testTraceID"
	spanID     = "testSpanID"
	newTraceID = func() string { return traceID }
	newID      = func() string { return spanID }
	discard    = regexp.MustCompile("Test$")
)

func TestUnaryServerTrace(t *testing.T) {
	var (
		unary = &grpc.UnaryServerInfo{FullMethod: "Test.Test"}
	)

	cases := map[string]struct {
		Rate                  int
		TraceID, ParentSpanID string
		Discard               *regexp.Regexp
		// output
		CtxTraceID, CtxSpanID, CtxParentID string
	}{
		"no-trace":             {100, "", "", nil, traceID, spanID, ""},
		"no-trace-discarded":   {100, "", "", discard, "", "", ""},
		"trace":                {100, "trace", "", nil, "trace", spanID, ""},
		"trace-not-discarded":  {100, "trace", "", discard, "trace", spanID, ""},
		"parent":               {100, "trace", "parent", nil, "trace", spanID, "parent"},
		"parent-not-discarded": {100, "trace", "parent", discard, "trace", spanID, "parent"},

		"zero-rate-no-trace":             {0, "", "", nil, "", "", ""},
		"zero-rate-trace":                {0, "trace", "", nil, "trace", spanID, ""},
		"zero-rate-trace-not-discarded":  {0, "trace", "", discard, "trace", spanID, ""},
		"zero-rate-parent":               {0, "trace", "parent", nil, "trace", spanID, "parent"},
		"zero-rate-parent-not-discarded": {0, "trace", "parent", discard, "trace", spanID, "parent"},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				var ctxTraceID, ctxSpanID, ctxParentID string
				{
					if traceID := ctx.Value(middleware.TraceIDKey); traceID != nil {
						ctxTraceID = traceID.(string)
					}
					if spanID := ctx.Value(middleware.TraceSpanIDKey); spanID != nil {
						ctxSpanID = spanID.(string)
					}
					if parentID := ctx.Value(middleware.TraceParentSpanIDKey); parentID != nil {
						ctxParentID = parentID.(string)
					}
				}
				if ctxTraceID != c.CtxTraceID {
					return nil, fmt.Errorf("invalid TraceID, expected %v - got %v", c.CtxTraceID, ctxTraceID)
				}
				if ctxSpanID != c.CtxSpanID {
					return nil, fmt.Errorf("invalid SpanID, expected %v - got %v", c.CtxSpanID, ctxSpanID)
				}
				if ctxParentID != c.CtxParentID {
					return nil, fmt.Errorf("invalid ParentSpanID, expected %v - got %v", c.CtxParentID, ctxParentID)
				}
				return "response", nil
			}

			md := metadata.MD{}
			if c.TraceID != "" {
				md.Set(grpcm.TraceIDMetadataKey, c.TraceID)
			}
			if c.ParentSpanID != "" {
				md.Set(grpcm.ParentSpanIDMetadataKey, c.ParentSpanID)
			}
			ctx := metadata.NewIncomingContext(context.Background(), md)
			traceOptions := []middleware.TraceOption{
				grpcm.SamplingPercent(c.Rate),
				grpcm.SpanIDFunc(newID),
				grpcm.TraceIDFunc(newTraceID),
			}
			if c.Discard != nil {
				traceOptions = append(traceOptions, grpcm.DiscardFromTrace(c.Discard))
			}
			_, err := grpcm.UnaryServerTrace(traceOptions...)(ctx, "request", unary, handler)
			if err != nil {
				t.Errorf("UnaryServerTrace error: %v", err)
			}
		})
	}
}

func TestStreamServerTrace(t *testing.T) {
	var (
		stream = &grpc.StreamServerInfo{FullMethod: "Test.Test"}
	)

	cases := map[string]struct {
		Rate                  int
		TraceID, ParentSpanID string
		Discard               *regexp.Regexp
		// output
		CtxTraceID, CtxSpanID, CtxParentID string
	}{
		"no-trace":             {100, "", "", nil, traceID, spanID, ""},
		"no-trace-discarded":   {100, "", "", discard, "", "", ""},
		"trace":                {100, "trace", "", nil, "trace", spanID, ""},
		"trace-not-discarded":  {100, "trace", "", discard, "trace", spanID, ""},
		"parent":               {100, "trace", "parent", nil, "trace", spanID, "parent"},
		"parent-not-discarded": {100, "trace", "parent", discard, "trace", spanID, "parent"},

		"zero-rate-no-trace":             {0, "", "", nil, "", "", ""},
		"zero-rate-trace":                {0, "trace", "", nil, "trace", spanID, ""},
		"zero-rate-trace-not-discarded":  {0, "trace", "", discard, "trace", spanID, ""},
		"zero-rate-parent":               {0, "trace", "parent", nil, "trace", spanID, "parent"},
		"zero-rate-parent-not-discarded": {0, "trace", "parent", discard, "trace", spanID, "parent"},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			handler := func(srv interface{}, stream grpc.ServerStream) error {
				ctx := stream.Context()
				var ctxTraceID, ctxSpanID, ctxParentID string
				{
					if traceID := ctx.Value(middleware.TraceIDKey); traceID != nil {
						ctxTraceID = traceID.(string)
					}
					if spanID := ctx.Value(middleware.TraceSpanIDKey); spanID != nil {
						ctxSpanID = spanID.(string)
					}
					if parentID := ctx.Value(middleware.TraceParentSpanIDKey); parentID != nil {
						ctxParentID = parentID.(string)
					}
				}
				if ctxTraceID != c.CtxTraceID {
					return fmt.Errorf("invalid TraceID, expected %v - got %v", c.CtxTraceID, ctxTraceID)
				}
				if ctxSpanID != c.CtxSpanID {
					return fmt.Errorf("invalid SpanID, expected %v - got %v", c.CtxSpanID, ctxSpanID)
				}
				if ctxParentID != c.CtxParentID {
					return fmt.Errorf("invalid ParentSpanID, expected %v - got %v", c.CtxParentID, ctxParentID)
				}
				return nil
			}

			md := metadata.MD{}
			if c.TraceID != "" {
				md.Set(grpcm.TraceIDMetadataKey, c.TraceID)
			}
			if c.ParentSpanID != "" {
				md.Set(grpcm.ParentSpanIDMetadataKey, c.ParentSpanID)
			}
			ctx := metadata.NewIncomingContext(context.Background(), md)
			wss := grpcm.NewWrappedServerStream(ctx, &testServerStream{})
			traceOptions := []middleware.TraceOption{
				grpcm.SamplingPercent(c.Rate),
				grpcm.SpanIDFunc(newID),
				grpcm.TraceIDFunc(newTraceID),
			}
			if c.Discard != nil {
				traceOptions = append(traceOptions, grpcm.DiscardFromTrace(c.Discard))
			}
			err := grpcm.StreamServerTrace(traceOptions...)(nil, wss, stream, handler)
			if err != nil {
				t.Errorf("StreamServerTrace error: %v", err)
			}
		})
	}
}

func TestUnaryClientTrace(t *testing.T) {
	cases := map[string]struct {
		TraceID, SpanID string
		Invoker         grpc.UnaryInvoker
	}{
		"no-trace": {"", "", func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.MD{}
			}
			if v := grpcm.MetadataValue(md, grpcm.TraceIDMetadataKey); v != "" {
				return fmt.Errorf("invalid TraceID, expected: \"\", got %q", v)
			}
			if v := grpcm.MetadataValue(md, grpcm.ParentSpanIDMetadataKey); v != "" {
				return fmt.Errorf("invalid TraceID, expected: \"\", got %q", v)
			}
			return nil
		}},
		"with-trace": {traceID, spanID, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.MD{}
			}
			if v := grpcm.MetadataValue(md, grpcm.TraceIDMetadataKey); v != traceID {
				return fmt.Errorf("invalid TraceID, expected: %q, got %q", traceID, v)
			}
			if v := grpcm.MetadataValue(md, grpcm.ParentSpanIDMetadataKey); v != spanID {
				return fmt.Errorf("invalid TraceID, expected: %q, got %q", spanID, v)
			}
			return nil
		}},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.Background()
			if c.TraceID != "" {
				ctx = context.WithValue(ctx, middleware.TraceIDKey, c.TraceID)
				ctx = context.WithValue(ctx, middleware.TraceSpanIDKey, c.SpanID)
			}
			if err := grpcm.UnaryClientTrace()(ctx, "Test.Test", nil, nil, nil, c.Invoker); err != nil {
				t.Errorf("UnaryClientTrace error: %v", err)
			}
		})
	}
}

func TestStreamClientTrace(t *testing.T) {
	cases := map[string]struct {
		TraceID, SpanID string
		Streamer        grpc.Streamer
	}{
		"no-trace": {"", "", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.MD{}
			}
			if v := grpcm.MetadataValue(md, grpcm.TraceIDMetadataKey); v != "" {
				return nil, fmt.Errorf("invalid TraceID, expected: \"\", got %q", v)
			}
			if v := grpcm.MetadataValue(md, grpcm.ParentSpanIDMetadataKey); v != "" {
				return nil, fmt.Errorf("invalid TraceID, expected: \"\", got %q", v)
			}
			return nil, nil
		}},
		"with-trace": {traceID, spanID, func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.MD{}
			}
			if v := grpcm.MetadataValue(md, grpcm.TraceIDMetadataKey); v != traceID {
				return nil, fmt.Errorf("invalid TraceID, expected: %q, got %q", traceID, v)
			}
			if v := grpcm.MetadataValue(md, grpcm.ParentSpanIDMetadataKey); v != spanID {
				return nil, fmt.Errorf("invalid TraceID, expected: %q, got %q", spanID, v)
			}
			return nil, nil
		}},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			ctx := context.Background()
			if c.TraceID != "" {
				ctx = context.WithValue(ctx, middleware.TraceIDKey, c.TraceID)
				ctx = context.WithValue(ctx, middleware.TraceSpanIDKey, c.SpanID)
			}
			if _, err := grpcm.StreamClientTrace()(ctx, nil, nil, "Test.Test", c.Streamer); err != nil {
				t.Errorf("UnaryClientTrace error: %v", err)
			}
		})
	}
}
