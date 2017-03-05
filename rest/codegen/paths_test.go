package codegen

import (
	"bytes"
	"testing"

	goadesign "goa.design/goa.v2/design"
	"goa.design/goa.v2/rest/design"
)

func TestPaths(t *testing.T) {
	const (
		pathWithoutParams = `
// AccountShowPath returns the URL path to the account service show HTTP endpoint.
func AccountShowPath() string {
	return "/account/test"
}

`

		pathWithOneParam = `
// AccountShowPath returns the URL path to the account service show HTTP endpoint.
func AccountShowPath(id int32) string {
	return fmt.Sprintf("/account/test/%v", id)
}

`
		pathWithMultipleParams = `
// AccountShowPath returns the URL path to the account service show HTTP endpoint.
func AccountShowPath(id int32, view string) string {
	return fmt.Sprintf("/account/test/%v/view/%v", id, view)
}

`

		pathWithAlternatives = `
// AccountShowPath returns the URL path to the account service show HTTP endpoint.
func AccountShowPath() string {
	return "/account/test"
}

// AccountShowAlternativePath returns the URL path to the account service show HTTP endpoint.
func AccountShowAlternativePath(id int32) string {
	return fmt.Sprintf("/account/test/%v", id)
}

// AccountShowAlternativePath1 returns the URL path to the account service show HTTP endpoint.
func AccountShowAlternativePath1(id int32, view string) string {
	return fmt.Sprintf("/account/test/%v/view/%v", id, view)
}

`
	)
	var (
		setParams = func(a *goadesign.AttributeExpr) {
			a.Type = goadesign.Object{
				"id":   {Type: goadesign.Int32},
				"view": {Type: goadesign.String},
			}
		}

		service = goadesign.ServiceExpr{
			Name: "Account",
		}

		endpoint = goadesign.EndpointExpr{
			Name:    "Show",
			Service: &service,
			//Payload: &goadesign.UserTypeExpr{AttributeExpr: &params },
		}

		resource = design.ResourceExpr{
			Path: "/account",
		}

		action = design.ActionExpr{
			EndpointExpr: &endpoint,
			Resource:     &resource,
			//Body:         &params,
			Routes: []*design.RouteExpr{
				{Path: "/test"},
			},
		}

		actionOneParam = design.ActionExpr{
			EndpointExpr: &endpoint,
			Resource:     &resource,
			Routes: []*design.RouteExpr{
				{Path: "/test/:id"},
			},
		}

		actionMultipleParams = design.ActionExpr{
			EndpointExpr: &endpoint,
			Resource:     &resource,
			Routes: []*design.RouteExpr{
				{Path: "/test/:id/view/:view"},
			},
		}

		actionWithAlternativePaths = design.ActionExpr{
			EndpointExpr: &endpoint,
			Resource:     &resource,
			Routes: []*design.RouteExpr{
				{Path: "/test"},
				{Path: "/test/:id"},
				{Path: "/test/:id/view/:view"},
			},
		}
	)

	setParams(actionOneParam.Params())
	setParams(actionMultipleParams.Params())
	setParams(actionWithAlternativePaths.Params())

	linkRouteToAction := func(a *design.ActionExpr) {
		for _, r := range a.Routes {
			r.Action = a
		}
	}

	cases := map[string]struct {
		Action   *design.ActionExpr
		Expected string
	}{
		"single-path-no-param":        {Action: &action, Expected: pathWithoutParams},
		"single-path-one-param":       {Action: &actionOneParam, Expected: pathWithOneParam},
		"single-path-multiple-params": {Action: &actionMultipleParams, Expected: pathWithMultipleParams},
		"alternative-paths":           {Action: &actionWithAlternativePaths, Expected: pathWithAlternatives},
	}

	for k, tc := range cases {
		linkRouteToAction(tc.Action)
		buf := new(bytes.Buffer)
		s := Path(tc.Action)
		s.Render(buf)
		actual := buf.String()

		if actual != tc.Expected {
			t.Errorf("%s: got %v, expected %v", k, actual, tc.Expected)
		}
	}
}
