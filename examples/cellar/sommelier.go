package service

import (
	"context"
	"log"

	"goa.design/goa.v2/examples/cellar/gen/sommelier"
)

// sommelier service implementation
type somsvc struct {
	logger *log.Logger
}

// NewSommelier returns the sommelier service implementation.
func NewSommelier(logger *log.Logger) sommelier.Service {
	// Build and return service implementation.
	return &somsvc{logger}
}

// Pick bottle.
func (s *somsvc) Pick(ctx context.Context, c *sommelier.Criteria) (sommelier.StoredBottleCollection, error) {
	return nil, nil
}
