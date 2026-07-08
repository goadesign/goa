package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestUsesStreamEnvelope(t *testing.T) {
	cases := []struct {
		Name     string
		Metadata metadata.MD
		Expected bool
	}{
		{"no incoming metadata", nil, false},
		{"no protocol declared", metadata.Pairs("foo", "bar"), false},
		{"envelope protocol declared", metadata.Pairs(StreamProtocolMetadataKey, StreamProtocolEnvelope), true},
		{"unknown protocol declared", metadata.Pairs(StreamProtocolMetadataKey, "0"), false},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ctx := context.Background()
			if c.Metadata != nil {
				ctx = metadata.NewIncomingContext(ctx, c.Metadata)
			}
			assert.Equal(t, c.Expected, UsesStreamEnvelope(ctx))
		})
	}
}
