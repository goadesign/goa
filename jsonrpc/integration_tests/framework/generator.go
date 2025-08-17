package framework

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"goa.design/goa/v3/codegen"
	goatemplate "goa.design/goa/v3/codegen/template"

	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"
)

//go:embed templates/*.tpl templates/dsl/*.tpl templates/impl/*.tpl templates/partial/*.tpl
var templateFS embed.FS

// generatorTemplates is the template reader for the test generator
var generatorTemplates = &goatemplate.TemplateReader{FS: templateFS}

// Generator generates test service code using templates
// Flow:
//  1. Build design data from scenarios
//  2. Render design (go.mod, design.go)
//  3. Run goa gen/example
//  4. Build implementation data (align method Go names with generated interface)
//  5. Render implementations
//
// Keeping the phases explicit here makes the generator logic predictable and
// easier to maintain/debug.
type Generator struct {
	workDir string
	methods map[string]MethodInfo
}

// NewGenerator creates a new generator
func NewGenerator(workDir string, methods map[string]MethodInfo) *Generator {
	return &Generator{workDir: workDir, methods: methods}
}

// Generate runs the full generation pipeline.
func (g *Generator) Generate() error {
	// 1. Build design data
	design := g.buildDesignData()
	// 2. Render design
	if err := g.renderDesign(design); err != nil {
		return err
	}
	// 3. Run goa gen/example
	if err := g.runPostGeneration(); err != nil {
		return fmt.Errorf("post generation: %w", err)
	}
	// 4. Build implementation data
	impl := g.buildImplementationData(design)
	// 5. Render implementations
	if err := g.renderImplementation(impl); err != nil {
		return err
	}
	return nil
}

// buildDesignData creates the semantic data for design generation.
func (g *Generator) buildDesignData() *DesignData {
	data := &DesignData{APIName: "TestAPI", APITitle: "JSON-RPC Integration Test API", APIDescription: "Auto-generated API for integration testing", Services: make([]*ServiceData, 0)}
	serviceMap := make(map[string]*ServiceData)
	for _, info := range g.methods {
		serviceName := g.getServiceName(info)
		if _, exists := serviceMap[serviceName]; !exists {
			serviceMap[serviceName] = &ServiceData{Name: serviceName, Title: goify(serviceName), Description: fmt.Sprintf("Test service for %s", serviceName), JSONRPCPath: g.getJSONRPCPath(serviceName), Methods: make([]*MethodData, 0)}
		}
		md := g.buildMethodData(info)
		serviceMap[serviceName].Methods = append(serviceMap[serviceName].Methods, md)
		if md.ReturnsError {
			serviceMap[serviceName].HasErrors = true
		}
	}
	for _, service := range serviceMap {
		data.Services = append(data.Services, service)
	}
	return data
}

// renderDesign writes the design files required prior to running goa gen/example.
func (g *Generator) renderDesign(design *DesignData) error {
	// go.mod
	gomod := &codegen.File{
		Path: "go.mod",
		SectionTemplates: []*codegen.SectionTemplate{{
			Name:   "go-mod",
			Source: generatorTemplates.Read("go_mod"),
			Data:   map[string]string{"GoaPath": g.repoRootReplace()},
		}},
	}
	if _, err := gomod.Render(g.workDir); err != nil {
		return fmt.Errorf("render go.mod: %w", err)
	}
	// design.go
	designFile := &codegen.File{
		Path: filepath.Join("design", "design.go"),
		SectionTemplates: []*codegen.SectionTemplate{{
			Name:    "design",
			Source:  generatorTemplates.Read("dsl/design", "method", "type"),
			FuncMap: g.templateFuncs(),
			Data:    design,
		}},
	}
	if _, err := designFile.Render(g.workDir); err != nil {
		return fmt.Errorf("render design.go: %w", err)
	}
	return nil
}

// renderImplementation writes the service implementation files.
func (g *Generator) renderImplementation(impl *ImplementationData) error {
	for _, f := range g.filesImpl(impl) {
		if _, err := f.Render(g.workDir); err != nil {
			return fmt.Errorf("render %s: %w", f.Path, err)
		}
	}
	return nil
}

