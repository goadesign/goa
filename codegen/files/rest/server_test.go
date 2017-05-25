package rest

import (
	"bytes"
	"testing"

	. "goa.design/goa.v2/codegen/testing"
	"goa.design/goa.v2/design"
	restdesign "goa.design/goa.v2/design/rest"
	. "goa.design/goa.v2/design/rest/testing"

	r "goa.design/goa.v2/dsl/rest"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		Name string
		DSL  func()
		Code string
	}{
		{Name: "body-string", DSL: PayloadBodyDSL, Code: PayloadBodyCode},
		{Name: "body-string-validate", DSL: PayloadBodyValidateDSL, Code: PayloadBodyValidateCode},
		{Name: "body-object", DSL: PayloadObjectBodyDSL, Code: PayloadObjectBodyCode},
		{Name: "body-object-validate", DSL: PayloadObjectBodyValidateDSL, Code: PayloadObjectBodyValidateCode},
		{Name: "body-user", DSL: PayloadUserBodyDSL, Code: PayloadUserBodyCode},
		{Name: "body-user-validate", DSL: PayloadUserBodyValidateDSL, Code: PayloadUserBodyValidateCode},
		{Name: "body-array-string", DSL: PayloadArrayStringBodyDSL, Code: PayloadArrayStringBodyCode},
		{Name: "body-array-string-validate", DSL: PayloadArrayStringBodyValidateDSL, Code: PayloadArrayStringBodyValidateCode},
		{Name: "body-array-user", DSL: PayloadArrayUserBodyDSL, Code: PayloadArrayUserBodyCode},
		{Name: "body-array-user-validate", DSL: PayloadArrayUserBodyValidateDSL, Code: PayloadArrayUserBodyValidateCode},
		{Name: "body-map-string", DSL: PayloadMapStringBodyDSL, Code: PayloadMapStringBodyCode},
		{Name: "body-map-string-validate", DSL: PayloadMapStringBodyValidateDSL, Code: PayloadMapStringBodyValidateCode},
		{Name: "body-map-user", DSL: PayloadMapUserBodyDSL, Code: PayloadMapUserBodyCode},
		{Name: "body-map-user-validate", DSL: PayloadMapUserBodyValidateDSL, Code: PayloadMapUserBodyValidateCode},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			RunRestDSL(t, c.DSL)
			fs := ServerFiles(restdesign.Root)
			if len(fs) != 1 {
				t.Fatalf("got %d files, expected one", len(fs))
			}
			sections := fs[0].Sections("")
			if len(sections) != 7 {
				t.Fatalf("got %d sections, expected 7", len(sections))
			}
			var code bytes.Buffer
			if err := sections[6].Write(&code); err != nil {
				t.Fatal(err)
			}
			if code.String() != c.Code {
				t.Errorf("invalid code, got:\n%s\ngot vs. expected:\n%s", code.String(), Diff(t, code.String(), c.Code))
			}
		})
	}
}

