package genapp

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	"github.com/goadesign/goa/goagen/utils"
	"github.com/spf13/cobra"
)

// Generator is the application code generator.
type Generator struct {
	genfiles []string
}

// Generate is the generator entry point called by the meta generator.
func Generate() (files []string, err error) {
	api := design.Design
	if err != nil {
		return nil, err
	}
	g := new(Generator)
	root := &cobra.Command{
		Use:   "goagen",
		Short: "Code generator",
		Long:  "application code generator",
		PreRunE: func(*cobra.Command, []string) error {
			outdir := AppOutputDir()
			os.RemoveAll(outdir)
			g.genfiles = []string{outdir}
			err = os.MkdirAll(outdir, 0777)
			return err
		},
		Run: func(*cobra.Command, []string) { files, err = g.Generate(api) },
	}
	codegen.RegisterFlags(root)
	NewCommand().RegisterFlags(root)
	root.Execute()
	return
}

// AppOutputDir returns the directory containing the generated files.
func AppOutputDir() string {
	return filepath.Join(codegen.OutputDir, TargetPackage)
}

// AppPackagePath returns the Go package path to the generated package.
func AppPackagePath() (string, error) {
	outputDir := AppOutputDir()
	gopaths := filepath.SplitList(os.Getenv("GOPATH"))
	for _, gopath := range gopaths {
		if strings.HasPrefix(outputDir, gopath) {
			path, err := filepath.Rel(filepath.Join(gopath, "src"), outputDir)
			if err != nil {
				return "", err
			}
			return filepath.ToSlash(path), nil
		}
	}
	return "", fmt.Errorf("output directory outside of Go workspace, make sure to define GOPATH correctly or change output directory")
}

// Generate the application code, implement codegen.Generator.
func (g *Generator) Generate(api *design.APIDefinition) (_ []string, err error) {
	if api == nil {
		return nil, fmt.Errorf("missing API definition, make sure design is properly initialized")
	}

	go utils.Catch(nil, func() { g.Cleanup() })

	defer func() {
		if err != nil {
			g.Cleanup()
		}
	}()

	os.RemoveAll(AppOutputDir())

	if err := os.MkdirAll(AppOutputDir(), 0755); err != nil {
		return nil, err
	}
	if err := g.generateContexts(api); err != nil {
		return nil, err
	}
	if err := g.generateControllers(api); err != nil {
		return nil, err
	}
	if err := g.generateSecurity(api); err != nil {
		return nil, err
	}
	if err := g.generateHrefs(api); err != nil {
		return nil, err
	}
	if err := g.generateMediaTypes(api); err != nil {
		return nil, err
	}
	if err := g.generateUserTypes(api); err != nil {
		return nil, err
	}

	return g.genfiles, nil
}

// Cleanup removes the entire "app" directory if it was created by this generator.
func (g *Generator) Cleanup() {
	if len(g.genfiles) == 0 {
		return
	}
	os.RemoveAll(AppOutputDir())
	g.genfiles = nil
}

// MergeResponses merge the response maps overriding the first argument map entries with the
// second argument map entries in case of collision.
func MergeResponses(l, r map[string]*design.ResponseDefinition) map[string]*design.ResponseDefinition {
	if l == nil {
		return r
	}
	if r == nil {
		return l
	}
	for n, r := range r {
		l[n] = r
	}
	return l
}

