package endpoints

import (
	"golang.org/x/net/context"

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

// NewAccount wraps the given account service with endpoints.
func NewAccount(s services.Account) *Account {
	ep := &Account{}

	ep.Create = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*services.CreateAccountPayload)
		return s.Create(ctx, p)
	}

	ep.List = func(ctx context.Context, req interface{}) (interface{}, error) {
		var p *services.ListAccountPayload
		if req != nil {
			p = req.(*services.ListAccountPayload)
		}
		return s.List(ctx, p)
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
