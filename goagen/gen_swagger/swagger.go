package genswagger

import (
	"fmt"
	"strconv"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/gen_schema"
)

type (
	// Swagger represents an instance of a swagger object.
	// See https://swagger.io/specification/
	Swagger struct {
		Swagger             string                           `json:"swagger,omitempty"`
		Info                *Info                            `json:"info,omitempty"`
		Host                string                           `json:"host,omitempty"`
		BasePath            string                           `json:"basePath,omitempty"`
		Schemes             []string                         `json:"schemes,omitempty"`
		Consumes            []string                         `json:"consumes,omitempty"`
		Produces            []string                         `json:"produces,omitempty"`
		Paths               map[string]*Path                 `json:"paths"`
		Definitions         map[string]*genschema.JSONSchema `json:"definitions,omitempty"`
		Parameters          map[string]*Parameter            `json:"parameters,omitempty"`
		Responses           map[string]*Response             `json:"responses,omitempty"`
		SecurityDefinitions map[string]*SecurityDefinition   `json:"securityDefinitions,omitempty"`
		Security            []map[string][]string            `json:"security,omitempty"`
		Tags                []Tag                            `json:"tags,omitempty"`
		ExternalDocs        *ExternalDocs                    `json:"externalDocs,omitempty"`
	}

	// Info provides metadata about the API. The metadata can be used by the clients if needed,
	// and can be presented in the Swagger-UI for convenience.
	Info struct {
		Title          string                    `json:"title,omitempty"`
		Description    string                    `json:"description,omitempty"`
		TermsOfService string                    `json:"termsOfService,omitempty"`
		Contact        *design.ContactDefinition `json:"contact,omitempty"`
		License        *design.LicenseDefinition `json:"license,omitempty"`
		Version        string                    `json:"version"`
	}

	// Path holds the relative paths to the individual endpoints.
	Path struct {
		// Ref allows for an external definition of this path item.
		Ref string `json:"$ref,omitempty"`
		// Get defines a GET operation on this path.
		Get *Operation `json:"get,omitempty"`
		// Put defines a PUT operation on this path.
		Put *Operation `json:"put,omitempty"`
		// Post defines a POST operation on this path.
		Post *Operation `json:"post,omitempty"`
		// Delete defines a DELETE operation on this path.
		Delete *Operation `json:"delete,omitempty"`
		// Options defines a OPTIONS operation on this path.
		Options *Operation `json:"options,omitempty"`
		// Head defines a HEAD operation on this path.
		Head *Operation `json:"head,omitempty"`
		// Patch defines a PATCH operation on this path.
		Patch *Operation `json:"patch,omitempty"`
		// Parameters is the list of parameters that are applicable for all the operations
		// described under this path.
		Parameters []*Parameter `parameters:"get,omitempty"`
	}

	// Operation describes a single API operation on a path.
	Operation struct {
		// Tags is a list of tags for API documentation control. Tags can be used for
		// logical grouping of operations by resources or any other qualifier.
		Tags []string `json:"tags,omitempty"`
		// Summary is a short summary of what the operation does. For maximum readability
		// in the swagger-ui, this field should be less than 120 characters.
		Summary string `json:"summary,omitempty"`
		// Description is a verbose explanation of the operation behavior.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// ExternalDocs points to additional external documentation for this operation.
		ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
		// OperationID is a unique string used to identify the operation.
		OperationID string `json:"operationId,omitempty"`
		// Consumes is a list of MIME types the operation can consume.
		Consumes []string `json:"consumes,omitempty"`
		// Produces is a list of MIME types the operation can produce.
		Produces []string `json:"produces,omitempty"`
		// Parameters is a list of parameters that are applicable for this operation.
		Parameters []*Parameter `json:"parameters,omitempty"`
		// Responses is the list of possible responses as they are returned from executing
		// this operation.
		Responses map[string]*Response `json:"responses,omitempty"`
		// Schemes is the transfer protocol for the operation.
		Schemes []string `json:"schemes,omitempty"`
		// Deprecated declares this operation to be deprecated.
		Deprecated bool `json:"deprecated,omitempty"`
		// Secury is a declaration of which security schemes are applied for this operation.
		Security []map[string][]string `json:"security,omitempty"`
	}

	// Parameter describes a single operation parameter.
	Parameter struct {
		// Name of the parameter. Parameter names are case sensitive.
		Name string `json:"name"`
		// In is the location of the parameter.
		// Possible values are "query", "header", "path", "formData" or "body".
		In string `json:"in"`
		// Description is`a brief description of the parameter.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// Required determines whether this parameter is mandatory.
		Required bool `json:"required"`
		// Schema defining the type used for the body parameter, only if "in" is body

		Schema *genschema.JSONSchema `json:"schema,omitempty"`
		// properties below only apply if "in" is not body

		//  Type of the parameter. Since the parameter is not located at the request body,
		// it is limited to simple types (that is, not an object).
		Type string `json:"type,omitempty"`
		// Format is the extending format for the previously mentioned type.
		Format string `json:"format,omitempty"`
		// AllowEmptyValue sets the ability to pass empty-valued parameters.
		// This is valid only for either query or formData parameters and allows you to
		// send a parameter with a name only or an empty value. Default value is false.
		AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
		// Items describes the type of items in the array if type is "array".
		Items *Items `json:"items,omitempty"`
		// CollectionFormat determines the format of the array if type array is used.
		// Possible values are csv, ssv, tsv, pipes and multi.
		CollectionFormat string `json:"collectionFormat,omitempty"`
		// Default declares the value of the parameter that the server will use if none is
		// provided, for example a "count" to control the number of results per page might
		// default to 100 if not supplied by the client in the request.
		Default          interface{}   `json:"default,omitempty"`
		Maximum          float64       `json:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
		Minimum          float64       `json:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
		MaxLength        int           `json:"maxLength,omitempty"`
		MinLength        int           `json:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty"`
		MaxItems         int           `json:"maxItems,omitempty"`
		MinItems         int           `json:"minItems,omitempty"`
		UniqueItems      bool          `json:"uniqueItems,omitempty"`
		Enum             []interface{} `json:"enum,omitempty"`
		MultipleOf       float64       `json:"multipleOf,omitempty"`
	}

	// Response describes an operation response.
	Response struct {
		// Description of the response. GFM syntax can be used for rich text representation.
		Description string `json:"description"`
		// Schema is a definition of the response structure. It can be a primitive,
		// an array or an object. If this field does not exist, it means no content is
		// returned as part of the response. As an extension to the Schema Object, its root
		// type value may also be "file".
		Schema *genschema.JSONSchema `json:"schema,omitempty"`
		// Headers is a list of headers that are sent with the response.
		Headers map[string]*Header `json:"headers,omitempty"`
	}

	// Header represents a header parameter.
	Header struct {
		// Description is`a brief description of the parameter.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		//  Type of the header. it is limited to simple types (that is, not an object).
		Type string `json:"type,omitempty"`
		// Format is the extending format for the previously mentioned type.
		Format string `json:"format,omitempty"`
		// Items describes the type of items in the array if type is "array".
		Items *Items `json:"items,omitempty"`
		// CollectionFormat determines the format of the array if type array is used.
		// Possible values are csv, ssv, tsv, pipes and multi.
		CollectionFormat string `json:"collectionFormat,omitempty"`
		// Default declares the value of the parameter that the server will use if none is
		// provided, for example a "count" to control the number of results per page might
		// default to 100 if not supplied by the client in the request.
		Default          interface{}   `json:"default,omitempty"`
		Maximum          float64       `json:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
		Minimum          float64       `json:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
		MaxLength        int           `json:"maxLength,omitempty"`
		MinLength        int           `json:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty"`
		MaxItems         int           `json:"maxItems,omitempty"`
		MinItems         int           `json:"minItems,omitempty"`
		UniqueItems      bool          `json:"uniqueItems,omitempty"`
		Enum             []interface{} `json:"enum,omitempty"`
		MultipleOf       float64       `json:"multipleOf,omitempty"`
	}

	// SecurityDefinition allows the definition of a security scheme that can be used by the
	// operations. Supported schemes are basic authentication, an API key (either as a header or
	// as a query parameter) and OAuth2's common flows (implicit, password, application and
	// access code).
	SecurityDefinition struct {
		// Type of the security scheme. Valid values are "basic", "apiKey" or "oauth2".
		Type string `json:"type"`
		// Description for security scheme
		Description string `json:"description,omitempty"`
		// Name of the header or query parameter to be used when type is "apiKey".
		Name string `json:"name,omitempty"`
		// In is the location of the API key when type is "apiKey".
		// Valid values are "query" or "header".
		In string `json:"in"`
		// Flow is the flow used by the OAuth2 security scheme when type is "oauth2"
		// Valid values are "implicit", "password", "application" or "accessCode".
		Flow string `json:"flow,omitempty"`
		// The oauth2 authorization URL to be used for this flow.
		AuthorizationURL string `json:"authorizationUrl,omitempty"`
		// TokenURL  is the token URL to be used for this flow.
		TokenURL string `json:"tokenUrl,omitempty"`
		// Scopes list the  available scopes for the OAuth2 security scheme.
		Scopes map[string]*Scope `json:"scopes,omitempty"`
	}

	// Scope corresponds to an available scope for an OAuth2 security scheme.
	Scope struct {
		// Description for scope
		Description string `json:"description,omitempty"`
	}

	// ExternalDocs allows referencing an external resource for extended documentation.
	ExternalDocs struct {
		// Description is a short description of the target documentation.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// URL for the target documentation.
		URL string `json:"url"`
	}

	// Items is a limited subset of JSON-Schema's items object. It is used by parameter
	// definitions that are not located in "body".
	Items struct {
		//  Type of the items. it is limited to simple types (that is, not an object).
		Type string `json:"type,omitempty"`
		// Format is the extending format for the previously mentioned type.
		Format string `json:"format,omitempty"`
		// Items describes the type of items in the array if type is "array".
		Items *Items `json:"items,omitempty"`
		// CollectionFormat determines the format of the array if type array is used.
		// Possible values are csv, ssv, tsv, pipes and multi.
		CollectionFormat string `json:"collectionFormat,omitempty"`
		// Default declares the value of the parameter that the server will use if none is
		// provided, for example a "count" to control the number of results per page might
		// default to 100 if not supplied by the client in the request.
		Default          interface{}   `json:"default,omitempty"`
		Maximum          float64       `json:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
		Minimum          float64       `json:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
		MaxLength        int           `json:"maxLength,omitempty"`
		MinLength        int           `json:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty"`
		MaxItems         int           `json:"maxItems,omitempty"`
		MinItems         int           `json:"minItems,omitempty"`
		UniqueItems      bool          `json:"uniqueItems,omitempty"`
		Enum             []interface{} `json:"enum,omitempty"`
		MultipleOf       float64       `json:"multipleOf,omitempty"`
	}

	// Tag allows adding meta data to a single tag that is used by the Operation Object. It is
	// not mandatory to have a Tag Object per tag used there.
	Tag struct {
		// Name of the tag.
		Name string `json:"name,omitempty"`
		// Description is a short description of the tag.
		// GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// ExternalDocs is additional external documentation for this tag.
		ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
	}
)

