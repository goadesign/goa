package openapi

// New creates a OpenAPI spec from a HTTP root expression.
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"goa.design/goa/codegen"
	"goa.design/goa/design"
	httpdesign "goa.design/goa/http/design"
)

// NewV2 returns the OpenAPI v2 specification for the given API.
func NewV2(root *httpdesign.RootExpr) (*V2, error) {
	if root == nil {
		return nil, nil
	}
	tags := tagsFromExpr(root.Metadata)
	u, err := url.Parse(root.Design.API.Servers[0].URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server URL: %s", err)
	}
	host := u.Host

	basePath := root.Path
	if hasAbsoluteRoutes(root) {
		basePath = ""
	}
	params, err := paramsFromExpr(root.MappedParams(), basePath)
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
	s := &V2{
		Swagger: "2.0",
		Info: &Info{
			Title:          root.Design.API.Title,
			Description:    root.Design.API.Description,
			TermsOfService: root.Design.API.TermsOfService,
			Contact:        root.Design.API.Contact,
			License:        root.Design.API.License,
			Version:        root.Design.API.Version,
			Extensions:     ExtensionsFromExpr(root.Metadata),
		},
		Host:         host,
		BasePath:     basePath,
		Paths:        make(map[string]interface{}),
		Consumes:     root.Consumes,
		Produces:     root.Produces,
		Parameters:   paramMap,
		Tags:         tags,
		ExternalDocs: docsFromExpr(root.Design.API.Docs),
	}

	for _, he := range root.HTTPErrors {
		res, err := responseSpecFromExpr(s, root, he.Response)
		if err != nil {
			return nil, err
		}
		if s.Responses == nil {
			s.Responses = make(map[string]*Response)
		}
		s.Responses[he.Name] = res
	}

	for _, res := range root.HTTPServices {
		for k, v := range ExtensionsFromExpr(res.Metadata) {
			s.Paths[k] = v
		}
		for _, fs := range res.FileServers {
			if mustGenerate(fs.Metadata) {
				if err := buildPathFromFileServer(s, root, fs); err != nil {
					return nil, err
				}
			}
		}
		for _, a := range res.HTTPEndpoints {
			if mustGenerate(a.Metadata) {
				for _, route := range a.Routes {
					if err := buildPathFromExpr(s, root, route, basePath); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	if err != nil {
		return nil, err
	}
	if len(Definitions) > 0 {
		s.Definitions = make(map[string]*Schema)
		for n, d := range Definitions {
			// sad but swagger doesn't support these
			d.Media = nil
			d.Links = nil
			s.Definitions[n] = d
		}
	}
	return s, nil
}

// ExtensionsFromExpr generates swagger extensions from the given metadata
// expression.
func ExtensionsFromExpr(mdata design.MetadataExpr) map[string]interface{} {
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

// hasAbsoluteRoutes returns true if any endpoint exposed by the API uses an
// absolute route of if the API has file servers. This is needed as OpenAPI does
// not support exceptions to the base path so if the API has any absolute route
// the base path must be "/" and all routes must be absolutes.
func hasAbsoluteRoutes(root *httpdesign.RootExpr) bool {
	hasAbsoluteRoutes := false
	for _, res := range root.HTTPServices {
		for _, fs := range res.FileServers {
			if !mustGenerate(fs.Metadata) {
				continue
			}
			hasAbsoluteRoutes = true
			break
		}
		for _, a := range res.HTTPEndpoints {
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

		tag.Extensions = ExtensionsFromExpr(mdata)

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

func paramsFromExpr(params *design.MappedAttributeExpr, path string) ([]*Parameter, error) {
	if params == nil {
		return nil, nil
	}
	var (
		res       []*Parameter
		wildcards = httpdesign.ExtractWildcards(path)
		i         = 0
	)
	codegen.WalkMappedAttr(params, func(n, pn string, required bool, at *design.AttributeExpr) error {
		in := "query"
		for _, w := range wildcards {
			if n == w {
				in = "path"
				required = true
				break
			}
		}
		param := paramFor(at, pn, in, required)
		res = append(res, param)
		i++
		return nil
	})
	return res, nil
}

func paramsFromHeaders(endpoint *httpdesign.EndpointExpr) []*Parameter {
	params := []*Parameter{}
	var (
		rma = endpoint.Service.MappedParams()
		ma  = endpoint.MappedHeaders()

		merged *design.MappedAttributeExpr
	)
	{
		if rma == nil {
			merged = ma
		} else if ma == nil {
			merged = rma
		} else {
			merged = design.DupMappedAtt(rma)
			merged.Merge(ma)
		}
	}

	for _, n := range *design.AsObject(merged.Type) {
		header := n.Attribute
		required := merged.IsRequiredNoDefault(n.Name)
		p := paramFor(header, n.Name, "header", required)
		params = append(params, p)
		return nil
	}
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
	if design.IsArray(at.Type) {
		p.Items = itemsFromExpr(design.AsArray(at.Type).ElemType)
		p.CollectionFormat = "multi"
	}
	switch at.Type {
	case design.Int, design.UInt, design.UInt32, design.UInt64:
		p.Type = "integer"
	case design.Int32, design.Int64:
		p.Type = "integer"
		p.Format = at.Type.Name()
	case design.Float32:
		p.Type = "number"
		p.Format = "float"
	case design.Float64:
		p.Type = "number"
		p.Format = "double"
	}
	p.Extensions = ExtensionsFromExpr(at.Metadata)
	initValidations(at, p)
	return p
}

func itemsFromExpr(at *design.AttributeExpr) *Items {
	items := &Items{Type: at.Type.Name()}
	initValidations(at, items)
	if design.IsArray(at.Type) {
		items.Items = itemsFromExpr(design.AsArray(at.Type).ElemType)
	}
	return items
}

func responseSpecFromExpr(s *V2, root *httpdesign.RootExpr, r *httpdesign.HTTPResponseExpr) (*Response, error) {
	var schema *Schema
	if mt, ok := r.Body.Type.(*design.ResultTypeExpr); ok {
		view := design.DefaultView
		if v, ok := r.Body.Metadata["view"]; ok {
			view = v[0]
		}
		schema = NewSchema()
		schema.Ref = ResultTypeRef(root.Design.API, mt, view)
	} else if r.Body.Type != design.Empty {
		schema = TypeSchema(root.Design.API, r.Body.Type)
	}
	headers, err := headersFromExpr(r.MappedHeaders())
	if err != nil {
		return nil, err
	}
	desc := r.Description
	if desc == "" {
		desc = fmt.Sprintf("%s response.", http.StatusText(r.StatusCode))
	}
	return &Response{
		Description: desc,
		Schema:      schema,
		Headers:     headers,
		Extensions:  ExtensionsFromExpr(r.Metadata),
	}, nil
}

func headersFromExpr(headers *design.MappedAttributeExpr) (map[string]*Header, error) {
	if headers == nil {
		return nil, nil
	}
	obj := design.AsObject(headers.Type)
	if obj == nil {
		return nil, fmt.Errorf("invalid headers definition, not an object")
	}
	res := make(map[string]*Header)
	codegen.WalkMappedAttr(headers, func(_, n string, required bool, at *design.AttributeExpr) error {
		header := &Header{
			Default:     at.DefaultValue,
			Description: at.Description,
			Type:        at.Type.Name(),
		}
		initValidations(at, header)
		res[n] = header
		return nil
	})
	if len(res) == 0 {
		return nil, nil
	}
	return res, nil
}

func buildPathFromFileServer(s *V2, root *httpdesign.RootExpr, fs *httpdesign.FileServerExpr) error {
	for _, path := range fs.RequestPaths {
		wcs := httpdesign.ExtractWildcards(path)
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
				Schema:      &Schema{Type: File},
			},
		}
		if len(wcs) > 0 {
			schema := TypeSchema(root.Design.API, design.ErrorResult)
			responses["404"] = &Response{Description: "File not found", Schema: schema}
		}

		operationID := fmt.Sprintf("%s#%s", fs.Service.Name(), path)
		schemes := root.Design.API.Schemes()

		operation := &Operation{
			Description:  fs.Description,
			Summary:      summaryFromExpr(fmt.Sprintf("Download %s", fs.FilePath), fs.Metadata),
			ExternalDocs: docsFromExpr(fs.Docs),
			OperationID:  operationID,
			Parameters:   param,
			Responses:    responses,
			Schemes:      schemes,
		}

		key := httpdesign.WildcardRegex.ReplaceAllStringFunc(
			path,
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
		p.Extensions = ExtensionsFromExpr(fs.Metadata)
	}

	return nil
}

func buildPathFromExpr(s *V2, root *httpdesign.RootExpr, route *httpdesign.RouteExpr, basePath string) error {
	endpoint := route.Endpoint

	tagNames := tagNamesFromExpr(endpoint.Service.Metadata, endpoint.Metadata)
	if len(tagNames) == 0 {
		// By default tag with service name
		tagNames = []string{route.Endpoint.Service.Name()}
	}
	for _, key := range route.FullPaths() {
		params, err := paramsFromExpr(endpoint.AllParams(), key)
		if err != nil {
			return err
		}

		params = append(params, paramsFromHeaders(endpoint)...)

		responses := make(map[string]*Response, len(endpoint.Responses))
		for _, r := range endpoint.Responses {
			resp, err := responseSpecFromExpr(s, root, r)
			if err != nil {
				return err
			}
			responses[strconv.Itoa(r.StatusCode)] = resp
		}
		for _, er := range endpoint.HTTPErrors {
			resp, err := responseSpecFromExpr(s, root, er.Response)
			if err != nil {
				return err
			}
			responses[strconv.Itoa(er.Response.StatusCode)] = resp
		}

		if endpoint.Body.Type != design.Empty {
			pp := &Parameter{
				Name:        endpoint.Body.Type.Name(),
				In:          "body",
				Description: endpoint.Body.Description,
				Required:    true,
				Schema:      TypeSchema(root.Design.API, endpoint.Body.Type),
			}
			params = append(params, pp)
		}

		operationID := fmt.Sprintf("%s#%s", endpoint.Service.Name(), endpoint.Name())
		index := 0
		for i, rt := range endpoint.Routes {
			if rt == route {
				index = i
				break
			}
		}
		if index > 0 {
			operationID = fmt.Sprintf("%s#%d", operationID, index)
		}

		schemes := endpoint.Service.Schemes()
		if len(schemes) == 0 {
			schemes = root.Design.API.Schemes()
		}

		operation := &Operation{
			Tags:         tagNames,
			Description:  endpoint.Description(),
			Summary:      summaryFromExpr(endpoint.Name()+" "+endpoint.Service.Name(), endpoint.Metadata),
			ExternalDocs: docsFromExpr(endpoint.MethodExpr.Docs),
			OperationID:  operationID,
			Parameters:   params,
			Responses:    responses,
			Schemes:      schemes,
			Deprecated:   false,
			Extensions:   ExtensionsFromExpr(route.Metadata),
		}

		if key == "" {
			key = "/"
		}
		bp := httpdesign.WildcardRegex.ReplaceAllStringFunc(
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
		switch route.Method {
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
		p.Extensions = ExtensionsFromExpr(route.Endpoint.Metadata)
	}
	return nil
}

func scopesList(scopes []string) string {
	sort.Strings(scopes)

	var lines []string
	for _, scope := range scopes {
		lines = append(lines, fmt.Sprintf("  * `%s`", scope))
	}
	return strings.Join(lines, "\n")
}

func docsFromExpr(docs *design.DocsExpr) *ExternalDocs {
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
	initFormatValidation(def, string(val.Format))
	initPatternValidation(def, val.Pattern)
	if val.Minimum != nil {
		initMinimumValidation(def, val.Minimum)
	}
	if val.Maximum != nil {
		initMaximumValidation(def, val.Maximum)
	}
	if val.MinLength != nil {
		initMinLengthValidation(def, design.IsArray(attr.Type), val.MinLength)
	}
	if val.MaxLength != nil {
		initMaxLengthValidation(def, design.IsArray(attr.Type), val.MaxLength)
	}
}
