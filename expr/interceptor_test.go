package expr

import (
	"testing"
)

func TestInterceptorExpr_Validate(t *testing.T) {
	cases := map[string]struct {
		intercept  *InterceptorExpr
		method     *MethodExpr
		wantErrors []string
	}{
		"valid-payload": {
			intercept: makeInterceptor(t, withReadPayload(t, namedAttr(t, "foo"))),
			method:    makeMethod(t, withPayload(t, namedAttr(t, "foo"))),
		},
		"valid-write-payload": {
			intercept: makeInterceptor(t, withWritePayload(t, namedAttr(t, "foo"))),
			method:    makeMethod(t, withPayload(t, namedAttr(t, "foo"))),
		},
		"payload-with-base": {
			intercept: makeInterceptor(t, withReadPayload(t, namedAttr(t, "bar"))),
			method: makeMethod(t,
				withPayload(t, namedAttr(t, "foo")),
				withPayloadBases(t, &UserTypeExpr{
					AttributeExpr: &AttributeExpr{
						Type: &Object{namedAttr(t, "bar")},
					},
				}),
			),
		},
		"result-with-base": {
			intercept: makeInterceptor(t, withReadResult(t, namedAttr(t, "bar"))),
			method: makeMethod(t,
				withResult(t, namedAttr(t, "foo")),
				withResultBases(t, &UserTypeExpr{
					AttributeExpr: &AttributeExpr{
						Type: &Object{namedAttr(t, "bar")},
					},
				}),
			),
		},
		"invalid-payload-not-object": {
			intercept: makeInterceptor(t, withReadPayload(t, namedAttr(t, "foo"))),
			method: makeMethod(t, func(m *MethodExpr) {
				m.Payload = &AttributeExpr{
					Type: String,
				}
			}),
			wantErrors: []string{
				`interceptor "test-interceptor" cannot be applied because the method payload is not an object`,
			},
		},
		"invalid-streaming-payload-not-streaming": {
			intercept: makeInterceptor(t, withReadStreamingPayload(t, namedAttr(t, "foo"))),
			method: makeMethod(t, func(m *MethodExpr) {
				m.Payload = &AttributeExpr{Type: &Object{}}
			}),
			wantErrors: []string{
				`interceptor "test-interceptor" cannot be applied because the method payload is not streaming`,
			},
		},
		"invalid-streaming-result-not-streaming": {
			intercept: makeInterceptor(t, withReadStreamingResult(t, namedAttr(t, "foo"))),
			method: makeMethod(t, func(m *MethodExpr) {
				m.Result = &AttributeExpr{Type: &Object{}}
			}),
			wantErrors: []string{
				`interceptor "test-interceptor" cannot be applied because the method result is not streaming`,
			},
		},
		"invalid-attribute-access": {
			intercept: makeInterceptor(t, withReadPayload(t, namedAttr(t, "bar"))),
			method:    makeMethod(t, withPayload(t, namedAttr(t, "foo"))),
			wantErrors: []string{
				`interceptor "test-interceptor" cannot read payload attribute "bar": attribute does not exist`,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			verr := tc.intercept.validate(tc.method)
			if len(tc.wantErrors) != len(verr.Errors) {
				t.Errorf("got %d errors, expected %d", len(verr.Errors), len(tc.wantErrors))
			}
			for i, err := range verr.Errors {
				if i >= len(tc.wantErrors) {
					break
				}
				if got := err.Error(); got != tc.wantErrors[i] {
					t.Errorf("got error %q, expected %q", got, tc.wantErrors[i])
				}
			}
		})
	}
}

// Test helpers (at the end of the file)
func withWritePayload(t *testing.T, attrs *NamedAttributeExpr) func(*InterceptorExpr) {
	t.Helper()
	return func(i *InterceptorExpr) {
		i.WritePayload = &AttributeExpr{Type: &Object{attrs}}
	}
}

func withResultBases(t *testing.T, bases ...DataType) func(*MethodExpr) {
	t.Helper()
	return func(m *MethodExpr) {
		if m.Result == nil {
			m.Result = &AttributeExpr{Type: &Object{}}
		}
		m.Result.Bases = bases
	}
}

func makeInterceptor(t *testing.T, opts ...func(*InterceptorExpr)) *InterceptorExpr {
	t.Helper()
	i := &InterceptorExpr{Name: "test-interceptor"}
	for _, opt := range opts {
		opt(i)
	}
	return i
}

// Helper functions need to be updated to handle empty attributes properly
func withReadPayload(t *testing.T, attrs *NamedAttributeExpr) func(*InterceptorExpr) {
	t.Helper()
	return func(i *InterceptorExpr) {
		i.ReadPayload = &AttributeExpr{Type: &Object{attrs}}
	}
}

func withReadResult(t *testing.T, attrs *NamedAttributeExpr) func(*InterceptorExpr) {
	t.Helper()
	return func(i *InterceptorExpr) {
		i.ReadResult = &AttributeExpr{Type: &Object{attrs}}
	}
}

func withReadStreamingPayload(t *testing.T, attrs *NamedAttributeExpr) func(*InterceptorExpr) {
	t.Helper()
	return func(i *InterceptorExpr) {
		i.ReadStreamingPayload = &AttributeExpr{Type: &Object{attrs}}
	}
}

func withReadStreamingResult(t *testing.T, attrs *NamedAttributeExpr) func(*InterceptorExpr) {
	t.Helper()
	return func(i *InterceptorExpr) {
		i.ReadStreamingResult = &AttributeExpr{Type: &Object{attrs}}
	}
}

func makeMethod(t *testing.T, opts ...func(*MethodExpr)) *MethodExpr {
	t.Helper()
	m := &MethodExpr{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func withPayload(t *testing.T, attrs *NamedAttributeExpr) func(*MethodExpr) {
	t.Helper()
	return func(m *MethodExpr) {
		m.Payload = &AttributeExpr{Type: &Object{attrs}}
	}
}

func withPayloadBases(t *testing.T, bases ...DataType) func(*MethodExpr) {
	t.Helper()
	return func(m *MethodExpr) {
		if m.Payload == nil {
			m.Payload = &AttributeExpr{Type: &Object{}}
		}
		m.Payload.Bases = bases
	}
}

func withResult(t *testing.T, attrs *NamedAttributeExpr) func(*MethodExpr) {
	t.Helper()
	return func(m *MethodExpr) {
		m.Result = &AttributeExpr{Type: &Object{attrs}}
	}
}

func namedAttr(t *testing.T, name string) *NamedAttributeExpr {
	t.Helper()
	return &NamedAttributeExpr{
		Name: name,
		Attribute: &AttributeExpr{
			Type: String,
		},
	}
}
