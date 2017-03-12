package v2

// New creates a OpenAPI spec from a REST root expression.
import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"goa.design/goa.v2/design"
	rest "goa.design/goa.v2/rest/design"
)

// New returns the OpenAPI v2 specification for the given API.
func New(api *design.APIExpr, r *rest.RootExpr) (*OpenAPI, error) {
	if r == nil {
		return nil, nil
	}
	tags := tagsFromExpr(r.Metadata)
	basePath := r.BasePath
	if hasAbsoluteRoutes(r) {
		basePath = ""
	}
	params, err := paramsFromExpr(r.Params(), basePath)
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
	s := &OpenAPI{
		Swagger: "2.0",
		Info: &Info{
			Title:          api.Title,
			Description:    api.Description,
			TermsOfService: api.TermsOfService,
			Contact:        api.Contact,
			License:        api.License,
			Version:        api.Version,
			Extensions:     extensionsFromExpr(r.Metadata),
		},
		Host:         api.Host,
		BasePath:     basePath,
		Paths:        make(map[string]interface{}),
		Consumes:     consumes,
		Produces:     produces,
		Parameters:   paramMap,
		Tags:         tags,
		ExternalDocs: docsFromExpr(api.Docs),
	}

	for _, he := range r.HTTPErrors {
		res, err := responseSpecFromExpr(s, r, he.Response)
		if err != nil {
			return err
		}
		if s.Responses == nil {
			s.Responses = make(map[string]*Response)
		}
		s.Responses[r.Name] = res
	}
	if err != nil {
		return nil, err
	}
	for _, res := range r.Resources {
		for k, v := range extensionsFromExpr(res.Metadata) {
			s.Paths[k] = v
		}
		for _, fs := range res.FileServers {
			if !mustGenerate(fs.Metadata) {
				return nil
			}
			return buildPathFromFileServer(s, api, fs)
		}
		if err != nil {
			return err
		}
		for _, a := range res.Actions {
			if !mustGenerate(a.Metadata) {
				return nil
			}
			for _, route := range a.Routes {
				if err := buildPathFromExpr(s, api, route, basePath); err != nil {
					return err
				}
			}
		}
	}
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

// mustGenerate returns true if the metadata indicates that a OpenAPI specification should be
// generated, false otherwise.
func mustGenerate(meta design.MetadataExpr) bool {
	if m, ok := meta["swagger:generate"]; ok {
		if len(m) > 0 && m[0] == "false" {
			return false
		}
	}
	return true
}

// hasAbsoluteRoutes returns true if any action exposed by the API uses an absolute route of if the
// API has file servers. This is needed as OpenAPI does not support exceptions to the base path so
// if the API has any absolute route the base path must be "/" and all routes must be absolutes.
func hasAbsoluteRoutes(api *design.APIExpr) bool {
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

func tagsFromExpr(mdata design.MetadataExpr) (tags []*Tag) {
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

		tag.Extensions = extensionsFromExpr(mdata)

		tags = append(tags, tag)
	}

	return
}

func tagNamesFromExpr(mdatas ...design.MetadataExpr) (tagNames []string) {
	for _, mdata := range mdatas {
		tags := tagsFromExpr(mdata)
		for _, tag := range tags {
			tagNames = append(tagNames, tag.Name)
		}
	}
	return
}

func summaryFromExpr(name string, metadata design.MetadataExpr) string {
	for n, mdata := range metadata {
		if n == "swagger:summary" && len(mdata) > 0 {
			return mdata[0]
		}
	}
	return name
}

func extensionsFromExpr(mdata design.MetadataExpr) map[string]interface{} {
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

func paramsFromExpr(params *design.AttributeExpr, path string) ([]*Parameter, error) {
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
	obj.IterateAttributes(func(n string, at *design.AttributeExpr) error {
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

func paramsFromHeaders(action *rest.ActionExpr) []*Parameter {
	params := []*Parameter{}
	action.IterateHeaders(func(name string, required bool, header *design.AttributeExpr) error {
		p := paramFor(header, name, "header", required)
		params = append(params, p)
		return nil
	})
	return params
}

func paramFor(at *design.AttributeExpr, name, in string, required bool) *Parameter {
	p := &Parameter{
		In:          in,
		Name:        name,
		Default:     toStringMap(at.DefaultValue),
		Description: at.Description,
		Required:    required,
		Type:        at.Type.Name(),
	}
	if at.Type.IsArray() {
		p.Items = itemsFromExpr(at.Type.ToArray().ElemType)
		p.CollectionFormat = "multi"
	}
	p.Extensions = extensionsFromExpr(at.Metadata)
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

func itemsFromExpr(at *design.AttributeExpr) *Items {
	items := &Items{Type: at.Type.Name()}
	initValidations(at, items)
	if at.Type.IsArray() {
		items.Items = itemsFromExpr(at.Type.ToArray().ElemType)
	}
	return items
}

func responseSpecFromExpr(s *OpenAPI, api *design.APIExpr, r *rest.HTTPResponseExpr) (*Response, error) {
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
	headers, err := headersFromExpr(r.Headers)
	if err != nil {
		return nil, err
	}
	return &Response{
		Description: r.Description,
		Schema:      schema,
		Headers:     headers,
		Extensions:  extensionsFromExpr(r.Metadata),
	}, nil
}

func responseFromExpr(s *OpenAPI, api *design.APIExpr, r *rest.HTTPResponseExpr) (*Response, error) {
	var (
		response *Response
		err      error
	)
	response, err = responseSpecFromExpr(s, api, r)
	if err != nil {
		return nil, err
	}
	if r.Standard {
		if s.Responses == nil {
			s.Responses = make(map[string]*Response)
		}
		if _, ok := s.Responses[r.Name]; !ok {
			sp, err := responseSpecFromExpr(s, api, r)
			if err != nil {
				return nil, err
			}
			s.Responses[r.Name] = sp
		}
	}
	return response, nil
}

func headersFromExpr(headers *design.AttributeExpr) (map[string]*Header, error) {
	if headers == nil {
		return nil, nil
	}
	obj := headers.Type.ToObject()
	if obj == nil {
		return nil, fmt.Errorf("invalid headers definition, not an object")
	}
	res := make(map[string]*Header)
	obj.IterateAttributes(func(n string, at *design.AttributeExpr) error {
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

func buildPathFromFileServer(s *OpenAPI, api *design.APIAPIDefinition, fs *design.FileServerDefinition) error {
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
		Summary:      summaryFromExpr(fmt.Sprintf("Download %s", fs.FilePath), fs.Metadata),
		ExternalDocs: docsFromExpr(fs.Docs),
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
	p.Extensions = extensionsFromExpr(fs.Metadata)

	return nil
}

func buildPathFromExpr(s *OpenAPI, api *design.APIExpr, route *design.RouteDefinition, basePath string) error {
	action := route.Parent

	tagNames := tagNamesFromExpr(action.Parent.Metadata, action.Metadata)
	if len(tagNames) == 0 {
		// By default tag with resource name
		tagNames = []string{route.Parent.Parent.Name}
	}
	params, err := paramsFromExpr(action.AllParams(), route.FullPath())
	if err != nil {
		return err
	}

	params = append(params, paramsFromHeaders(action)...)

	responses := make(map[string]*Response, len(action.Responses))
	for _, r := range action.Responses {
		resp, err := responseFromExpr(s, api, r)
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
		Summary:      summaryFromExpr(action.Name+" "+action.Parent.Name, action.Metadata),
		ExternalDocs: docsFromExpr(action.Docs),
		OperationID:  operationID,
		Parameters:   params,
		Responses:    responses,
		Schemes:      schemes,
		Deprecated:   false,
		Extensions:   extensionsFromExpr(route.Metadata),
	}

	applySecurity(operation, action.Security)

	key := design.WildcardRegex.ReplaceAllStringFunc(
		route.FullPath(),
		func(w string) string {
			return fmt.Sprintf("/{%s}", w[2:])
		},
	)
	if key == "" {
		key = "/"
	}
	bp := design.WildcardRegex.ReplaceAllStringFunc(
		basePath,
		func(w string) string {
			return fmt.Sprintf("/{%s}", w[2:])
		},
	)
	if bp != "/" {
		key = strings.TrimPrefix(key, bp)
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
	p.Extensions = extensionsFromExpr(route.Parent.Metadata)
	return nil
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

func docsFromExpr(docs *design.DocsDefinition) *ExternalDocs {
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

func initValidations(attr *design.AttributeExpr, def interface{}) {
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