// generateContexts iterates through the API resources and actions and generates the action
// contexts.
func (g *Generator) generateContexts(api *design.APIDefinition) error {
	ctxFile := filepath.Join(AppOutputDir(), "contexts.go")
	ctxWr, err := NewContextsWriter(ctxFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Contexts", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("golang.org/x/net/context"),
		codegen.SimpleImport("strconv"),
		codegen.SimpleImport("strings"),
		codegen.SimpleImport("time"),
		codegen.SimpleImport("github.com/goadesign/goa"),
	}
	ctxWr.WriteHeader(title, TargetPackage, imports)
	err = api.IterateResources(func(r *design.ResourceDefinition) error {
		return r.IterateActions(func(a *design.ActionDefinition) error {
			ctxName := codegen.Goify(a.Name, true) + codegen.Goify(a.Parent.Name, true) + "Context"
			headers := r.Headers.Merge(a.Headers)
			if headers != nil && len(headers.Type.ToObject()) == 0 {
				headers = nil // So that {{if .Headers}} returns false in templates
			}
			params := a.AllParams()
			if params != nil && len(params.Type.ToObject()) == 0 {
				params = nil // So that {{if .Params}} returns false in templates
			}

			ctxData := ContextTemplateData{
				Name:         ctxName,
				ResourceName: r.Name,
				ActionName:   a.Name,
				Payload:      a.Payload,
				Params:       params,
				Headers:      headers,
				Routes:       a.Routes,
				Responses:    MergeResponses(r.Responses, a.Responses),
				API:          api,
				DefaultPkg:   TargetPackage,
				Security:     a.Security,
			}
			return ctxWr.Execute(&ctxData)
		})
	})
	g.genfiles = append(g.genfiles, ctxFile)
	if err != nil {
		return err
	}
	return ctxWr.FormatCode()
}

// BuildEncoders builds the template data needed to render the given encoding definitions.
// This extra map is needed to handle the case where a single encoding definition maps to multiple
// encoding packages. The data is indexed by mime type.
func BuildEncoders(info []*design.EncodingDefinition, encoder bool) ([]*EncoderTemplateData, error) {
	if len(info) == 0 {
		return nil, nil
	}
	// knownStdPackages lists the stdlib packages known by BuildEncoders
	var knownStdPackages = map[string]string{
		"encoding/json": "json",
		"encoding/xml":  "xml",
		"encoding/gob":  "gob",
	}
	encs := normalizeEncodingDefinitions(info)
	data := make([]*EncoderTemplateData, len(encs))
	defaultMediaType := info[0].MIMETypes[0]
	for i, enc := range encs {
		var pkgName string
		if name, ok := knownStdPackages[enc.PackagePath]; ok {
			pkgName = name
		} else {
			srcPath, err := codegen.PackageSourcePath(enc.PackagePath)
			if err != nil {
				return nil, fmt.Errorf("failed to locate package source of %s (%s)",
					enc.PackagePath, err)
			}
			pkgName, err = codegen.PackageName(srcPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load package %s (%s)",
					enc.PackagePath, err)
			}
		}
		isDefault := false
		for _, m := range enc.MIMETypes {
			if m == defaultMediaType {
				isDefault = true
			}
		}
		d := &EncoderTemplateData{
			PackagePath: enc.PackagePath,
			PackageName: pkgName,
			Function:    enc.Function,
			MIMETypes:   enc.MIMETypes,
			Default:     isDefault,
		}
		data[i] = d
	}
	return data, nil
}

// normalizeEncodingDefinitions figures out the package path and function of all encoding definitions
// and groups them by package and function name.
// We're going for simple rather than efficient (this is codegen after all)
// Also we assume that the encoding definitions have been validated: they have at least
// one mime type and definitions with no package path use known encoders.
func normalizeEncodingDefinitions(defs []*design.EncodingDefinition) []*design.EncodingDefinition {
	// First splat all definitions so each only have one mime type
	var encs []*design.EncodingDefinition
	for _, enc := range defs {
		if len(enc.MIMETypes) == 1 {
			encs = append(encs, enc)
			continue
		}
		for _, m := range enc.MIMETypes {
			encs = append(encs, &design.EncodingDefinition{
				MIMETypes:   []string{m},
				PackagePath: enc.PackagePath,
				Function:    enc.Function,
				Encoder:     enc.Encoder,
			})
		}
	}

	// Next make sure all definitions have a package path
	for _, enc := range encs {
		if enc.PackagePath == "" {
			enc.PackagePath = design.KnownEncoders[enc.MIMETypes[0]]
		}
		if enc.Function == "" {
			idx := 0
			if !enc.Encoder {
				idx = 1
			}
			enc.Function = design.KnownEncoderFunctions[enc.MIMETypes[0]][idx]
		}
	}

	// Regroup by package and function name
	byfn := make(map[string][]*design.EncodingDefinition)
	var first string
	for _, enc := range encs {
		key := enc.PackagePath + "#" + enc.Function
		if first == "" {
			first = key
		}
		if _, ok := byfn[key]; ok {
			byfn[key] = append(byfn[key], enc)
		} else {
			byfn[key] = []*design.EncodingDefinition{enc}
		}
	}

	// Reserialize into array keeping the first element identical since it's the default
	// encoder.
	return serialize(byfn, first)
}

