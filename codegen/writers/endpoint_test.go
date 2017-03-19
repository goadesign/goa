package writers

import (
	"bytes"
	"testing"

	"goa.design/goa.v2/design"
)

func TestEndpoint(t *testing.T) {
	const (
		singleMethod = `type (
	// Account lists the account service endpoints.
	Account struct {
		Create goa.Endpoint
	}
)

// NewAccount wraps the given account service with endpoints.
func NewAccount(s services.Account) *Account {
	ep := &Account{}

	ep.Create = func(ctx context.Context, req interface{}) (interface{}, error) {
		var p *services.CreateAccountPayload
		if req != nil {
			p = req.(*services.CreateAccountPayload)
		}
		return s.Create(ctx, p)
	}

	return ep
}`

		multipleMethods = `type (
	// Account lists the account service endpoints.
	Account struct {
		Create goa.Endpoint
		List goa.Endpoint
		Show goa.Endpoint
		Delete goa.Endpoint
	}
)

// NewAccount wraps the given account service with endpoints.
func NewAccount(s services.Account) *Account {
	ep := &Account{}

	ep.Create = func(ctx context.Context, req interface{}) (interface{}, error) {
		var p *services.CreateAccountPayload
		if req != nil {
			p = req.(*services.CreateAccountPayload)
		}
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
		var p *services.ShowAccountPayload
		if req != nil {
			p = req.(*services.ShowAccountPayload)
		}
		return s.Show(ctx, p)
	}

	ep.Delete = func(ctx context.Context, req interface{}) (interface{}, error) {
		var p *services.DeleteAccountPayload
		if req != nil {
			p = req.(*services.DeleteAccountPayload)
		}
		return s.Delete(ctx, p)
	}

	return ep
}`
	)
	var (
		create = design.EndpointExpr{
			Name: "Create",
			Payload: &design.UserTypeExpr{
				TypeName: "CreateAccountPayload",
			},
		}

		list = design.EndpointExpr{
			Name: "List",
			Payload: &design.UserTypeExpr{
				TypeName: "ListAccountPayload",
			},
		}

		show = design.EndpointExpr{
			Name: "Show",
			Payload: &design.UserTypeExpr{
				TypeName: "ShowAccountPayload",
			},
		}

		delete = design.EndpointExpr{
			Name: "Delete",
			Payload: &design.UserTypeExpr{
				TypeName: "DeleteAccountPayload",
			},
		}

		withSingleEndpoint = design.ServiceExpr{
			Name: "Account",
			Endpoints: []*design.EndpointExpr{
				&create,
			},
		}

		withMultipleEndpoints = design.ServiceExpr{
			Name: "Account",
			Endpoints: []*design.EndpointExpr{
				&create,
				&list,
				&show,
				&delete,
			},
		}
	)
	cases := map[string]struct {
		API      *design.APIExpr
		Service  *design.ServiceExpr
		Expected string
	}{
		"single-method":    {Service: &withSingleEndpoint, Expected: singleMethod},
		"multiple-methods": {Service: &withMultipleEndpoints, Expected: multipleMethods},
	}
	for k, tc := range cases {
		buf := new(bytes.Buffer)
		s := EndpointSection(tc.API, tc.Service)
		s.Render(buf)
		actual := buf.String()
		if actual != tc.Expected {
			t.Errorf("%s: got %v, expected %v", k, actual, tc.Expected)
		}
	}
}
