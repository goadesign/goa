package goa

import "context"

// Endpoint exposes service methods to remote clients.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)
