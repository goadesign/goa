package framework

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"goa.design/goa/v3/codegen"
	goatemplate "goa.design/goa/v3/codegen/template"
)

//go:embed templates/*.tpl templates/dsl/*.tpl templates/impl/*.tpl templates/partial/*.tpl
var templateFS embed.FS

// generatorTemplates is the template reader for the test generator
var generatorTemplates = &goatemplate.TemplateReader{FS: templateFS}

// Generator generates test service code using templates
type Generator struct {
	workDir string
	methods map[string]MethodInfo
}

// NewGenerator creates a new generator
func NewGenerator(workDir string, methods map[string]MethodInfo) *Generator {
	return &Generator{
		workDir: workDir,
		methods: methods,
	}
}

// Generate creates the complete test service
func (g *Generator) Generate() error {
	// Build semantic data
	designData := g.buildDesignData()
	implData := g.buildImplementationData(designData)

	// Generate files
	files := g.Files(designData, implData)

	// Render all files
	for _, f := range files {
		if _, err := f.Render(g.workDir); err != nil {
			return fmt.Errorf("render %s: %w", f.Path, err)
		}
	}

	// Run post-generation commands
	if err := g.runPostGeneration(); err != nil {
		return fmt.Errorf("post generation: %w", err)
	}

	return nil
}

// buildDesignData creates the semantic data for design generation
func (g *Generator) buildDesignData() *DesignData {
	data := &DesignData{
		APIName:        "TestAPI",
		APITitle:       "JSON-RPC Integration Test API",
		APIDescription: "Auto-generated API for integration testing",
		Services:       make([]*ServiceData, 0),
	}

	// Group methods by service
	serviceMap := make(map[string]*ServiceData)

	for _, info := range g.methods {
		serviceName := g.getServiceName(info)

		if _, exists := serviceMap[serviceName]; !exists {
			serviceMap[serviceName] = &ServiceData{
				Name:        serviceName,
				Title:       goify(serviceName),
				Description: fmt.Sprintf("Test service for %s", serviceName),
				JSONRPCPath: g.getJSONRPCPath(serviceName),
				Methods:     make([]*MethodData, 0),
			}
		}

		methodData := g.buildMethodData(info)
		serviceMap[serviceName].Methods = append(serviceMap[serviceName].Methods, methodData)

		if methodData.ReturnsError {
			serviceMap[serviceName].HasErrors = true
		}
	}

	// Convert map to slice
	for _, service := range serviceMap {
		data.Services = append(data.Services, service)
	}

	return data
}

// buildMethodData creates semantic data for a method
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

	// Set payload for non-notification methods that don't have streaming payload
	// SSE methods can have regular payload since they don't support streaming payload
	// Generate methods don't have payload
	if info.Modifier != ModifierNotify && info.Action != ActionGenerate && (!info.HasStreamingPayload() || info.IsSSE()) {
		data.Payload = g.buildTypeSpec(info.Type, info.Action, info.Modifier)
	}

	// Handle streaming
	if info.IsStreaming() {
		// Determine if bidirectional
		isBidirectional := info.IsWebSocket() && info.HasStreamingPayload() && info.HasStreamingResult()

		if info.HasStreamingPayload() {
			data.StreamingPayload = g.buildStreamingTypeSpec(info.Type, true, isBidirectional)
			data.StreamKind = "payload"
		}

		if info.HasStreamingResult() {
			data.Result = g.buildStreamingTypeSpec(info.Type, false, isBidirectional)
			if data.StreamKind == "payload" {
				data.StreamKind = "bidirectional"
			} else {
				data.StreamKind = "result"
			}

			// For SSE with final modifier, add ID field to result
			if info.IsSSE() && info.Modifier == ModifierFinal && data.Result != nil {
				data.Result.Fields = append(data.Result.Fields, FieldSpec{
					Position:    len(data.Result.Fields) + 1,
					Name:        "id",
					GoName:      "ID",
					Type:        &TypeSpec{Kind: "primitive", Primitive: "String"},
					Description: "Response ID (for final response)",
					Required:    false,
				})
			}
		}
	} else if info.Modifier != ModifierNotify && info.Modifier != ModifierError {
		// Non-streaming result
		data.Result = g.buildTypeSpec(info.Type, info.Action, "")
	}

	return data
}