// buildMethodData creates semantic data for a method.
func (g *Generator) buildMethodData(info MethodInfo) *MethodData {
	data := &MethodData{
		Name:             info.Name(),
		GoName:           goify(info.Name()),
		Description:      g.getMethodDescription(info),
		Info:             info,
		IsNotification:   info.Modifier == ModifierNotify,
		ReturnsError:     info.Modifier == ModifierError,
		HasValidation:    info.Modifier == ModifierValidate,
		HasFinalResponse: info.Modifier == ModifierFinal,
		Transport:        info.Transport,
		IsStreaming:      info.IsStreaming(),
	}
	// Non-streaming payload
	if info.Modifier != ModifierNotify && info.Action != ActionGenerate && (!info.HasStreamingPayload() || info.IsSSE()) {
		data.Payload = g.buildTypeSpec(info.Type, info.Modifier)
	}
	// Streaming
	if info.IsStreaming() {
		isBidi := info.IsWebSocket() && info.HasStreamingPayload() && info.HasStreamingResult()
		if info.HasStreamingPayload() {
			data.StreamingPayload = g.buildStreamingTypeSpec(info.Type, true, isBidi, info)
			data.StreamKind = "payload"
		}
		if info.HasStreamingResult() {
			data.Result = g.buildStreamingTypeSpec(info.Type, false, isBidi, info)
			if data.StreamKind == "payload" {
				data.StreamKind = "bidirectional"
			} else {
				data.StreamKind = "result"
			}
			if info.IsSSE() && info.Modifier == ModifierFinal && data.Result != nil {
				data.Result.Fields = append(data.Result.Fields, FieldSpec{Position: len(data.Result.Fields) + 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Description: "Response ID (for final response)", Required: false})
			}
		}
	} else if info.Modifier != ModifierNotify && info.Modifier != ModifierError {
		data.Result = g.buildTypeSpec(info.Type, "")
	}
	return data
}

// buildTypeSpec creates a TypeSpec based on the type string and modifier.
func (g *Generator) buildTypeSpec(typeStr, modifier string) *TypeSpec {
	// Special handling for id mapping: force object with value + id + request_id (strings)
	if modifier == ModifierIDMap {
		wrap := func(val *TypeSpec) *TypeSpec {
			return &TypeSpec{
				Kind: "object",
				Fields: []FieldSpec{
					{Position: 1, Name: "value", GoName: "Value", Type: val, Required: true},
					{Position: 2, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}},
					{Position: 3, Name: "request_id", GoName: "RequestID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}},
				},
			}
		}
		switch typeStr {
		case TypeString:
			return wrap(&TypeSpec{Kind: "primitive", Primitive: "String"})
		case TypeInt:
			return wrap(&TypeSpec{Kind: "primitive", Primitive: "Int"})
		case TypeBool:
			return wrap(&TypeSpec{Kind: "primitive", Primitive: "Boolean"})
		case TypeArray:
			return wrap(&TypeSpec{Kind: "array", ArrayElem: &TypeSpec{Kind: "primitive", Primitive: "String"}})
		case TypeObject:
			return wrap(&TypeSpec{Kind: "object", Fields: []FieldSpec{
				{Position: 1, Name: "field1", GoName: "Field1", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true},
				{Position: 2, Name: "field2", GoName: "Field2", Type: &TypeSpec{Kind: "primitive", Primitive: "Int"}, Required: true},
				{Position: 3, Name: "field3", GoName: "Field3", Type: &TypeSpec{Kind: "primitive", Primitive: "Boolean"}, Required: true},
			}})
		case TypeMap:
			return wrap(&TypeSpec{Kind: "map", MapKey: &TypeSpec{Kind: "primitive", Primitive: "String"}, MapValue: &TypeSpec{Kind: "primitive", Primitive: "Any"}})
		default:
			return wrap(&TypeSpec{Kind: "any"})
		}
	}
	switch typeStr {
	case TypeString:
		// For validated primitives, wrap in object for JSON-RPC
		if modifier == ModifierValidate {
			return &TypeSpec{
				Kind: "object",
				Fields: []FieldSpec{
					{Position: 1, Name: "value", GoName: "Value", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true},
				},
			}
		}
		return &TypeSpec{Kind: "primitive", Primitive: "String"}
	case TypeInt:
		return &TypeSpec{Kind: "primitive", Primitive: "Int"}
	case TypeBool:
		return &TypeSpec{Kind: "primitive", Primitive: "Boolean"}
	case TypeArray:
		return &TypeSpec{
			Kind: "object",
			Fields: []FieldSpec{
				{Position: 1, Name: "items", GoName: "Items", Type: &TypeSpec{Kind: "array", ArrayElem: &TypeSpec{Kind: "primitive", Primitive: "String"}}, Required: true},
			},
		}
	case TypeObject:
		return &TypeSpec{
			Kind: "object",
			Fields: []FieldSpec{
				{Position: 1, Name: "field1", GoName: "Field1", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true},
				{Position: 2, Name: "field2", GoName: "Field2", Type: &TypeSpec{Kind: "primitive", Primitive: "Int"}, Required: true},
				{Position: 3, Name: "field3", GoName: "Field3", Type: &TypeSpec{Kind: "primitive", Primitive: "Boolean"}, Required: true},
			},
		}
	case TypeMap:
		return &TypeSpec{
			Kind:     "map",
			MapKey:   &TypeSpec{Kind: "primitive", Primitive: "String"},
			MapValue: &TypeSpec{Kind: "primitive", Primitive: "Any"},
		}
	default:
		return &TypeSpec{Kind: "any"}
	}
}

