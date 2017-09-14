package genswagger

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/dslengine"
	"github.com/goadesign/goa/goagen/gen_schema"
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
		Paths               map[string]interface{}           `json:"paths"`
		Definitions         map[string]*genschema.JSONSchema `json:"definitions,omitempty"`
		Parameters          map[string]*Parameter            `json:"parameters,omitempty"`
		Responses           map[string]*Response             `json:"responses,omitempty"`
		SecurityDefinitions map[string]*SecurityDefinition   `json:"securityDefinitions,omitempty"`
		Tags                []*Tag                           `json:"tags,omitempty"`
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
		Extensions     map[string]interface{}    `json:"-"`
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
		Parameters []*Parameter `json:"parameters,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-"`
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
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-"`
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
		Maximum          *float64      `json:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
		Minimum          *float64      `json:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
		MaxLength        *int          `json:"maxLength,omitempty"`
		MinLength        *int          `json:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty"`
		MaxItems         *int          `json:"maxItems,omitempty"`
		MinItems         *int          `json:"minItems,omitempty"`
		UniqueItems      bool          `json:"uniqueItems,omitempty"`
		Enum             []interface{} `json:"enum,omitempty"`
		MultipleOf       float64       `json:"multipleOf,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-"`
	}

	// Response describes an operation response.
	Response struct {
		// Description of the response. GFM syntax can be used for rich text representation.
		Description string `json:"description,omitempty"`
		// Schema is a definition of the response structure. It can be a primitive,
		// an array or an object. If this field does not exist, it means no content is
		// returned as part of the response. As an extension to the Schema Object, its root
		// type value may also be "file".
		Schema *genschema.JSONSchema `json:"schema,omitempty"`
		// Headers is a list of headers that are sent with the response.
		Headers map[string]*Header `json:"headers,omitempty"`
		// Ref references a global API response.
		// This field is exclusive with the other fields of Response.
		Ref string `json:"$ref,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-"`
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
		Maximum          *float64      `json:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
		Minimum          *float64      `json:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
		MaxLength        *int          `json:"maxLength,omitempty"`
		MinLength        *int          `json:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty"`
		MaxItems         *int          `json:"maxItems,omitempty"`
		MinItems         *int          `json:"minItems,omitempty"`
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
		In string `json:"in,omitempty"`
		// Flow is the flow used by the OAuth2 security scheme when type is "oauth2"
		// Valid values are "implicit", "password", "application" or "accessCode".
		Flow string `json:"flow,omitempty"`
		// The oauth2 authorization URL to be used for this flow.
		AuthorizationURL string `json:"authorizationUrl,omitempty"`
		// TokenURL  is the token URL to be used for this flow.
		TokenURL string `json:"tokenUrl,omitempty"`
		// Scopes list the  available scopes for the OAuth2 security scheme.
		Scopes map[string]string `json:"scopes,omitempty"`
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-"`
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
		Maximum          *float64      `json:"maximum,omitempty"`
		ExclusiveMaximum bool          `json:"exclusiveMaximum,omitempty"`
		Minimum          *float64      `json:"minimum,omitempty"`
		ExclusiveMinimum bool          `json:"exclusiveMinimum,omitempty"`
		MaxLength        *int          `json:"maxLength,omitempty"`
		MinLength        *int          `json:"minLength,omitempty"`
		Pattern          string        `json:"pattern,omitempty"`
		MaxItems         *int          `json:"maxItems,omitempty"`
		MinItems         *int          `json:"minItems,omitempty"`
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
		// Extensions defines the swagger extensions.
		Extensions map[string]interface{} `json:"-"`
	}

	// These types are used in marshalJSON() to avoid recursive call of json.Marshal().
	_Info               Info
	_Path               Path
	_Operation          Operation
	_Parameter          Parameter
	_Response           Response
	_SecurityDefinition SecurityDefinition
	_Tag                Tag
)

func marshalJSON(v interface{}, extensions map[string]interface{}) ([]byte, error) {
	marshaled, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if len(extensions) == 0 {
		return marshaled, nil
	}
	var unmarshaled interface{}
	if err := json.Unmarshal(marshaled, &unmarshaled); err != nil {
		return nil, err
	}
	asserted := unmarshaled.(map[string]interface{})
	for k, v := range extensions {
		asserted[k] = v
	}
	merged, err := json.Marshal(asserted)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

// MarshalJSON returns the JSON encoding of i.
func (i Info) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Info(i), i.Extensions)
}

