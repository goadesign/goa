// Code generated with goagen v2.0.0-wip, DO NOT EDIT.
//
// accountEndpoints
//
// Command:
// $ goagen server goa.design/goa.v2/examples/account/design

package endpoints

import (
	"context"

	"goa.design/goa.v2"
	"goa.design/goa.v2/examples/account/gen/service"
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
func NewAccount(s service.Account) *Account {
	ep := new(Account)

	ep.Create = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*service.CreateAccount)
		return s.Create(ctx, p)
	}

	ep.List = func(ctx context.Context, req interface{}) (interface{}, error) {
		var p *service.ListAccount
		if req != nil {
			p = req.(*service.ListAccount)
		}
		return s.List(ctx, p)
	}

	ep.Show = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*service.ShowAccountPayload)
		return s.Show(ctx, p)
	}

	ep.Delete = func(ctx context.Context, req interface{}) (interface{}, error) {
		p := req.(*service.DeleteAccountPayload)
		return nil, s.Delete(ctx, p)
	}

	return ep
}