var PayloadBodyDSL = func() {
	r.Service("ServiceBody", func() {
		r.Endpoint("EndpointBody", func() {
			r.Payload(design.String)
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadBodyCode = `// DecodeEndpointBodyRequest returns a decoder for requests sent to the
// ServiceBody EndpointBody endpoint.
func DecodeEndpointBodyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (string, error) {
		var (
			body string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadBodyValidateDSL = func() {
	r.Service("ServiceBodyValidate", func() {
		r.Endpoint("EndpointBodyValidate", func() {
			r.Payload(design.String, func() {
				r.Pattern("pattern")
			})
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadBodyValidateCode = `// DecodeEndpointBodyValidateRequest returns a decoder for requests sent to the
// ServiceBodyValidate EndpointBodyValidate endpoint.
func DecodeEndpointBodyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (string, error) {
		var (
			body string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
err = goa.MergeErrors(err, goa.ValidatePattern("body", body, "pattern"))
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadObjectBodyDSL = func() {
	r.Service("ServiceObjectBody", func() {
		r.Endpoint("EndpointObjectBody", func() {
			r.Payload(func() {
				r.Attribute("a", design.String)
			})
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadObjectBodyCode = `// DecodeEndpointObjectBodyRequest returns a decoder for requests sent to the
// ServiceObjectBody EndpointObjectBody endpoint.
func DecodeEndpointObjectBodyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointObjectBodyPayload, error) {
		var (
			body EndpointObjectBodyPayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadObjectBodyValidateDSL = func() {
	r.Service("ServiceObjectBodyValidate", func() {
		r.Endpoint("EndpointObjectBodyValidate", func() {
			r.Payload(func() {
				r.Attribute("a", design.String, func() {
					r.Pattern("pattern")
				})
				r.Required("a")
			})
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadObjectBodyValidateCode = `// DecodeEndpointObjectBodyValidateRequest returns a decoder for requests sent
// to the ServiceObjectBodyValidate EndpointObjectBodyValidate endpoint.
func DecodeEndpointObjectBodyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*EndpointObjectBodyValidatePayload, error) {
		var (
			body EndpointObjectBodyValidatePayload
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadUserBodyDSL = func() {
	var PayloadType = r.Type("PayloadType", func() {
		r.Attribute("a", design.String)
	})
	r.Service("ServiceUserBody", func() {
		r.Endpoint("EndpointUserBody", func() {
			r.Payload(PayloadType)
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadUserBodyCode = `// DecodeEndpointUserBodyRequest returns a decoder for requests sent to the
// ServiceUserBody EndpointUserBody endpoint.
func DecodeEndpointUserBodyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadUserBodyValidateDSL = func() {
	var PayloadType = r.Type("PayloadType", func() {
		r.Attribute("a", design.String, func() {
			r.Pattern("apattern")
		})
	})
	r.Service("ServiceUserBodyValidate", func() {
		r.Endpoint("EndpointUserBodyValidate", func() {
			r.Payload(PayloadType)
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadUserBodyValidateCode = `// DecodeEndpointUserBodyValidateRequest returns a decoder for requests sent to
// the ServiceUserBodyValidate EndpointUserBodyValidate endpoint.
func DecodeEndpointUserBodyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (*PayloadType, error) {
		var (
			body PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
		if err := body.Validate(); err != nil {
			return nil, err
		}

		return &body, nil
	}
}
`

var PayloadArrayStringBodyDSL = func() {
	r.Service("ServiceArrayStringBody", func() {
		r.Endpoint("EndpointArrayStringBody", func() {
			r.Payload(r.ArrayOf(design.String))
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadArrayStringBodyCode = `// DecodeEndpointArrayStringBodyRequest returns a decoder for requests sent to
// the ServiceArrayStringBody EndpointArrayStringBody endpoint.
func DecodeEndpointArrayStringBodyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]string, error) {
		var (
			body []string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadArrayStringBodyValidateDSL = func() {
	r.Service("ServiceArrayStringBodyValidate", func() {
		r.Endpoint("EndpointArrayStringBodyValidate", func() {
			r.Payload(r.ArrayOf(design.String), func() {
				r.MinLength(2)
				r.Elem(func() {
					r.MinLength(3)
				})
			})
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadArrayStringBodyValidateCode = `// DecodeEndpointArrayStringBodyValidateRequest returns a decoder for requests
// sent to the ServiceArrayStringBodyValidate EndpointArrayStringBodyValidate
// endpoint.
func DecodeEndpointArrayStringBodyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]string, error) {
		var (
			body []string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
if len(body) < 2 {
	err = goa.MergeErrors(err, goa.InvalidLengthError("body", body, len(body), 2, true))
}
for _, e := range body {
if utf8.RuneCountInString(e) < 3 {
	err = goa.MergeErrors(err, goa.InvalidLengthError("body[*]", e, utf8.RuneCountInString(e), 3, true))
}
}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadArrayUserBodyDSL = func() {
	var PayloadType = r.Type("PayloadType", func() {
		r.Attribute("a", design.String, func() {
			r.Pattern("apattern")
		})
	})
	r.Service("ServiceArrayUserBody", func() {
		r.Endpoint("EndpointArrayUserBody", func() {
			r.Payload(r.ArrayOf(PayloadType))
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadArrayUserBodyCode = `// DecodeEndpointArrayUserBodyRequest returns a decoder for requests sent to
// the ServiceArrayUserBody EndpointArrayUserBody endpoint.
func DecodeEndpointArrayUserBodyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]*PayloadType, error) {
		var (
			body []*PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
for _, e := range body {
if e != nil {
	if err2 := e.Validate(); err2 != nil {
	err = goa.MergeErrors(err, err2)
}
}
}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadArrayUserBodyValidateDSL = func() {
	var PayloadType = r.Type("PayloadType", func() {
		r.Attribute("a", design.String, func() {
			r.Pattern("apattern")
		})
	})
	r.Service("ServiceArrayUserBodyValidate", func() {
		r.Endpoint("EndpointArrayUserBodyValidate", func() {
			r.Payload(r.ArrayOf(PayloadType), func() {
				r.MinLength(2)
			})
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadArrayUserBodyValidateCode = `// DecodeEndpointArrayUserBodyValidateRequest returns a decoder for requests
// sent to the ServiceArrayUserBodyValidate EndpointArrayUserBodyValidate
// endpoint.
func DecodeEndpointArrayUserBodyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) ([]*PayloadType, error) {
		var (
			body []*PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
if len(body) < 2 {
	err = goa.MergeErrors(err, goa.InvalidLengthError("body", body, len(body), 2, true))
}
for _, e := range body {
if e != nil {
	if err2 := e.Validate(); err2 != nil {
	err = goa.MergeErrors(err, err2)
}
}
}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadMapStringBodyDSL = func() {
	r.Service("ServiceMapStringBody", func() {
		r.Endpoint("EndpointMapStringBody", func() {
			r.Payload(r.MapOf(design.String, design.String))
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadMapStringBodyCode = `// DecodeEndpointMapStringBodyRequest returns a decoder for requests sent to
// the ServiceMapStringBody EndpointMapStringBody endpoint.
func DecodeEndpointMapStringBodyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (map[string]string, error) {
		var (
			body map[string]string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadMapStringBodyValidateDSL = func() {
	r.Service("ServiceMapStringBodyValidate", func() {
		r.Endpoint("EndpointMapStringBodyValidate", func() {
			r.Payload(r.MapOf(design.String, design.String, func() {
				r.Elem(func() {
					r.MinLength(2)
				})
			}))
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadMapStringBodyValidateCode = `// DecodeEndpointMapStringBodyValidateRequest returns a decoder for requests
// sent to the ServiceMapStringBodyValidate EndpointMapStringBodyValidate
// endpoint.
func DecodeEndpointMapStringBodyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (map[string]string, error) {
		var (
			body map[string]string
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
for _, v := range body {
if utf8.RuneCountInString(v) < 2 {
	err = goa.MergeErrors(err, goa.InvalidLengthError("body[key]", v, utf8.RuneCountInString(v), 2, true))
}
}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadMapUserBodyDSL = func() {
	var PayloadType = r.Type("PayloadType", func() {
		r.Attribute("a", design.String, func() {
			r.Pattern("apattern")
		})
	})
	r.Service("ServiceMapUserBody", func() {
		r.Endpoint("EndpointMapUserBody", func() {
			r.Payload(r.MapOf(design.String, PayloadType))
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadMapUserBodyCode = `// DecodeEndpointMapUserBodyRequest returns a decoder for requests sent to the
// ServiceMapUserBody EndpointMapUserBody endpoint.
func DecodeEndpointMapUserBodyRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (map[string]*PayloadType, error) {
		var (
			body map[string]*PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
for _, v := range body {
if v != nil {
	if err2 := v.Validate(); err2 != nil {
	err = goa.MergeErrors(err, err2)
}
}
}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`

var PayloadMapUserBodyValidateDSL = func() {
	var PayloadType = r.Type("PayloadType", func() {
		r.Attribute("a", design.String, func() {
			r.Pattern("apattern")
		})
	})
	r.Service("ServiceMapUserBodyValidate", func() {
		r.Endpoint("EndpointMapUserBodyValidate", func() {
			r.Payload(r.MapOf(design.String, PayloadType, func() {
				r.Key(func() {
					r.MinLength(2)
				})
			}))
			r.HTTP(func() {
				r.POST("/")
			})
		})
	})
}

var PayloadMapUserBodyValidateCode = `// DecodeEndpointMapUserBodyValidateRequest returns a decoder for requests sent
// to the ServiceMapUserBodyValidate EndpointMapUserBodyValidate endpoint.
func DecodeEndpointMapUserBodyValidateRequest(decoder func(*http.Request) rest.Decoder) func(*http.Request) (interface{}, error) {
	return func(r *http.Request) (map[string]*PayloadType, error) {
		var (
			body map[string]*PayloadType
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				err = goa.MissingPayloadError()
			}
			return nil, err
		}
for k, v := range body {
if utf8.RuneCountInString(k) < 2 {
	err = goa.MergeErrors(err, goa.InvalidLengthError("body.key", k, utf8.RuneCountInString(k), 2, true))
}
if v != nil {
	if err2 := v.Validate(); err2 != nil {
	err = goa.MergeErrors(err, err2)
}
}
}
		if err != nil {
			return nil, err
		}

		return body, nil
	}
}
`
