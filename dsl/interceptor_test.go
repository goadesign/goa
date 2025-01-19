package dsl_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestInterceptor(t *testing.T) {
	cases := map[string]struct {
		DSL    func()
		Assert func(t *testing.T, intr *expr.InterceptorExpr)
	}{
		"valid-minimal": {
			func() {
				Interceptor("minimal", func() {})
			},
			func(t *testing.T, intr *expr.InterceptorExpr) {
				require.NotNil(t, intr, "interceptor should not be nil")
				assert.Equal(t, "minimal", intr.Name)
			},
		},
		"valid-complete": {
			func() {
				Interceptor("complete", func() {
					Description("test interceptor")
					ReadPayload(func() {
						Attribute("foo", String)
					})
					WritePayload(func() {
						Attribute("bar", String)
					})
					ReadResult(func() {
						Attribute("baz", String)
					})
					WriteResult(func() {
						Attribute("qux", String)
					})
				})
			},
			func(t *testing.T, intr *expr.InterceptorExpr) {
				require.NotNil(t, intr, "interceptor should not be nil")
				assert.Equal(t, "test interceptor", intr.Description)

				require.NotNil(t, intr.ReadPayload, "ReadPayload should not be nil")
				rp := expr.AsObject(intr.ReadPayload.Type)
				require.NotNil(t, rp, "ReadPayload should be an object")
				assert.NotNil(t, rp.Attribute("foo"), "ReadPayload should have a foo attribute")

				require.NotNil(t, intr.WritePayload, "WritePayload should not be nil")
				wp := expr.AsObject(intr.WritePayload.Type)
				require.NotNil(t, wp, "WritePayload should be an object")
				assert.NotNil(t, wp.Attribute("bar"), "WritePayload should have a bar attribute")

				require.NotNil(t, intr.ReadResult, "ReadResult should not be nil")
				rr := expr.AsObject(intr.ReadResult.Type)
				require.NotNil(t, rr, "ReadResult should be an object")
				assert.NotNil(t, rr.Attribute("baz"), "ReadResult should have a baz attribute")

				require.NotNil(t, intr.WriteResult, "WriteResult should not be nil")
				wr := expr.AsObject(intr.WriteResult.Type)
				require.NotNil(t, wr, "WriteResult should be an object")
				assert.NotNil(t, wr.Attribute("qux"), "WriteResult should have a qux attribute")
			},
		},
		"empty-name": {
			func() {
				Interceptor("", func() {})
			},
			func(t *testing.T, intr *expr.InterceptorExpr) {
				assert.NotNil(t, eval.Context.Errors, "expected a validation error")
			},
		},
		"duplicate-name": {
			func() {
				Interceptor("duplicate", func() {})
				Interceptor("duplicate", func() {})
			},
			func(t *testing.T, intr *expr.InterceptorExpr) {
				if eval.Context.Errors == nil {
					t.Error("expected a validation error, got none")
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			expr.Root = new(expr.RootExpr)
			tc.DSL()
			if len(expr.Root.Interceptors) > 0 {
				tc.Assert(t, expr.Root.Interceptors[0])
			}
		})
	}
}

func TestServerInterceptor(t *testing.T) {
	cases := map[string]struct {
		DSL    func()
		Assert func(t *testing.T, svc *expr.ServiceExpr, err error)
	}{
		"valid-reference": {
			func() {
				var testInterceptor = Interceptor("test", func() {})
				Service("Service", func() {
					ServerInterceptor(testInterceptor)
				})
			},
			func(t *testing.T, svc *expr.ServiceExpr, err error) {
				require.NoError(t, err)
				require.NotNil(t, svc)
				require.Len(t, svc.ServerInterceptors, 1, "should have 1 server interceptor")
				assert.Equal(t, "test", svc.ServerInterceptors[0].Name)
			},
		},
		"valid-by-name": {
			func() {
				Interceptor("test", func() {})
				Service("Service", func() {
					ServerInterceptor("test")
				})
			},
			func(t *testing.T, svc *expr.ServiceExpr, err error) {
				require.NoError(t, err)
				require.NotNil(t, svc)
				require.Len(t, svc.ServerInterceptors, 1, "should have 1 server interceptor")
				assert.Equal(t, "test", svc.ServerInterceptors[0].Name)
			},
		},
		"invalid-reference": {
			func() {
				Service("Service", func() {
					ServerInterceptor(42) // Invalid type
				})
			},
			func(t *testing.T, svc *expr.ServiceExpr, err error) {
				require.Error(t, err)
			},
		},
		"invalid-name": {
			func() {
				Service("Service", func() {
					ServerInterceptor("invalid")
				})
			},
			func(t *testing.T, svc *expr.ServiceExpr, err error) {
				require.Error(t, err)
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			expr.Root = new(expr.RootExpr)
			tc.DSL()
			root, err := runDSL(t, tc.DSL)
			tc.Assert(t, root.Services[0], err)
		})
	}
}

func TestClientInterceptor(t *testing.T) {
	cases := map[string]struct {
		DSL    func()
		Assert func(t *testing.T, svc *expr.ServiceExpr, err error)
	}{
		"valid-reference": {
			func() {
				var testInterceptor = Interceptor("test", func() {})
				Service("Service", func() {
					ClientInterceptor(testInterceptor)
				})
			},
			func(t *testing.T, svc *expr.ServiceExpr, err error) {
				require.NoError(t, err)
				require.NotNil(t, svc)
				require.Len(t, svc.ClientInterceptors, 1, "should have 1 client interceptor")
			},
		},
		"valid-by-name": {
			func() {
				Interceptor("test", func() {})
				Service("Service", func() {
					ClientInterceptor("test")
				})
			},
			func(t *testing.T, svc *expr.ServiceExpr, err error) {
				require.NoError(t, err)
				require.NotNil(t, svc)
				require.Len(t, svc.ClientInterceptors, 1, "should have 1 client interceptor")
			},
		},
		"invalid-reference": {
			func() {
				Service("Service", func() {
					ClientInterceptor(42) // Invalid type
				})
			},
			func(t *testing.T, svc *expr.ServiceExpr, err error) {
				require.Error(t, err)
			},
		},
		"invalid-name": {
			func() {
				Service("Service", func() {
					ClientInterceptor("invalid")
				})
			},
			func(t *testing.T, svc *expr.ServiceExpr, err error) {
				require.Error(t, err)
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eval.Context = &eval.DSLContext{}
			expr.Root = new(expr.RootExpr)
			tc.DSL()
			root, err := runDSL(t, tc.DSL)
			tc.Assert(t, root.Services[0], err)
		})
	}
}

// runDSL returns the DSL root resulting from running the given DSL.
func runDSL(t *testing.T, dsl func()) (*expr.RootExpr, error) {
	t.Helper()
	eval.Reset()
	expr.Root = new(expr.RootExpr)
	expr.GeneratedResultTypes = new(expr.ResultTypesRoot)
	require.NoError(t, eval.Register(expr.Root))
	require.NoError(t, eval.Register(expr.GeneratedResultTypes))
	expr.Root.API = expr.NewAPIExpr("test api", func() {})
	expr.Root.API.Servers = []*expr.ServerExpr{expr.Root.API.DefaultServer()}
	if eval.Execute(dsl, nil) {
		return expr.Root, eval.RunDSL()
	} else {
		return expr.Root, errors.New(eval.Context.Error())
	}
}