// buildStreamingTypeSpec creates a TypeSpec for streaming types
func (g *Generator) buildStreamingTypeSpec(typeStr string, _ bool, isBidirectional bool, info MethodInfo) *TypeSpec {
	// For WebSocket bidirectional methods, include mapping fields only when explicitly requested via idmap
	if isBidirectional {
		switch typeStr {
		case TypeString:
			if info.Modifier == ModifierIDMap {
				return &TypeSpec{
					Kind:    "object",
					NeedsID: true,
					Fields: []FieldSpec{
						{Position: 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: false, Description: "Business-level ID"},
						{Position: 2, Name: "request_id", GoName: "RequestID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: false, Description: "Mapped JSON-RPC envelope ID"},
						{Position: 3, Name: "value", GoName: "Value", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true, Description: "String value"},
					},
				}
			}
			// No id mapping: only value field
			return &TypeSpec{Kind: "object", Fields: []FieldSpec{{Position: 1, Name: "value", GoName: "Value", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true}}}
		case TypeArray:
			if info.Modifier == ModifierIDMap {
				return &TypeSpec{Kind: "object", NeedsID: true, Fields: []FieldSpec{
					{Position: 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}},
					{Position: 2, Name: "request_id", GoName: "RequestID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}},
					{Position: 3, Name: "items", GoName: "Items", Type: &TypeSpec{Kind: "array", ArrayElem: &TypeSpec{Kind: "primitive", Primitive: "String"}}, Required: true},
				}}
			}
			return &TypeSpec{Kind: "object", Fields: []FieldSpec{{Position: 1, Name: "items", GoName: "Items", Type: &TypeSpec{Kind: "array", ArrayElem: &TypeSpec{Kind: "primitive", Primitive: "String"}}, Required: true}}}
		case TypeObject:
			if info.Modifier == ModifierIDMap {
				return &TypeSpec{Kind: "object", NeedsID: true, Fields: []FieldSpec{
					{Position: 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}},
					{Position: 2, Name: "request_id", GoName: "RequestID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}},
					{Position: 3, Name: "field1", GoName: "Field1", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true},
					{Position: 4, Name: "field2", GoName: "Field2", Type: &TypeSpec{Kind: "primitive", Primitive: "Int"}, Required: true},
					{Position: 5, Name: "field3", GoName: "Field3", Type: &TypeSpec{Kind: "primitive", Primitive: "Boolean"}, Required: true},
				}}
			}
			return &TypeSpec{Kind: "object", Fields: []FieldSpec{
				{Position: 1, Name: "field1", GoName: "Field1", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true},
				{Position: 2, Name: "field2", GoName: "Field2", Type: &TypeSpec{Kind: "primitive", Primitive: "Int"}, Required: true},
				{Position: 3, Name: "field3", GoName: "Field3", Type: &TypeSpec{Kind: "primitive", Primitive: "Boolean"}, Required: true},
			}}
		default:
			if info.Modifier == ModifierIDMap {
				return &TypeSpec{Kind: "object", NeedsID: true, Fields: []FieldSpec{
					{Position: 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}},
					{Position: 2, Name: "request_id", GoName: "RequestID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}},
					{Position: 3, Name: "data", GoName: "Data", Type: &TypeSpec{Kind: "primitive", Primitive: "Any"}, Required: true},
				}}
			}
			return &TypeSpec{Kind: "object", Fields: []FieldSpec{{Position: 1, Name: "data", GoName: "Data", Type: &TypeSpec{Kind: "primitive", Primitive: "Any"}, Required: true}}}
		}
	}

	// For non-bidirectional streaming, wrap primitives in objects
	spec := g.buildTypeSpec(typeStr, "")
	if spec.Kind == "primitive" {
		return &TypeSpec{Kind: "object", Fields: []FieldSpec{{Position: 1, Name: "value", GoName: "Value", Type: spec, Required: true, Description: fmt.Sprintf("%s value", spec.Primitive)}}}
	}
	return spec
}

// buildImplementationData creates the semantic data for implementation.
// Crucially, this method aligns GoName with the generated service interface
// names so that duplicates are handled consistently.
func (g *Generator) buildImplementationData(design *DesignData) *ImplementationData {
	data := &ImplementationData{PackageName: "testservice", Services: make([]*ServiceImplData, 0)}
	for _, service := range design.Services {
		implService := &ServiceImplData{ServiceData: service, ServicePackage: service.Name, Methods: make([]*MethodImplData, 0)}
		if pairs := g.parseServiceMethodPairs(service.Name); len(pairs) > 0 {
			used := make([]bool, len(service.Methods))
			for _, p := range pairs {
				for j := 0; j < len(service.Methods); j++ {
					if used[j] {
						continue
					}
					if service.Methods[j].Name == p.dsl {
						service.Methods[j].GoName = p.goName
						used[j] = true
						break
					}
				}
			}
		}
		for _, method := range service.Methods {
			implMethod := g.buildMethodImplData(method, service.Name)
			implService.Methods = append(implService.Methods, implMethod)
		}
		data.Services = append(data.Services, implService)
	}
	return data
}

// buildMethodImplData creates implementation data for a method
func (g *Generator) buildMethodImplData(method *MethodData, serviceName string) *MethodImplData {
	data := &MethodImplData{MethodData: method, ServicePackage: serviceName, HasPayload: method.Payload != nil || method.StreamingPayload != nil, HasResult: method.Result != nil}
	if method.Payload != nil {
		if method.Payload.Kind == "primitive" {
			data.PayloadRef = strings.ToLower(method.Payload.Primitive)
		} else {
			data.PayloadRef = fmt.Sprintf("*%s.%sPayload", serviceName, method.GoName)
		}
	} else if method.StreamingPayload != nil && data.StreamKind == "bidirectional" {
		data.PayloadRef = fmt.Sprintf("*%s.%sPayload", serviceName, method.GoName)
	}
	if method.Result != nil {
		if method.Result.Kind == "primitive" {
			data.ResultRef = strings.ToLower(method.Result.Primitive)
		} else {
			data.ResultRef = fmt.Sprintf("*%s.%sResult", serviceName, method.GoName)
		}
	}
	if method.IsStreaming {
		data.StreamInterface = fmt.Sprintf("%sServerStream", method.GoName)
	}
	return data
}

// templateFuncs returns the template functions used by templates.
func (g *Generator) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"goify": goify,
		"hasStreamingMethod": func(methods []*MethodImplData) bool {
			for _, m := range methods {
				if m.IsStreaming {
					return true
				}
			}
			return false
		},
		"collectRequired": func(fields []FieldSpec) []string {
			var required []string
			for _, f := range fields {
				if f.Required {
					required = append(required, f.Name)
				}
			}
			return required
		},
	}
}

