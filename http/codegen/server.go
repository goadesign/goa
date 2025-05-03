package codegen

import (
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"goa.design/goa/v3/codegen"
	"goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/expr"
)

// ServerFiles returns the generated HTTP server files.
func ServerFiles(genpkg string, root *expr.RootExpr) []*codegen.File {
	var files []*codegen.File
	for _, svc := range root.API.HTTP.Services {
		files = append(files, serverFile(genpkg, svc))
		if f := websocketServerFile(genpkg, svc); f != nil {
			files = append(files, f)
		}
		if f := sseServerFile(genpkg, svc); f != nil {
			files = append(files, f)
		}
	}
	for _, svc := range root.API.HTTP.Services {
		if f := serverEncodeDecodeFile(genpkg, svc); f != nil {
			files = append(files, f)
		}
	}
	return files
}

// server returns the file implementing the HTTP server.
func serverFile(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	data := HTTPServices.Get(svc.Name())
	svcName := data.Service.PathName
	fpath := filepath.Join(codegen.Gendir, "http", svcName, "server", "server.go")
	title := fmt.Sprintf("%s HTTP server", svc.Name())
	funcs := map[string]any{
		"join":                strings.Join,
		"hasWebSocket":        hasWebSocket,
		"isWebSocketEndpoint": isWebSocketEndpoint,
		"isSSEEndpoint":       isSSEEndpoint,
		"viewedServerBody":    viewedServerBody,
		"mustDecodeRequest":   mustDecodeRequest,
		"addLeadingSlash":     addLeadingSlash,
		"dir":                 path.Dir,
	}
	imports := []*codegen.ImportSpec{
		{Path: "bufio"},
		{Path: "context"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "mime/multipart"},
		{Path: "net/http"},
		{Path: "path"},
		{Path: "strings"},
		{Path: "github.com/gorilla/websocket"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
		{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
	}
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{
		codegen.Header(title, "server", imports),
	}

	sections = append(sections,
		&codegen.SectionTemplate{Name: "server-struct", Source: readTemplate("server_struct"), Data: data},
		&codegen.SectionTemplate{Name: "server-mountpoint", Source: readTemplate("mount_point_struct"), Data: data})

	for _, e := range data.Endpoints {
		if e.MultipartRequestDecoder != nil {
			sections = append(sections, &codegen.SectionTemplate{
				Name:   "multipart-request-decoder-type",
				Source: readTemplate("multipart_request_decoder_type"),
				Data:   e.MultipartRequestDecoder,
			})
		}
	}

	sections = append(sections,
		&codegen.SectionTemplate{Name: "server-init", Source: readTemplate("server_init"), Data: data, FuncMap: funcs},
		&codegen.SectionTemplate{Name: "server-service", Source: readTemplate("server_service"), Data: data},
		&codegen.SectionTemplate{Name: "server-use", Source: readTemplate("server_use"), Data: data},
		&codegen.SectionTemplate{Name: "server-method-names", Source: readTemplate("server_method_names"), Data: data},
		&codegen.SectionTemplate{Name: "server-mount", Source: readTemplate("server_mount"), Data: data, FuncMap: funcs})

	for _, e := range data.Endpoints {
		sections = append(sections,
			&codegen.SectionTemplate{Name: "server-handler", Source: readTemplate("server_handler"), Data: e},
			&codegen.SectionTemplate{Name: "server-handler-init", Source: readTemplate("server_handler_init"), FuncMap: funcs, Data: e})
	}
	if len(data.FileServers) > 0 {
		mappedFiles := make(map[string]string)
		for _, fs := range data.FileServers {
			if !fs.IsDir {
				for _, p := range fs.RequestPaths {
					baseFilePath := "/" + filepath.Base(fs.FilePath)
					baseRequestPath := "/" + filepath.Base(p)
					if baseFilePath == baseRequestPath {
						continue
					}
					mappedFiles[baseRequestPath] = baseFilePath
				}
			}
		}
		sections = append(sections, &codegen.SectionTemplate{Name: "append-fs", Source: readTemplate("append_fs"), FuncMap: funcs, Data: mappedFiles})
	}
	for _, s := range data.FileServers {
		sections = append(sections, &codegen.SectionTemplate{Name: "server-files", Source: readTemplate("file_server"), FuncMap: funcs, Data: s})
	}

	return &codegen.File{Path: fpath, SectionTemplates: sections}
}

// serverEncodeDecodeFile returns the file defining the HTTP server encoding and
// decoding logic.
func serverEncodeDecodeFile(genpkg string, svc *expr.HTTPServiceExpr) *codegen.File {
	data := HTTPServices.Get(svc.Name())
	svcName := data.Service.PathName
	path := filepath.Join(codegen.Gendir, "http", svcName, "server", "encode_decode.go")
	title := fmt.Sprintf("%s HTTP server encoders and decoders", svc.Name())
	imports := []*codegen.ImportSpec{
		{Path: "context"},
		{Path: "errors"},
		{Path: "fmt"},
		{Path: "io"},
		{Path: "net/http"},
		{Path: "strconv"},
		{Path: "strings"},
		{Path: "encoding/json"},
		{Path: "mime/multipart"},
		{Path: "unicode/utf8"},
		codegen.GoaImport(""),
		codegen.GoaNamedImport("http", "goahttp"),
		{Path: genpkg + "/" + svcName, Name: data.Service.PkgName},
		{Path: genpkg + "/" + svcName + "/" + "views", Name: data.Service.ViewsPkg},
	}
	imports = append(imports, data.Service.UserTypeImports...)
	sections := []*codegen.SectionTemplate{codegen.Header(title, "server", imports)}

	for _, e := range data.Endpoints {
		if e.Redirect == nil && !isWebSocketEndpoint(e) {
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "response-encoder",
				FuncMap: transTmplFuncs(svc),
				Source:  readTemplate("response_encoder", "response", "header_conversion"),
				Data:    e,
			})
		}
		if mustDecodeRequest(e) {
			fm := transTmplFuncs(svc)
			fm["mapQueryDecodeData"] = mapQueryDecodeData
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "request-decoder",
				Source:  readTemplate("request_decoder", "request_elements", "slice_item_conversion", "element_slice_conversion", "query_slice_conversion", "query_type_conversion", "query_map_conversion", "path_conversion"),
				FuncMap: fm,
				Data:    e,
			})
		}
		if e.MultipartRequestDecoder != nil {
			fm := transTmplFuncs(svc)
			fm["mapQueryDecodeData"] = mapQueryDecodeData
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "multipart-request-decoder",
				Source:  readTemplate("multipart_request_decoder", "request_elements", "slice_item_conversion", "element_slice_conversion", "query_slice_conversion", "query_type_conversion", "query_map_conversion", "path_conversion"),
				FuncMap: fm,
				Data:    e.MultipartRequestDecoder,
			})
		}
		if len(e.Errors) > 0 {
			sections = append(sections, &codegen.SectionTemplate{
				Name:    "error-encoder",
				Source:  readTemplate("error_encoder", "response", "header_conversion"),
				FuncMap: transTmplFuncs(svc),
				Data:    e,
			})
		}
	}
	for _, h := range data.ServerTransformHelpers {
		sections = append(sections, &codegen.SectionTemplate{
			Name:   "server-transform-helper",
			Source: readTemplate("transform_helper"),
			Data:   h,
		})
	}

	// If all endpoints use skip encoding and decoding of both payloads and
	// results and define no error then this file is irrelevant.
	if len(sections) == 1 {
		return nil
	}

	return &codegen.File{Path: path, SectionTemplates: sections}
}

