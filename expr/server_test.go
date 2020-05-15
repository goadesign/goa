package expr

import (
	"fmt"
	"reflect"
	"testing"

	"goa.design/goa/v3/eval"
)

func TestServerExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name     string
		expected string
	}{
		"foo": {name: "foo", expected: "Server foo"},
	}

	for k, tc := range cases {
		server := ServerExpr{Name: tc.name}
		if actual := server.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestServerExprValidate(t *testing.T) {
	const (
		foo = "foo"
		bar = "bar"
	)
	var (
		validURI  = URIExpr("http://example.com")
		validURIs = []URIExpr{
			validURI,
		}
		errNoURI            = fmt.Errorf("host must define at least one URI")
		errServiceUndefined = fmt.Errorf("service %q undefined", bar)
	)

	cases := map[string]struct {
		hosts    []*HostExpr
		services []string
		expected *eval.ValidationErrors
	}{
		"no error": {
			hosts: []*HostExpr{
				{
					URIs: validURIs,
				},
			},
			services: []string{
				foo,
			},
			expected: &eval.ValidationErrors{
				Errors: []error{},
			},
		},
		"error only in hosts": {
			hosts: []*HostExpr{
				{
					URIs: []URIExpr{},
				},
			},
			expected: &eval.ValidationErrors{
				Errors: []error{
					errNoURI,
				},
			},
		},
		"error only in services": {
			services: []string{
				bar,
			},
			expected: &eval.ValidationErrors{
				Errors: []error{
					errServiceUndefined,
				},
			},
		},
		"error in both": {
			hosts: []*HostExpr{
				{
					URIs: []URIExpr{},
				},
			},
			services: []string{
				bar,
			},
			expected: &eval.ValidationErrors{
				Errors: []error{
					errNoURI,
					errServiceUndefined,
				},
			},
		},
	}

	services := Root.Services
	Root.Services = []*ServiceExpr{
		{
			Name: foo,
		},
	}
	defer func() {
		Root.Services = services
	}()

	for k, tc := range cases {
		s := ServerExpr{
			Hosts:    tc.hosts,
			Services: tc.services,
		}
		if actual := s.Validate().(*eval.ValidationErrors); len(tc.expected.Errors) != len(actual.Errors) {
			t.Errorf("%s: expected the number of error values to match %d got %d ", k, len(tc.expected.Errors), len(actual.Errors))
		} else {
			for i, err := range actual.Errors {
				if err.Error() != tc.expected.Errors[i].Error() {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, err, tc.expected.Errors[i], i)
				}
			}
		}
	}
}