// getServiceName returns which service bucket a method belongs to.
func (g *Generator) getServiceName(info MethodInfo) string {
	if info.IsSSE() {
		return "testsse"
	}
	if info.IsWebSocket() {
		return "testws"
	}
	return "test"
}

// getJSONRPCPath returns the JSON-RPC base path for a service.
func (g *Generator) getJSONRPCPath(serviceName string) string {
	switch serviceName {
	case "testsse":
		return "/jsonrpc/sse"
	case "testws":
		return "/jsonrpc/ws"
	default:
		return "/jsonrpc"
	}
}

// getMethodDescription returns a human-friendly method description.
func (g *Generator) getMethodDescription(info MethodInfo) string {
	desc := fmt.Sprintf("%s %s", info.Action, info.Type)
	if info.Modifier != "" {
		desc += fmt.Sprintf(" (%s)", info.Modifier)
	}
	return desc
}

// runPostGeneration runs goa gen and example assuming goa is in PATH.
func (g *Generator) runPostGeneration() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = g.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w\nOutput: %s", err, output)
	}
	cmd = exec.Command("goa", "gen", "testservice/design", "-o", g.workDir)
	cmd.Dir = g.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("goa gen failed: %w\nOutput: %s", err, output)
	}
	cmd = exec.Command("goa", "example", "testservice/design", "-o", g.workDir)
	cmd.Dir = g.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("goa example failed: %w\nOutput: %s", err, output)
	}
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = g.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("final go mod tidy failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// repoRootReplace returns a path suitable for the replace directive in go.mod.
// Prefer GOA_REPO env var for tests; otherwise fall back to relative path.
func (g *Generator) repoRootReplace() string {
	if p := os.Getenv("GOA_REPO"); p != "" {
		return p
	}
	absPath, err := filepath.Abs("../../..")
	if err != nil {
		return "../../../.."
	}
	return absPath
}

// goify converts a string to a Go identifier.
func goify(s string) string { return codegen.Goify(s, true) }

// methodPair captures a DSL name and its corresponding Go name from the generated code.
type methodPair struct{ dsl, goName string }

// parseServiceMethodPairs parses gen/<service>/service.go and returns (dslName, goName) pairs.
func (g *Generator) parseServiceMethodPairs(serviceName string) []methodPair {
	path := filepath.Join(g.workDir, "gen", serviceName, "service.go")
	fs := gotoken.NewFileSet()
	af, err := goparser.ParseFile(fs, path, nil, goparser.ParseComments)
	if err != nil {
		return nil
	}
	var goNames, dslNames []string
	goast.Inspect(af, func(n goast.Node) bool {
		ts, ok := n.(*goast.TypeSpec)
		if !ok || ts.Name == nil || ts.Name.Name != "Service" {
			return true
		}
		it, ok := ts.Type.(*goast.InterfaceType)
		if !ok || it.Methods == nil {
			return false
		}
		for _, fld := range it.Methods.List {
			if len(fld.Names) == 0 {
				continue
			}
			name := fld.Names[0].Name
			if name == "HandleStream" {
				continue
			}
			goNames = append(goNames, name)
		}
		return false
	})
	goast.Inspect(af, func(n goast.Node) bool {
		gen, ok := n.(*goast.GenDecl)
		if !ok || gen.Tok != gotoken.VAR {
			return true
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*goast.ValueSpec)
			if !ok || len(vs.Names) == 0 || vs.Names[0].Name != "MethodNames" {
				continue
			}
			if len(vs.Values) == 0 {
				continue
			}
			cl, ok := vs.Values[0].(*goast.CompositeLit)
			if !ok {
				continue
			}
			for _, elt := range cl.Elts {
				bl, ok := elt.(*goast.BasicLit)
				if !ok || bl.Kind != gotoken.STRING {
					continue
				}
				dslNames = append(dslNames, strings.Trim(bl.Value, "\""))
			}
			return false
		}
		return true
	})
	pairs := make([]methodPair, 0, len(goNames))
	for i := 0; i < len(goNames) && i < len(dslNames); i++ {
		pairs = append(pairs, methodPair{dsl: dslNames[i], goName: goNames[i]})
	}
	return pairs
}

// filesImpl returns files needed after goa gen (service implementations)
func (g *Generator) filesImpl(impl *ImplementationData) []*codegen.File {
	files := make([]*codegen.File, 0, len(impl.Services))
	for _, service := range impl.Services {
		imports := []*codegen.ImportSpec{
			{Path: "context"},
			{Path: "log"},
			{Path: "fmt"},
			{Path: "time"},
			{Path: "strings"},
			{Path: "sort"},
			{Path: "io"},
			{Name: "goa", Path: "goa.design/goa/v3/pkg"},
			{Name: service.ServicePackage, Path: fmt.Sprintf("testservice/gen/%s", service.ServicePackage)},
		}
		sections := []*codegen.SectionTemplate{
			codegen.Header(fmt.Sprintf("%s service implementation", service.Title), "testservice", imports),
			{
				Name:    "service-impl",
				Source:  generatorTemplates.Read("impl/service", "method_signature", "error", "echo", "transform", "generate", "streaming_sse", "streaming_websocket", "notify", "validate"),
				FuncMap: g.templateFuncs(),
				Data:    service,
			},
		}
		files = append(files, &codegen.File{Path: fmt.Sprintf("%s.go", service.Name), SectionTemplates: sections})
	}
	return files
}
