package xray

import (
	"context"
	"encoding/json"
	"fmt"

	"goa.design/goa/grpc/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewClient returns a client-side middleware that creates AWS X-Ray
// sub-segments in the outgoing gRPC request metadata for traced requests.
func NewClient(host string) grpc.UnaryClientInterceptor {
	return grpc.UnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.MD{}
		}

		var seg *Segment
		{
			s := middleware.MetadataValue(md, SegmentMetadataKey)
			if s != "" {
				if err := json.Unmarshal([]byte(s), seg); err != nil {
					return fmt.Errorf("error unmarshalling segment from metadata: %v", err)
				}
			}
		}
		if seg == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		sub := seg.NewSubsegment(host)
		defer sub.Close()

		// update the context with the latest segment
		md = middleware.WithSpan(md, sub.TraceID, sub.ID, sub.ParentID)
		if b, err := json.Marshal(sub); err == nil {
			md.Set(SegmentMetadataKey, string(b))
		}
		ctx = metadata.NewOutgoingContext(ctx, md)

		sub.RecordRequest(ctx, method, "remote")
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			sub.RecordError(err)
		} else {
			sub.RecordResponse()
		}
		return err
	})
}
