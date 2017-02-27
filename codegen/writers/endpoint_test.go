package writers

import (
	"bytes"
	"testing"

	"goa.design/goa.v2/design"
)

func TestEndpointTmpl(t *testing.T) {
	cases := map[string]struct {
		data     endpointData
		expected string
	}{
		"a simple endpoint": {
			data: endpointData{
				Name:        "Account",
				Description: "",
				Methods: []*endpointMethod{
					&endpointMethod{
						Name:        "Create",
						Description: "",
						Payload: &design.UserTypeExpr{
							TypeName: "CreateAccountPayload",
						},
					},
					&endpointMethod{
						Name:        "List",
						Description: "",
						Payload: &design.UserTypeExpr{
							TypeName: "ListAccountPayload",
						},
					},
					&endpointMethod{
						Name:        "Show",
						Description: "",
						Payload: &design.UserTypeExpr{
							TypeName: "ShowAccountPayload",
						},
					},
					&endpointMethod{
						Name:        "Delete",
						Description: "",
						Payload: &design.UserTypeExpr{
							TypeName: "DeleteAccountPayload",
						},
					},
				},
			},
			expected: `type (
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
}`,
		},
	}
	for k, tc := range cases {
		buf := new(bytes.Buffer)
		if err := endpointTmpl.ExecuteTemplate(buf, "endpoint", tc.data); err != nil {
			t.Fatalf("Execute returned %s", err)
		}
		actual := buf.String()
		if actual != tc.expected {
			t.Errorf("%s: got %v, expected %v", k, actual, tc.expected)
		}
	}
}