// New creates a Swagger spec from an API definition.
func New(api *design.APIDefinition) (*Swagger, error) {
	params, err := paramsFromDefinition(api.BaseParams, api.BasePath)
	if err != nil {
		return nil, err
	}
	var paramMap map[string]*Parameter
	if len(params) > 0 {
		paramMap = make(map[string]*Parameter, len(params))
		for _, p := range params {
			paramMap[p.Name] = p
		}
	}
	s := &Swagger{
		Swagger: "2.0",
		Info: &Info{
			Title:          api.Title,
			Description:    api.Description,
			TermsOfService: api.TermsOfService,
			Contact:        api.Contact,
			License:        api.License,
			Version:        "",
		},
		Host:         api.Host,
		BasePath:     api.BasePath,
		Paths:        make(map[string]*Path),
		Schemes:      []string{"https"},
		Consumes:     []string{"application/json"},
		Produces:     []string{"application/json"},
		Parameters:   paramMap,
		ExternalDocs: docsFromDefinition(api.Docs),
	}
	err = api.IterateResponses(func(r *design.ResponseDefinition) error {
		res, err := responseFromDefinition(api, r)
		if err != nil {
			return err
		}
		if s.Responses == nil {
			s.Responses = make(map[string]*Response)
		}
		s.Responses[r.Name] = res
		return nil
	})
	if err != nil {
		return nil, err
	}
	err = api.IterateResources(func(res *design.ResourceDefinition) error {
		return res.IterateActions(func(a *design.ActionDefinition) error {
			for _, route := range a.Routes {
				path, err := pathFromDefinition(api, route)
				if err != nil {
					return err
				}
				s.Paths[route.Path] = path
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	if len(genschema.Definitions) > 0 {
		s.Definitions = make(map[string]*genschema.JSONSchema)
		for n, d := range genschema.Definitions {
			// sad but swagger doesn't support these
			d.Media = nil
			d.Links = nil
			s.Definitions[n] = d
		}
	}
	return s, nil
}

func paramsFromDefinition(params *design.AttributeDefinition, path string) ([]*Parameter, error) {
	if params == nil {
		return nil, nil
	}
	obj := params.Type.ToObject()
	if obj == nil {
		return nil, fmt.Errorf("invalid parameters definition, not an object")
	}
	res := make([]*Parameter, len(obj))
	i := 0
	wildcards := design.ExtractWildcards(path)
	obj.IterateAttributes(func(n string, at *design.AttributeDefinition) error {
		in := "query"
		required := params.IsRequired(n)
		for _, w := range wildcards {
			if n == w {
				in = "path"
				required = true
				break
			}
		}
		param := &Parameter{
			Name:        n,
			Default:     at.DefaultValue,
			Description: at.Description,
			Required:    required,
			In:          in,
			Type:        at.Type.Name(),
		}
		var items *Items
		if at.Type.IsArray() {
			items = itemsFromDefinition(at)
		}
		param.Items = items
		initValidations(at, param)
		res[i] = param
		i++
		return nil
	})
	return res, nil
}

func itemsFromDefinition(at *design.AttributeDefinition) *Items {
	items := &Items{Type: at.Type.Name()}
	initValidations(at, items)
	if at.Type.IsArray() {
		items.Items = itemsFromDefinition(at.Type.ToArray().ElemType)
	}
	return items
}

func responseFromDefinition(api *design.APIDefinition, r *design.ResponseDefinition) (*Response, error) {
	var schema *genschema.JSONSchema
	if r.MediaType != "" {
		if mt, ok := api.MediaTypes[r.MediaType]; ok {
			schema = genschema.TypeSchema(api, mt)
		}
	}
	headers, err := headersFromDefinition(r.Headers)
	if err != nil {
		return nil, err
	}
	return &Response{
		Description: r.Description,
		Schema:      schema,
		Headers:     headers,
	}, nil
}

func headersFromDefinition(headers *design.AttributeDefinition) (map[string]*Header, error) {
	if headers == nil {
		return nil, nil
	}
	obj := headers.Type.ToObject()
	if obj == nil {
		return nil, fmt.Errorf("invalid headers definition, not an object")
	}
	res := make(map[string]*Header)
	obj.IterateAttributes(func(n string, at *design.AttributeDefinition) error {
		header := &Header{
			Default:     at.DefaultValue,
			Description: at.Description,
			Type:        at.Type.Name(),
		}
		initValidations(at, header)
		res[n] = header
		return nil
	})
	return res, nil
}

func pathFromDefinition(api *design.APIDefinition, route *design.RouteDefinition) (*Path, error) {
	action := route.Parent
	params, err := paramsFromDefinition(action.Params, route.Path)
	if err != nil {
		return nil, err
	}
	responses := make(map[string]*Response, len(action.Responses))
	for _, r := range action.Responses {
		resp, err := responseFromDefinition(api, r)
		if err != nil {
			return nil, err
		}
		responses[strconv.Itoa(r.Status)] = resp
	}
	operation := &Operation{
		Description:  action.Description,
		ExternalDocs: docsFromDefinition(action.Docs),
		OperationID:  fmt.Sprintf("%s#%s", action.Parent.Name, action.Name),
		Consumes:     []string{"application/json"},
		Produces:     []string{"application/json"},
		Parameters:   params,
		Responses:    responses,
		Schemes:      []string{"https"},
		Deprecated:   false,
	}
	var path *Path
	switch route.Verb {
	case "GET":
		path = &Path{Get: operation}
	case "PUT":
		path = &Path{Put: operation}
	case "POST":
		path = &Path{Post: operation}
	case "DELETE":
		path = &Path{Delete: operation}
	case "OPTIONS":
		path = &Path{Options: operation}
	case "HEAD":
		path = &Path{Head: operation}
	case "PATCH":
		path = &Path{Patch: operation}
	}
	return path, nil
}

func docsFromDefinition(docs *design.DocsDefinition) *ExternalDocs {
	if docs == nil {
		return nil
	}
	return &ExternalDocs{
		Description: docs.Description,
		URL:         docs.URL,
	}
}

func initValidations(attr *design.AttributeDefinition, def interface{}) {
	for _, v := range attr.Validations {
		switch val := v.(type) {
		case *design.EnumValidationDefinition:
			switch actual := def.(type) {
			case *Parameter:
				actual.Enum = val.Values
			case *Header:
				actual.Enum = val.Values
			case *Items:
				actual.Enum = val.Values
			}
		case *design.FormatValidationDefinition:
			switch actual := def.(type) {
			case *Parameter:
				actual.Format = val.Format
			case *Header:
				actual.Format = val.Format
			case *Items:
				actual.Format = val.Format
			}
		case *design.PatternValidationDefinition:
			switch actual := def.(type) {
			case *Parameter:
				actual.Pattern = val.Pattern
			case *Header:
				actual.Pattern = val.Pattern
			case *Items:
				actual.Pattern = val.Pattern
			}
		case *design.MinimumValidationDefinition:
			switch actual := def.(type) {
			case *Parameter:
				actual.Minimum = val.Min
				actual.ExclusiveMinimum = true
			case *Header:
				actual.Minimum = val.Min
				actual.ExclusiveMinimum = true
			case *Items:
				actual.Minimum = val.Min
				actual.ExclusiveMinimum = true
			}
		case *design.MaximumValidationDefinition:
			switch actual := def.(type) {
			case *Parameter:
				actual.Maximum = val.Max
				actual.ExclusiveMaximum = true
			case *Header:
				actual.Maximum = val.Max
				actual.ExclusiveMaximum = true
			case *Items:
				actual.Maximum = val.Max
				actual.ExclusiveMaximum = true
			}
		case *design.MinLengthValidationDefinition:
			switch actual := def.(type) {
			case *Parameter:
				actual.MinLength = val.MinLength
			case *Header:
				actual.MinLength = val.MinLength
			case *Items:
				actual.MinLength = val.MinLength
			}
		case *design.MaxLengthValidationDefinition:
			switch actual := def.(type) {
			case *Parameter:
				actual.MaxLength = val.MaxLength
			case *Header:
				actual.MaxLength = val.MaxLength
			case *Items:
				actual.MaxLength = val.MaxLength
			}
		}
	}
}