// MarshalJSON returns the JSON encoding of p.
func (p Path) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Path(p), p.Extensions)
}

// MarshalJSON returns the JSON encoding of o.
func (o Operation) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Operation(o), o.Extensions)
}

// MarshalJSON returns the JSON encoding of p.
func (p Parameter) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Parameter(p), p.Extensions)
}

// MarshalJSON returns the JSON encoding of r.
func (r Response) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Response(r), r.Extensions)
}

// MarshalJSON returns the JSON encoding of s.
func (s SecurityDefinition) MarshalJSON() ([]byte, error) {
	return marshalJSON(_SecurityDefinition(s), s.Extensions)
}

// MarshalJSON returns the JSON encoding of t.
func (t Tag) MarshalJSON() ([]byte, error) {
	return marshalJSON(_Tag(t), t.Extensions)
}

// New creates a Swagger spec from an API definition.
func New(api *design.APIDefinition) (*Swagger, error) {
	if api == nil {
		return nil, nil
	}
	tags := tagsFromDefinition(api.Metadata)
	basePath := api.BasePath
	if hasAbsoluteRoutes(api) {
		basePath = ""
	}
	params, err := paramsFromDefinition(api.Params, basePath)
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
	var consumes []string
	for _, c := range api.Consumes {
		consumes = append(consumes, c.MIMETypes...)
	}
	var produces []string
	for _, p := range api.Produces {
		produces = append(produces, p.MIMETypes...)
	}
	s := &Swagger{
		Swagger: "2.0",
		Info: &Info{
			Title:          api.Title,
			Description:    api.Description,
			TermsOfService: api.TermsOfService,
			Contact:        api.Contact,
			License:        api.License,
			Version:        api.Version,
			Extensions:     extensionsFromDefinition(api.Metadata),
		},
		Host:                api.Host,
		BasePath:            basePath,
		Paths:               make(map[string]interface{}),
		Schemes:             api.Schemes,
		Consumes:            consumes,
		Produces:            produces,
		Parameters:          paramMap,
		Tags:                tags,
		ExternalDocs:        docsFromDefinition(api.Docs),
		SecurityDefinitions: securityDefsFromDefinition(api.SecuritySchemes),
	}

	err = api.IterateResponses(func(r *design.ResponseDefinition) error {
		res, err := responseSpecFromDefinition(s, api, r)
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
		for k, v := range extensionsFromDefinition(res.Metadata) {
			s.Paths[k] = v
		}
		err := res.IterateFileServers(func(fs *design.FileServerDefinition) error {
			if !mustGenerate(fs.Metadata) {
				return nil
			}
			return buildPathFromFileServer(s, api, fs)
		})
		if err != nil {
			return err
		}
		return res.IterateActions(func(a *design.ActionDefinition) error {
			if !mustGenerate(a.Metadata) {
				return nil
			}
			for _, route := range a.Routes {
				if err := buildPathFromDefinition(s, api, route, basePath); err != nil {
					return err
				}
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

// mustGenerate returns true if the metadata indicates that a Swagger specification should be
// generated, false otherwise.
func mustGenerate(meta dslengine.MetadataDefinition) bool {
	if m, ok := meta["swagger:generate"]; ok {
		if len(m) > 0 && m[0] == "false" {
			return false
		}
	}
	return true
}

// hasAbsoluteRoutes returns true if any action exposed by the API uses an absolute route of if the
// API has file servers. This is needed as Swagger does not support exceptions to the base path so
// if the API has any absolute route the base path must be "/" and all routes must be absolutes.
func hasAbsoluteRoutes(api *design.APIDefinition) bool {
	hasAbsoluteRoutes := false
	for _, res := range api.Resources {
		for _, fs := range res.FileServers {
			if !mustGenerate(fs.Metadata) {
				continue
			}
			hasAbsoluteRoutes = true
			break
		}
		for _, a := range res.Actions {
			if !mustGenerate(a.Metadata) {
				continue
			}
			for _, ro := range a.Routes {
				if ro.IsAbsolute() {
					hasAbsoluteRoutes = true
					break
				}
			}
			if hasAbsoluteRoutes {
				break
			}
		}
		if hasAbsoluteRoutes {
			break
		}
	}
	return hasAbsoluteRoutes
}

func securityDefsFromDefinition(schemes []*design.SecuritySchemeDefinition) map[string]*SecurityDefinition {
	if len(schemes) == 0 {
		return nil
	}

	defs := make(map[string]*SecurityDefinition)
	for _, scheme := range schemes {
		def := &SecurityDefinition{
			Type:             scheme.Type,
			Description:      scheme.Description,
			Name:             scheme.Name,
			In:               scheme.In,
			Flow:             scheme.Flow,
			AuthorizationURL: scheme.AuthorizationURL,
			TokenURL:         scheme.TokenURL,
			Scopes:           scheme.Scopes,
			Extensions:       extensionsFromDefinition(scheme.Metadata),
		}
		if scheme.Kind == design.JWTSecurityKind {
			if def.TokenURL != "" {
				def.Description += fmt.Sprintf("\n\n**Token URL**: %s", def.TokenURL)
				def.TokenURL = ""
			}
			if len(def.Scopes) != 0 {
				def.Description += fmt.Sprintf("\n\n**Security Scopes**:\n%s", scopesMapList(def.Scopes))
				def.Scopes = nil
			}
		}
		defs[scheme.SchemeName] = def
	}
	return defs
}

func scopesMapList(scopes map[string]string) string {
	names := []string{}
	for name := range scopes {
		names = append(names, name)
	}
	sort.Strings(names)

	lines := []string{}
	for _, name := range names {
		lines = append(lines, fmt.Sprintf("  * `%s`: %s", name, scopes[name]))
	}
	return strings.Join(lines, "\n")
}

func tagsFromDefinition(mdata dslengine.MetadataDefinition) (tags []*Tag) {
	var keys []string
	for k := range mdata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		chunks := strings.Split(key, ":")
		if len(chunks) != 3 {
			continue
		}
		if chunks[0] != "swagger" || chunks[1] != "tag" {
			continue
		}

		tag := &Tag{Name: chunks[2]}

		mdata[key] = mdata[fmt.Sprintf("%s:desc", key)]
		if len(mdata[key]) != 0 {
			tag.Description = mdata[key][0]
		}

		hasDocs := false
		docs := &ExternalDocs{}

		mdata[key] = mdata[fmt.Sprintf("%s:url", key)]
		if len(mdata[key]) != 0 {
			docs.URL = mdata[key][0]
			hasDocs = true
		}

		mdata[key] = mdata[fmt.Sprintf("%s:url:desc", key)]
		if len(mdata[key]) != 0 {
			docs.Description = mdata[key][0]
			hasDocs = true
		}

		if hasDocs {
			tag.ExternalDocs = docs
		}

		tag.Extensions = extensionsFromDefinition(mdata)

		tags = append(tags, tag)
	}

	return
}

func tagNamesFromDefinitions(mdatas ...dslengine.MetadataDefinition) (tagNames []string) {
	for _, mdata := range mdatas {
		tags := tagsFromDefinition(mdata)
		for _, tag := range tags {
			tagNames = append(tagNames, tag.Name)
		}
	}
	return
}

func summaryFromDefinition(name string, metadata dslengine.MetadataDefinition) string {
	for n, mdata := range metadata {
		if n == "swagger:summary" && len(mdata) > 0 {
			return mdata[0]
		}
	}
	return name
}

func extensionsFromDefinition(mdata dslengine.MetadataDefinition) map[string]interface{} {
	extensions := make(map[string]interface{})
	for key, value := range mdata {
		chunks := strings.Split(key, ":")
		if len(chunks) != 3 {
			continue
		}
		if chunks[0] != "swagger" || chunks[1] != "extension" {
			continue
		}
		if strings.HasPrefix(chunks[2], "x-") != true {
			continue
		}
		val := value[0]
		ival := interface{}(val)
		if err := json.Unmarshal([]byte(val), &ival); err != nil {
			extensions[chunks[2]] = val
			continue
		}
		extensions[chunks[2]] = ival
	}
	if len(extensions) == 0 {
		return nil
	}
	return extensions
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
		param := paramFor(at, n, in, required)
		res[i] = param
		i++
		return nil
	})
	return res, nil
}

func paramsFromHeaders(action *design.ActionDefinition) []*Parameter {
	params := []*Parameter{}
	action.IterateHeaders(func(name string, required bool, header *design.AttributeDefinition) error {
		p := paramFor(header, name, "header", required)
		params = append(params, p)
		return nil
	})
	return params
}

func paramFor(at *design.AttributeDefinition, name, in string, required bool) *Parameter {
	p := &Parameter{
		In:          in,
		Name:        name,
		Default:     toStringMap(at.DefaultValue),
		Description: at.Description,
		Required:    required,
		Type:        at.Type.Name(),
	}
	if at.Type.IsArray() {
		p.Items = itemsFromDefinition(at.Type.ToArray().ElemType)
		p.CollectionFormat = "multi"
	}
	p.Extensions = extensionsFromDefinition(at.Metadata)
	initValidations(at, p)
	return p
}

// toStringMap converts map[interface{}]interface{} to a map[string]interface{} when possible.
func toStringMap(val interface{}) interface{} {
	switch actual := val.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for k, v := range actual {
			m[toString(k)] = toStringMap(v)
		}
		return m
	case []interface{}:
		mapSlice := make([]interface{}, len(actual))
		for i, e := range actual {
			mapSlice[i] = toStringMap(e)
		}
		return mapSlice
	default:
		return actual
	}
}

// toString returns the string representation of the given type.
func toString(val interface{}) string {
	switch actual := val.(type) {
	case string:
		return actual
	case int:
		return strconv.Itoa(actual)
	case float64:
		return strconv.FormatFloat(actual, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(actual)
	default:
		panic("unexpected key type")
	}
}

func itemsFromDefinition(at *design.AttributeDefinition) *Items {
	items := &Items{Type: at.Type.Name()}
	initValidations(at, items)
	if at.Type.IsArray() {
		items.Items = itemsFromDefinition(at.Type.ToArray().ElemType)
	}
	return items
}

func responseSpecFromDefinition(s *Swagger, api *design.APIDefinition, r *design.ResponseDefinition) (*Response, error) {
	var schema *genschema.JSONSchema
	if r.MediaType != "" {
		if mt, ok := api.MediaTypes[design.CanonicalIdentifier(r.MediaType)]; ok {
			view := r.ViewName
			if view == "" {
				view = design.DefaultView
			}
			schema = genschema.NewJSONSchema()
			schema.Ref = genschema.MediaTypeRef(api, mt, view)
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
		Extensions:  extensionsFromDefinition(r.Metadata),
	}, nil
}

func responseFromDefinition(s *Swagger, api *design.APIDefinition, r *design.ResponseDefinition) (*Response, error) {
	var (
		response *Response
		err      error
	)
	response, err = responseSpecFromDefinition(s, api, r)
	if err != nil {
		return nil, err
	}
	if r.Standard {
		if s.Responses == nil {
			s.Responses = make(map[string]*Response)
		}
		if _, ok := s.Responses[r.Name]; !ok {
			sp, err := responseSpecFromDefinition(s, api, r)
			if err != nil {
				return nil, err
			}
			s.Responses[r.Name] = sp
		}
	}
	return response, nil
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

func buildPathFromFileServer(s *Swagger, api *design.APIDefinition, fs *design.FileServerDefinition) error {
	wcs := design.ExtractWildcards(fs.RequestPath)
	var param []*Parameter
	if len(wcs) > 0 {
		param = []*Parameter{{
			In:          "path",
			Name:        wcs[0],
			Description: "Relative file path",
			Required:    true,
			Type:        "string",
		}}
	}

	responses := map[string]*Response{
		"200": {
			Description: "File downloaded",
			Schema:      &genschema.JSONSchema{Type: genschema.JSONFile},
		},
	}
	if len(wcs) > 0 {
		schema := genschema.TypeSchema(api, design.ErrorMedia)
		responses["404"] = &Response{Description: "File not found", Schema: schema}
	}

	operationID := fmt.Sprintf("%s#%s", fs.Parent.Name, fs.RequestPath)
	schemes := api.Schemes

	operation := &Operation{
		Description:  fs.Description,
		Summary:      summaryFromDefinition(fmt.Sprintf("Download %s", fs.FilePath), fs.Metadata),
		ExternalDocs: docsFromDefinition(fs.Docs),
		OperationID:  operationID,
		Parameters:   param,
		Responses:    responses,
		Schemes:      schemes,
	}

	applySecurity(operation, fs.Security)

	key := design.WildcardRegex.ReplaceAllStringFunc(
		fs.RequestPath,
		func(w string) string {
			return fmt.Sprintf("/{%s}", w[2:])
		},
	)
	if key == "" {
		key = "/"
	}
	var path interface{}
	var ok bool
	if path, ok = s.Paths[key]; !ok {
		path = new(Path)
		s.Paths[key] = path
	}
	p := path.(*Path)
	p.Get = operation
	p.Extensions = extensionsFromDefinition(fs.Metadata)

	return nil
}

func buildPathFromDefinition(s *Swagger, api *design.APIDefinition, route *design.RouteDefinition, basePath string) error {
	action := route.Parent

	tagNames := tagNamesFromDefinitions(action.Parent.Metadata, action.Metadata)
	if len(tagNames) == 0 {
		// By default tag with resource name
		tagNames = []string{route.Parent.Parent.Name}
	}
	params, err := paramsFromDefinition(action.AllParams(), route.FullPath())
	if err != nil {
		return err
	}

	params = append(params, paramsFromHeaders(action)...)

	responses := make(map[string]*Response, len(action.Responses))
	for _, r := range action.Responses {
		resp, err := responseFromDefinition(s, api, r)
		if err != nil {
			return err
		}
		responses[strconv.Itoa(r.Status)] = resp
	}

	if action.Payload != nil {
		payloadSchema := genschema.TypeSchema(api, action.Payload)
		pp := &Parameter{
			Name:        "payload",
			In:          "body",
			Description: action.Payload.Description,
			Required:    !action.PayloadOptional,
			Schema:      payloadSchema,
		}
		params = append(params, pp)
	}

	operationID := fmt.Sprintf("%s#%s", action.Parent.Name, action.Name)
	index := 0
	for i, rt := range action.Routes {
		if rt == route {
			index = i
			break
		}
	}
	if index > 0 {
		operationID = fmt.Sprintf("%s#%d", operationID, index)
	}

	schemes := action.Schemes
	if len(schemes) == 0 {
		schemes = api.Schemes
	}

	operation := &Operation{
		Tags:         tagNames,
		Description:  action.Description,
		Summary:      summaryFromDefinition(action.Name+" "+action.Parent.Name, action.Metadata),
		ExternalDocs: docsFromDefinition(action.Docs),
		OperationID:  operationID,
		Parameters:   params,
		Responses:    responses,
		Schemes:      schemes,
		Deprecated:   false,
		Extensions:   extensionsFromDefinition(route.Metadata),
	}

	computeProduces(operation, s, action)
	applySecurity(operation, action.Security)

	key := design.WildcardRegex.ReplaceAllStringFunc(
		route.FullPath(),
		func(w string) string {
			return fmt.Sprintf("/{%s}", w[2:])
		},
	)
	bp := design.WildcardRegex.ReplaceAllStringFunc(
		basePath,
		func(w string) string {
			return fmt.Sprintf("/{%s}", w[2:])
		},
	)
	if bp != "/" {
		key = strings.TrimPrefix(key, bp)
	}
	if key == "" {
		key = "/"
	}
	var path interface{}
	var ok bool
	if path, ok = s.Paths[key]; !ok {
		path = new(Path)
		s.Paths[key] = path
	}
	p := path.(*Path)
	switch route.Verb {
	case "GET":
		p.Get = operation
	case "PUT":
		p.Put = operation
	case "POST":
		p.Post = operation
	case "DELETE":
		p.Delete = operation
	case "OPTIONS":
		p.Options = operation
	case "HEAD":
		p.Head = operation
	case "PATCH":
		p.Patch = operation
	}
	p.Extensions = extensionsFromDefinition(route.Parent.Metadata)
	return nil
}

func computeProduces(operation *Operation, s *Swagger, action *design.ActionDefinition) {
	produces := make(map[string]struct{})
	action.IterateResponses(func(resp *design.ResponseDefinition) error {
		if resp.MediaType != "" {
			produces[resp.MediaType] = struct{}{}
		}
		return nil
	})
	subset := true
	for p := range produces {
		found := false
		for _, p2 := range s.Produces {
			if p == p2 {
				found = true
				break
			}
		}
		if !found {
			subset = false
			break
		}
	}
	if !subset {
		operation.Produces = make([]string, len(produces))
		i := 0
		for p := range produces {
			operation.Produces[i] = p
			i++
		}
		sort.Strings(operation.Produces)
	}
}

func applySecurity(operation *Operation, security *design.SecurityDefinition) {
	if security != nil && security.Scheme.Kind != design.NoSecurityKind {
		if security.Scheme.Kind == design.JWTSecurityKind && len(security.Scopes) > 0 {
			if operation.Description != "" {
				operation.Description += "\n\n"
			}
			operation.Description += fmt.Sprintf("Required security scopes:\n%s", scopesList(security.Scopes))
		}
		scopes := security.Scopes
		if scopes == nil {
			scopes = make([]string, 0)
		}
		sec := []map[string][]string{{security.Scheme.SchemeName: scopes}}
		operation.Security = sec
	}
}

func scopesList(scopes []string) string {
	sort.Strings(scopes)

	var lines []string
	for _, scope := range scopes {
		lines = append(lines, fmt.Sprintf("  * `%s`", scope))
	}
	return strings.Join(lines, "\n")
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

func initEnumValidation(def interface{}, values []interface{}) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Enum = values
	case *Header:
		actual.Enum = values
	case *Items:
		actual.Enum = values
	}
}

func initFormatValidation(def interface{}, format string) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Format = format
	case *Header:
		actual.Format = format
	case *Items:
		actual.Format = format
	}
}

func initPatternValidation(def interface{}, pattern string) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Pattern = pattern
	case *Header:
		actual.Pattern = pattern
	case *Items:
		actual.Pattern = pattern
	}
}