func TestHostExprValidate(t *testing.T) {
	const (
		foo = "foo"
		bar = "bar"
	)
	var (
		validURI         = URIExpr("http://example.com")
		malformedURI     = URIExpr("http://%")
		missingSchemeURI = URIExpr("example.com")
		invalidSchemeURI = URIExpr("ftp:example.com")
		validURIs        = []URIExpr{
			validURI,
		}
		malformedURIs = []URIExpr{
			malformedURI,
		}
		missingSchemeURIs = []URIExpr{
			missingSchemeURI,
		}
		invalidSchemeURIs = []URIExpr{
			invalidSchemeURI,
		}
		objectPrimitive = func(d interface{}, v *ValidationExpr) *Object {
			return &Object{
				{
					Name: foo,
					Attribute: &AttributeExpr{
						Type:         Boolean,
						DefaultValue: d,
						Validation:   v,
					},
				},
			}
		}
		objectNonPrimitive = func(d interface{}) *Object {
			return &Object{
				{
					Name: bar,
					Attribute: &AttributeExpr{
						Type:         objectPrimitive(nil, nil),
						DefaultValue: d,
					},
				},
			}
		}
		attribute = func(d DataType) *AttributeExpr {
			return &AttributeExpr{
				Type: d,
			}
		}
		errNoURI                          = fmt.Errorf("host must define at least one URI")
		errMalformedURI                   = fmt.Errorf("malformed URI %q", malformedURI)
		errMissingSchemeURI               = fmt.Errorf("missing scheme for URI %q, scheme must be one of 'http', 'https', 'grpc' or 'grpcs'", missingSchemeURI)
		errInvalidSchemeURI               = fmt.Errorf("invalid scheme for URI %q, scheme must be one of 'http', 'https', 'grpc' or 'grpcs'", invalidSchemeURI)
		errInvalidType                    = fmt.Errorf("invalid type for URI variable %q: type must be a primitive", bar)
		errNoDefaultValueOrEnumValidation = fmt.Errorf("URI variable %q must have a default value or an enum validation", foo)
	)

	cases := map[string]struct {
		uris      []URIExpr
		variables *AttributeExpr
		expected  *eval.ValidationErrors
	}{
		"no error": {
			uris: validURIs,
			expected: &eval.ValidationErrors{
				Errors: []error{},
			},
		},
		"no uri": {
			uris: []URIExpr{},
			expected: &eval.ValidationErrors{
				Errors: []error{
					errNoURI,
				},
			},
		},
		"malformed uri": {
			uris: malformedURIs,
			expected: &eval.ValidationErrors{
				Errors: []error{
					errMalformedURI,
				},
			},
		},
		"missing scheme": {
			uris: missingSchemeURIs,
			expected: &eval.ValidationErrors{
				Errors: []error{
					errMissingSchemeURI,
				},
			},
		},
		"invalid scheme": {
			uris: invalidSchemeURIs,
			expected: &eval.ValidationErrors{
				Errors: []error{
					errInvalidSchemeURI,
				},
			},
		},
		"invalid type for uri variable": {
			uris: validURIs,
			variables: attribute(objectNonPrimitive(map[string]int{
				"a": 1,
			})),
			expected: &eval.ValidationErrors{
				Errors: []error{
					errInvalidType,
				},
			},
		},
		"uri variable has no default value and no enum validation": {
			uris:      validURIs,
			variables: attribute(objectPrimitive(nil, nil)),
			expected: &eval.ValidationErrors{
				Errors: []error{
					errNoDefaultValueOrEnumValidation,
				},
			},
		},
		"uri variable has no default value and no enum varidation values": {
			uris: validURIs,
			variables: attribute(objectPrimitive(nil, &ValidationExpr{
				Values: []interface{}{},
			})),
			expected: &eval.ValidationErrors{
				Errors: []error{
					errNoDefaultValueOrEnumValidation,
				},
			},
		},
	}

	for k, tc := range cases {
		h := HostExpr{
			URIs:      tc.uris,
			Variables: tc.variables,
		}
		if actual := h.Validate().(*eval.ValidationErrors); len(tc.expected.Errors) != len(actual.Errors) {
			t.Errorf("%s: expected the number of error values to match %d got %d ", k, len(tc.expected.Errors), len(actual.Errors))
		} else {
			for i, err := range actual.Errors {
				if err.Error() != tc.expected.Errors[i].Error() {
					t.Errorf("%s: got %#v, expected %#v at index %d", k, err, tc.expected.Errors[i], i)
				}
			}
		}
	}
}

func TestHostExprEvalName(t *testing.T) {
	cases := map[string]struct {
		name       string
		serverName string
		expected   string
	}{
		"foo": {name: "foo", serverName: "bar", expected: `host "foo" of server "bar"`},
	}

	for k, tc := range cases {
		host := HostExpr{Name: tc.name, ServerName: tc.serverName}
		if actual := host.EvalName(); actual != tc.expected {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}

func TestHostExprAttribute(t *testing.T) {
	cases := map[string]struct {
		attributeExpr *AttributeExpr
		expected      *AttributeExpr
	}{
		"nil": {
			attributeExpr: nil,
			expected:      &AttributeExpr{Type: &Object{}},
		},
		"non-nil": {
			attributeExpr: &AttributeExpr{Description: "foo"},
			expected:      &AttributeExpr{Description: "foo"},
		},
	}

	for k, tc := range cases {
		host := HostExpr{Variables: tc.attributeExpr}
		actual := host.Attribute()

		actualType := reflect.TypeOf(actual)
		expectedValue := reflect.ValueOf(tc.expected)

		if !reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual) {
			t.Errorf("%s: got %#v, expected %#v", k, actual, tc.expected)
		}
	}
}
