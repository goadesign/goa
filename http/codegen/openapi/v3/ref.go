package openapiv3

import (
	"encoding/json"
)

type (
	// ParameterRef represents an OpenAPI reference to a Parameter object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	ParameterRef struct {
		Ref   string
		Value *Parameter
	}

	// ResponseRef represents an OpenAPI reference to a Response object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	ResponseRef struct {
		Ref   string
		Value *Response
	}

	// HeaderRef represents an OpenAPI reference to a Header object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	HeaderRef struct {
		Ref   string
		Value *Header
	}

	// CallbackRef represents an OpenAPI reference to a Callback object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	CallbackRef struct {
		Ref   string
		Value map[string]*PathItem
	}

	// ExampleRef represents an OpenAPI reference to a Example object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	ExampleRef struct {
		Ref   string
		Value *Example

		// LinkRef represents an OpenAPI reference to a Link object as defined in
		// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	}

	// LinkRef represents an OpenAPI reference to a Link object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	LinkRef struct {
		Ref   string
		Value *Link
	}

	// RequestBodyRef represents an OpenAPI reference to a RequestBody object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	RequestBodyRef struct {
		Ref   string
		Value *RequestBody
	}

	// SecuritySchemeRef represents an OpenAPI reference to a SecurityScheme object as defined in
	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.3.md#referenceObject
	SecuritySchemeRef struct {
		Ref   string
		Value *SecurityScheme
	}
)

func (r *ParameterRef) MarshalJSON() ([]byte, error)      { return marshalJSONRef(r.Ref, r.Value) }
func (r *ResponseRef) MarshalJSON() ([]byte, error)       { return marshalJSONRef(r.Ref, r.Value) }
func (r *HeaderRef) MarshalJSON() ([]byte, error)         { return marshalJSONRef(r.Ref, r.Value) }
func (r *CallbackRef) MarshalJSON() ([]byte, error)       { return marshalJSONRef(r.Ref, r.Value) }
func (r *ExampleRef) MarshalJSON() ([]byte, error)        { return marshalJSONRef(r.Ref, r.Value) }
func (r *LinkRef) MarshalJSON() ([]byte, error)           { return marshalJSONRef(r.Ref, r.Value) }
func (r *RequestBodyRef) MarshalJSON() ([]byte, error)    { return marshalJSONRef(r.Ref, r.Value) }
func (r *SecuritySchemeRef) MarshalJSON() ([]byte, error) { return marshalJSONRef(r.Ref, r.Value) }

func (r *ParameterRef) MarshalYAML() (interface{}, error)      { return marshalYAMLRef(r.Ref, r.Value) }
func (r *ResponseRef) MarshalYAML() (interface{}, error)       { return marshalYAMLRef(r.Ref, r.Value) }
func (r *HeaderRef) MarshalYAML() (interface{}, error)         { return marshalYAMLRef(r.Ref, r.Value) }
func (r *CallbackRef) MarshalYAML() (interface{}, error)       { return marshalYAMLRef(r.Ref, r.Value) }
func (r *ExampleRef) MarshalYAML() (interface{}, error)        { return marshalYAMLRef(r.Ref, r.Value) }
func (r *LinkRef) MarshalYAML() (interface{}, error)           { return marshalYAMLRef(r.Ref, r.Value) }
func (r *RequestBodyRef) MarshalYAML() (interface{}, error)    { return marshalYAMLRef(r.Ref, r.Value) }
func (r *SecuritySchemeRef) MarshalYAML() (interface{}, error) { return marshalYAMLRef(r.Ref, r.Value) }

func (r *ParameterRef) UnmarshalJSON(d []byte) error   { return unmarshalJSONRef(d, &r.Ref, &r.Value) }
func (r *ResponseRef) UnmarshalJSON(d []byte) error    { return unmarshalJSONRef(d, &r.Ref, &r.Value) }
func (r *HeaderRef) UnmarshalJSON(d []byte) error      { return unmarshalJSONRef(d, &r.Ref, &r.Value) }
func (r *CallbackRef) UnmarshalJSON(d []byte) error    { return unmarshalJSONRef(d, &r.Ref, &r.Value) }
func (r *ExampleRef) UnmarshalJSON(d []byte) error     { return unmarshalJSONRef(d, &r.Ref, &r.Value) }
func (r *LinkRef) UnmarshalJSON(d []byte) error        { return unmarshalJSONRef(d, &r.Ref, &r.Value) }
func (r *RequestBodyRef) UnmarshalJSON(d []byte) error { return unmarshalJSONRef(d, &r.Ref, &r.Value) }
func (r *SecuritySchemeRef) UnmarshalJSON(d []byte) error {
	return unmarshalJSONRef(d, &r.Ref, &r.Value)
}

func (r *ParameterRef) UnmarshalYAML(u func(interface{}) error) error {
	return unmarshalYAMLRef(u, &r.Ref, &r.Value)
}
func (r *ResponseRef) UnmarshalYAML(u func(interface{}) error) error {
	return unmarshalYAMLRef(u, &r.Ref, &r.Value)
}
func (r *HeaderRef) UnmarshalYAML(u func(interface{}) error) error {
	return unmarshalYAMLRef(u, &r.Ref, &r.Value)
}
func (r *CallbackRef) UnmarshalYAML(u func(interface{}) error) error {
	return unmarshalYAMLRef(u, &r.Ref, &r.Value)
}
func (r *ExampleRef) UnmarshalYAML(u func(interface{}) error) error {
	return unmarshalYAMLRef(u, &r.Ref, &r.Value)
}
func (r *LinkRef) UnmarshalYAML(u func(interface{}) error) error {
	return unmarshalYAMLRef(u, &r.Ref, &r.Value)
}
func (r *RequestBodyRef) UnmarshalYAML(u func(interface{}) error) error {
	return unmarshalYAMLRef(u, &r.Ref, &r.Value)
}
func (r *SecuritySchemeRef) UnmarshalYAML(u func(interface{}) error) error {
	return unmarshalYAMLRef(u, &r.Ref, &r.Value)
}

type refs struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

func marshalJSONRef(ref string, v interface{}) ([]byte, error) {
	if len(ref) > 0 {
		return json.Marshal(&refs{ref})
	}
	return json.Marshal(v)
}

func unmarshalJSONRef(data []byte, ref *string, v interface{}) error {
	refs := &refs{}
	if err := json.Unmarshal(data, refs); err == nil {
		if len(refs.Ref) > 0 {
			*ref = refs.Ref
			return nil
		}
	}
	return json.Unmarshal(data, v)
}

func marshalYAMLRef(ref string, v interface{}) (interface{}, error) {
	if len(ref) > 0 {
		return &refs{ref}, nil
	}
	return v, nil
}

func unmarshalYAMLRef(u func(interface{}) error, ref *string, v interface{}) error {
	refs := &refs{}
	if err := u(refs); err == nil {
		if len(refs.Ref) > 0 {
			*ref = refs.Ref
			return nil
		}
	}
	return u(v)
}