func transTmplFuncs(s *expr.HTTPServiceExpr) map[string]any {
	return map[string]any{
		"goTypeRef": func(dt expr.DataType) string {
			return service.Services.Get(s.Name()).Scope.GoTypeRef(&expr.AttributeExpr{Type: dt})
		},
		"isAliased": func(dt expr.DataType) bool {
			_, ok := dt.(expr.UserType)
			return ok
		},
		"conversionData":       conversionData,
		"headerConversionData": headerConversionData,
		"printValue":           printValue,
		"viewedServerBody":     viewedServerBody,
	}
}

// mustDecodeRequest returns true if the Payload type is not empty.
func mustDecodeRequest(e *EndpointData) bool {
	return e.Payload.Ref != ""
}

// conversionData creates a template context suitable for executing the
// "type_conversion" template.
func conversionData(varName, name string, dt expr.DataType) map[string]any {
	return map[string]any{
		"VarName": varName,
		"Name":    name,
		"Type":    dt,
	}
}

// headerConversionData produces the template data suitable for executing the
// "header_conversion" template.
func headerConversionData(dt expr.DataType, varName string, required bool, target string) map[string]any {
	return map[string]any{
		"Type":     dt,
		"VarName":  varName,
		"Required": required,
		"Target":   target,
	}
}

// printValue generates the Go code for a literal string containing the given
// value. printValue panics if the data type is not a primitive or an array.
func printValue(dt expr.DataType, v any) string {
	switch actual := dt.(type) {
	case *expr.Array:
		val := reflect.ValueOf(v)
		elems := make([]string, val.Len())
		for i := 0; i < val.Len(); i++ {
			elems[i] = printValue(actual.ElemType.Type, val.Index(i).Interface())
		}
		return strings.Join(elems, ", ")
	case expr.Primitive:
		return fmt.Sprintf("%v", v)
	default:
		panic("unsupported type value " + dt.Name()) // bug
	}
}

// viewedServerBody returns the type data that uses the given view for
// rendering.
func viewedServerBody(sbd []*TypeData, view string) *TypeData {
	for _, v := range sbd {
		if v.View == view {
			return v
		}
	}
	panic("view not found in server body types: " + view)
}

func addLeadingSlash(s string) string {
	if s == "" || s[0] != '/' {
		return "/" + s
	}
	return s
}

func mapQueryDecodeData(dt expr.DataType, varName string, inc int) map[string]any {
	return map[string]any{
		"Type":      dt,
		"VarName":   varName,
		"Loop":      string(rune(97 + inc)),
		"Increment": inc + 1,
		"Depth":     codegen.MapDepth(expr.AsMap(dt)),
	}
}