// buildTypeSpec creates a TypeSpec based on the type string
func (g *Generator) buildTypeSpec(typeStr, _, modifier string) *TypeSpec {
	switch typeStr {
	case TypeString:
		// For validated primitives, wrap in object for JSON-RPC
		if modifier == ModifierValidate {
			return &TypeSpec{
				Kind: "object",
				Fields: []FieldSpec{
					{
						Position: 1,
						Name:     "value",
						GoName:   "Value",
						Type: &TypeSpec{
							Kind:      "primitive",
							Primitive: "String",
						},
						Required: true,
					},
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
func (g *Generator) buildStreamingTypeSpec(typeStr string, _ bool, isBidirectional bool) *TypeSpec {
	// For WebSocket bidirectional methods, we need ID fields
	if isBidirectional {
		switch typeStr {
		case TypeString:
			return &TypeSpec{
				Kind:    "object",
				NeedsID: true,
				Fields: []FieldSpec{
					{Position: 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true, Description: "Request/Response ID"},
					{Position: 2, Name: "value", GoName: "Value", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true, Description: "String value"},
				},
			}
		case TypeArray:
			return &TypeSpec{
				Kind:    "object",
				NeedsID: true,
				Fields: []FieldSpec{
					{Position: 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true, Description: "Request/Response ID"},
					{Position: 2, Name: "items", GoName: "Items", Type: &TypeSpec{Kind: "array", ArrayElem: &TypeSpec{Kind: "primitive", Primitive: "String"}}, Required: true},
				},
			}
		case TypeObject:
			return &TypeSpec{
				Kind:    "object",
				NeedsID: true,
				Fields: []FieldSpec{
					{Position: 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true, Description: "Request/Response ID"},
					{Position: 2, Name: "field1", GoName: "Field1", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true},
					{Position: 3, Name: "field2", GoName: "Field2", Type: &TypeSpec{Kind: "primitive", Primitive: "Int"}, Required: true},
					{Position: 4, Name: "field3", GoName: "Field3", Type: &TypeSpec{Kind: "primitive", Primitive: "Boolean"}, Required: true},
				},
			}
		default:
			return &TypeSpec{
				Kind:    "object",
				NeedsID: true,
				Fields: []FieldSpec{
					{Position: 1, Name: "id", GoName: "ID", Type: &TypeSpec{Kind: "primitive", Primitive: "String"}, Required: true, Description: "Request/Response ID"},
					{Position: 2, Name: "data", GoName: "Data", Type: &TypeSpec{Kind: "primitive", Primitive: "Any"}, Required: true},
				},
			}
		}
	}

	// For non-bidirectional streaming, wrap primitives in objects
	spec := g.buildTypeSpec(typeStr, "", "")
	if spec.Kind == "primitive" {
		return &TypeSpec{
			Kind: "object",
			Fields: []FieldSpec{
				{Position: 1, Name: "value", GoName: "Value", Type: spec, Required: true, Description: fmt.Sprintf("%s value", spec.Primitive)},
			},
		}
	}
	return spec
}

// buildImplementationData creates the semantic data for implementation
func (g *Generator) buildImplementationData(design *DesignData) *ImplementationData {
	data := &ImplementationData{
		PackageName: "testservice",
		Services:    make([]*ServiceImplData, 0),
	}

	for _, service := range design.Services {
		implService := &ServiceImplData{
			ServiceData:    service,
			ServicePackage: service.Name,
			Methods:        make([]*MethodImplData, 0),
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
	data := &MethodImplData{
		MethodData:     method,
		ServicePackage: serviceName,
		HasPayload:     method.Payload != nil || method.StreamingPayload != nil,
		HasResult:      method.Result != nil,
	}

	// Set type references
	if method.Payload != nil {
		if method.Payload.Kind == "primitive" {
			data.PayloadRef = strings.ToLower(method.Payload.Primitive)
		} else {
			data.PayloadRef = fmt.Sprintf("*%s.%sPayload", serviceName, method.GoName)
		}
	} else if method.StreamingPayload != nil && data.StreamKind == "bidirectional" {
		// For bidirectional methods with empty Payload(), Goa still generates a payload type
		data.PayloadRef = fmt.Sprintf("*%s.%sPayload", serviceName, method.GoName)
	}

	if method.Result != nil {
		if method.Result.Kind == "primitive" {
			data.ResultRef = strings.ToLower(method.Result.Primitive)
		} else {
			data.ResultRef = fmt.Sprintf("*%s.%sResult", serviceName, method.GoName)
		}
	}

	// Set stream interface
	if method.IsStreaming {
		data.StreamInterface = fmt.Sprintf("%sServerStream", method.GoName)
	}

	return data
}

// Files returns the list of files to generate
func (g *Generator) Files(design *DesignData, impl *ImplementationData) []*codegen.File {
	files := make([]*codegen.File, 0, 3+len(impl.Services))

	// go.mod file
	files = append(files, &codegen.File{
		Path: "go.mod",
		SectionTemplates: []*codegen.SectionTemplate{
			{
				Name:   "go-mod",
				Source: generatorTemplates.Read("go_mod"),
				Data:   map[string]string{"GoaPath": g.getGoaPath()},
			},
		},
	})

	// Design file
	files = append(files, &codegen.File{
		Path: filepath.Join("design", "design.go"),
		SectionTemplates: []*codegen.SectionTemplate{
			{
				Name:    "design",
				Source:  generatorTemplates.Read("dsl/design", "method", "type"),
				FuncMap: g.templateFuncs(),
				Data:    design,
			},
		},
	})

	// Service implementations
	for _, service := range impl.Services {
		// Build imports
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

		files = append(files, &codegen.File{
			Path:             fmt.Sprintf("%s.go", service.Name),
			SectionTemplates: sections,
		})
	}

	return files
}

// templateFuncs returns the template functions
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
		"dict": func(values ...any) map[string]any {
			if len(values)%2 != 0 {
				panic("dict requires even number of arguments")
			}
			dict := make(map[string]any)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					panic(fmt.Sprintf("dict keys must be strings, got %T", values[i]))
				}
				dict[key] = values[i+1]
			}
			return dict
		},
	}
}

// Helper methods
func (g *Generator) getServiceName(info MethodInfo) string {
	if info.IsSSE() {
		return "testsse"
	}
	if info.IsWebSocket() {
		return "testws"
	}
	return "test"
}

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

func (g *Generator) getMethodDescription(info MethodInfo) string {
	desc := fmt.Sprintf("%s %s", info.Action, info.Type)
	if info.Modifier != "" {
		desc += fmt.Sprintf(" (%s)", info.Modifier)
	}
	return desc
}

func (g *Generator) getGoaPath() string {
	// Get absolute path to the Goa root directory
	absPath, err := filepath.Abs("../../..")
	if err != nil {
		return "../../../.."
	}
	return absPath
}

func (g *Generator) runPostGeneration() error {
	goaBinary := g.getGoaBinary()

	// Run go mod tidy first
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = g.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w\nOutput: %s", err, output)
	}

	// Run goa gen
	cmd = exec.Command(goaBinary, "gen", "testservice/design", "-o", g.workDir)
	cmd.Dir = g.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("goa gen failed: %w\nOutput: %s", err, output)
	}

	// Run goa example
	cmd = exec.Command(goaBinary, "example", "testservice/design", "-o", g.workDir)
	cmd.Dir = g.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("goa example failed: %w\nOutput: %s", err, output)
	}

	// Run go mod tidy again to fix dependencies
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = g.workDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("final go mod tidy failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// getGoaBinary returns the path to the goa binary
// It checks for environment variables first, then falls back to system PATH
func (g *Generator) getGoaBinary() string {
	// Check for GOA_BINARY environment variable first
	if goaBinary := os.Getenv("GOA_BINARY"); goaBinary != "" {
		return goaBinary
	}

	// Check for GOPATH/bin/goa (common CI location)
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		goaBinName := "goa"
		if runtime.GOOS == "windows" {
			goaBinName = "goa.exe"
		}
		goaBin := filepath.Join(gopath, "bin", goaBinName)
		if _, err := os.Stat(goaBin); err == nil {
			return goaBin
		}
	}

	// Fall back to system PATH
	goaBinName := "goa"
	if runtime.GOOS == "windows" {
		goaBinName = "goa.exe"
	}
	return goaBinName
}

// goify converts a string to Go identifier
func goify(s string) string {
	return codegen.Goify(s, true)
}
