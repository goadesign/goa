package endpoints

import (
	"context"

	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/services"
)

type (
	// Account lists the account service endpoints.
	Account struct {
		Create goa.Endpoint
		List   goa.Endpoint
		Show   goa.Endpoint
		Delete goa.Endpoint
	}
)

// NewAccount creates a new account service.
func NewAccount(s services.Account) *Account {
	ep := &Account{}

	ep.Create = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*services.CreateAccountPayload)
		return s.Create(ctx, p)
	}

	ep.List = func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.List(ctx)
	}

	ep.Show = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*services.ShowAccountPayload)
		return s.Show(ctx, p)
	}

	ep.Delete = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*services.DeleteAccountPayload)
		return nil, s.Delete(ctx, p)
	}

	return ep
}