func serialize(byfn map[string][]*design.EncodingDefinition, first string) []*design.EncodingDefinition {
	res := make([]*design.EncodingDefinition, len(byfn))
	i := 0
	keys := make([]string, len(byfn))
	for k := range byfn {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var idx int
	for j, k := range keys {
		if k == first {
			idx = j
			break
		}
	}
	keys[0], keys[idx] = keys[idx], keys[0]
	i = 0
	for _, key := range keys {
		encs := byfn[key]
		res[i] = &design.EncodingDefinition{
			MIMETypes:   encs[0].MIMETypes,
			PackagePath: encs[0].PackagePath,
			Function:    encs[0].Function,
		}
		if len(encs) > 0 {
			encs = encs[1:]
			for _, enc := range encs {
				for _, m := range enc.MIMETypes {
					found := false
					for _, rm := range res[i].MIMETypes {
						if m == rm {
							found = true
							break
						}
					}
					if !found {
						res[i].MIMETypes = append(res[i].MIMETypes, m)
					}
				}
			}
		}
		i++
	}
	return res
}

// generateControllers iterates through the API resources and generates the low level
// controllers.
func (g *Generator) generateControllers(api *design.APIDefinition) error {
	ctlFile := filepath.Join(AppOutputDir(), "controllers.go")
	ctlWr, err := NewControllersWriter(ctlFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Controllers", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("golang.org/x/net/context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/cors"),
	}
	encoders, err := BuildEncoders(api.Produces, true)
	if err != nil {
		return err
	}
	decoders, err := BuildEncoders(api.Consumes, false)
	if err != nil {
		return err
	}
	encoderImports := make(map[string]bool)
	for _, data := range encoders {
		encoderImports[data.PackagePath] = true
	}
	for _, data := range decoders {
		encoderImports[data.PackagePath] = true
	}
	for packagePath := range encoderImports {
		if !strings.Contains(packagePath, "/goadesign/goa/") {
			imports = append(imports, codegen.SimpleImport(packagePath))
		}
	}
	ctlWr.WriteHeader(title, TargetPackage, imports)
	ctlWr.WriteInitService(encoders, decoders)

	var controllersData []*ControllerTemplateData
	api.IterateResources(func(r *design.ResourceDefinition) error {
		data := &ControllerTemplateData{API: api, Resource: codegen.Goify(r.Name, true)}
		err := r.IterateActions(func(a *design.ActionDefinition) error {
			context := fmt.Sprintf("%s%sContext", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			unmarshal := fmt.Sprintf("unmarshal%s%sPayload", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			action := map[string]interface{}{
				"Name":           codegen.Goify(a.Name, true),
				"Routes":         a.Routes,
				"Context":        context,
				"Unmarshal":      unmarshal,
				"Payload":        a.Payload,
				"Security":       a.Security,
				"PreflightPaths": a.PreflightPaths(),
			}
			data.Actions = append(data.Actions, action)
			return nil
		})
		if err != nil {
			return err
		}
		if len(data.Actions) > 0 {
			data.Encoders = encoders
			data.Decoders = decoders
			data.Origins = r.AllOrigins()
			controllersData = append(controllersData, data)
		}
		return nil
	})
	g.genfiles = append(g.genfiles, ctlFile)
	if err = ctlWr.Execute(controllersData); err != nil {
		return err
	}
	return ctlWr.FormatCode()
}

// generateControllers iterates through the API resources and generates the low level
// controllers.
func (g *Generator) generateSecurity(api *design.APIDefinition) error {
	if len(api.SecuritySchemes) == 0 {
		return nil
	}

	secFile := filepath.Join(AppOutputDir(), "security.go")
	secWr, err := NewSecurityWriter(secFile)
	if err != nil {
		panic(err) // bug
	}

	title := fmt.Sprintf("%s: Application Security", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("errors"),
		codegen.SimpleImport("golang.org/x/net/context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
	}
	secWr.WriteHeader(title, TargetPackage, imports)

	g.genfiles = append(g.genfiles, secFile)

	if err = secWr.Execute(design.Design.SecuritySchemes); err != nil {
		return err
	}

	return secWr.FormatCode()
}

// generateHrefs iterates through the API resources and generates the href factory methods.
func (g *Generator) generateHrefs(api *design.APIDefinition) error {
	hrefFile := filepath.Join(AppOutputDir(), "hrefs.go")
	resWr, err := NewResourcesWriter(hrefFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Resource Href Factories", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("fmt"),
	}
	resWr.WriteHeader(title, TargetPackage, imports)
	err = api.IterateResources(func(r *design.ResourceDefinition) error {
		m := api.MediaTypeWithIdentifier(r.MediaType)
		var identifier string
		if m != nil {
			identifier = m.Identifier
		} else {
			identifier = "plain/text"
		}
		canoTemplate := r.URITemplate()
		canoTemplate = design.WildcardRegex.ReplaceAllLiteralString(canoTemplate, "/%v")
		var canoParams []string
		if ca := r.CanonicalAction(); ca != nil {
			if len(ca.Routes) > 0 {
				canoParams = ca.Routes[0].Params()
			}
		}
		data := ResourceData{
			Name:              codegen.Goify(r.Name, true),
			Identifier:        identifier,
			Description:       r.Description,
			Type:              m,
			CanonicalTemplate: canoTemplate,
			CanonicalParams:   canoParams,
		}
		return resWr.Execute(&data)
	})
	g.genfiles = append(g.genfiles, hrefFile)
	if err != nil {
		return err
	}
	return resWr.FormatCode()
}

// generateMediaTypes iterates through the media types and generate the data structures and
// marshaling code.
func (g *Generator) generateMediaTypes(api *design.APIDefinition) error {
	mtFile := filepath.Join(AppOutputDir(), "media_types.go")
	mtWr, err := NewMediaTypesWriter(mtFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Media Types", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
	}
	mtWr.WriteHeader(title, TargetPackage, imports)
	err = api.IterateMediaTypes(func(mt *design.MediaTypeDefinition) error {
		if mt.Type.IsObject() || mt.Type.IsArray() {
			return mtWr.Execute(mt)
		}
		return nil
	})
	g.genfiles = append(g.genfiles, mtFile)
	if err != nil {
		return err
	}
	return mtWr.FormatCode()
}

// generateUserTypes iterates through the user types and generates the data structures and
// marshaling code.
func (g *Generator) generateUserTypes(api *design.APIDefinition) error {
	utFile := filepath.Join(AppOutputDir(), "user_types.go")
	utWr, err := NewUserTypesWriter(utFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application User Types", api.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("time"),
	}
	utWr.WriteHeader(title, TargetPackage, imports)
	err = api.IterateUserTypes(func(t *design.UserTypeDefinition) error {
		return utWr.Execute(t)
	})
	g.genfiles = append(g.genfiles, utFile)
	if err != nil {
		return err
	}
	return utWr.FormatCode()
}
