package app

import (
	"context"

	goa "goa.design/goa.v2"
)

type (
	// AccountEndpoints lists the account service endpoints.
	AccountEndpoints struct {
		Create goa.Endpoint
		List   goa.Endpoint
		Show   goa.Endpoint
		Delete goa.Endpoint
	}
)

func NewAccountEndpoints(s AccountService) *AccountEndpoints {
	ep := &AccountEndpoints{}

	ep.Create = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*CreateAccountPayload)
		return s.Create(ctx, p)
	}

	ep.List = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.List(ctx)
	}

	ep.Show = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*ShowAccountPayload)
		return s.Show(ctx, p)
	}

	ep.Delete = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*DeleteAccountPayload)
		return nil, s.Delete(ctx, p)
	}

	return ep
}