func initMinimumValidation(def interface{}, min *float64) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Minimum = min
		actual.ExclusiveMinimum = false
	case *Header:
		actual.Minimum = min
		actual.ExclusiveMinimum = false
	case *Items:
		actual.Minimum = min
		actual.ExclusiveMinimum = false
	}
}

func initMaximumValidation(def interface{}, max *float64) {
	switch actual := def.(type) {
	case *Parameter:
		actual.Maximum = max
		actual.ExclusiveMaximum = false
	case *Header:
		actual.Maximum = max
		actual.ExclusiveMaximum = false
	case *Items:
		actual.Maximum = max
		actual.ExclusiveMaximum = false
	}
}

func initMinLengthValidation(def interface{}, isArray bool, min *int) {
	switch actual := def.(type) {
	case *Parameter:
		if isArray {
			actual.MinItems = min
		} else {
			actual.MinLength = min
		}
	case *Header:
		actual.MinLength = min
	case *Items:
		actual.MinLength = min
	}
}

func initMaxLengthValidation(def interface{}, isArray bool, max *int) {
	switch actual := def.(type) {
	case *Parameter:
		if isArray {
			actual.MaxItems = max
		} else {
			actual.MaxLength = max
		}
	case *Header:
		actual.MaxLength = max
	case *Items:
		actual.MaxLength = max
	}
}

func initValidations(attr *design.AttributeDefinition, def interface{}) {
	val := attr.Validation
	if val == nil {
		return
	}
	initEnumValidation(def, val.Values)
	initFormatValidation(def, val.Format)
	initPatternValidation(def, val.Pattern)
	if val.Minimum != nil {
		initMinimumValidation(def, val.Minimum)
	}
	if val.Maximum != nil {
		initMaximumValidation(def, val.Maximum)
	}
	if val.MinLength != nil {
		initMinLengthValidation(def, attr.Type.IsArray(), val.MinLength)
	}
	if val.MaxLength != nil {
		initMaxLengthValidation(def, attr.Type.IsArray(), val.MaxLength)
	}
}
