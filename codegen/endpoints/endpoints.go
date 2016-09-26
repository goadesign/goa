package genserver

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/goadesign/goa/codegen"
)

// GenerateController
func (g *Generator) generateControllers() error {
	ctlFile := filepath.Join(g.OutDir, "controllers.go")
	ctlWr, err := NewControllersWriter(ctlFile)
	if err != nil {
		panic(err) // bug
	}
	title := fmt.Sprintf("%s: Application Controllers", g.Root.Context())
	imports := []*codegen.ImportSpec{
		codegen.SimpleImport("net/http"),
		codegen.SimpleImport("fmt"),
		codegen.SimpleImport("context"),
		codegen.SimpleImport("github.com/goadesign/goa"),
		codegen.SimpleImport("github.com/goadesign/goa/cors"),
		codegen.SimpleImport("regexp"),
	}
	encoders, err := BuildEncoders(g.Root.Produces, true)
	if err != nil {
		return err
	}
	decoders, err := BuildEncoders(g.Root.Consumes, false)
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
	var packagePaths []string
	for packagePath := range encoderImports {
		if packagePath != "github.com/goadesign/goa" {
			packagePaths = append(packagePaths, packagePath)
		}
	}
	sort.Strings(packagePaths)
	for _, packagePath := range packagePaths {
		imports = append(imports, codegen.SimpleImport(packagePath))
	}
	ctlWr.WriteHeader(title, g.OutPkg, imports)
	ctlWr.WriteInitService(encoders, decoders)

	var controllersData []*ControllerTemplateData
	err = g.Root.IterateResources(func(r *design.ResourceDefinition) error {
		// Create file servers for all directory file servers that serve index.html.
		fileServers := r.FileServers
		for _, fs := range r.FileServers {
			if fs.IsDir() {
				rpath := design.WildcardRegex.ReplaceAllLiteralString(fs.RequestPath, "")
				rpath += "/"
				fileServers = append(fileServers, &design.FileServerDefinition{
					Parent:      fs.Parent,
					Description: fs.Description,
					Docs:        fs.Docs,
					FilePath:    filepath.Join(fs.FilePath, "index.html"),
					RequestPath: rpath,
					Metadata:    fs.Metadata,
					Security:    fs.Security,
				})
			}
		}
		data := &ControllerTemplateData{
			API:            g.Root,
			Resource:       codegen.Goify(r.Name, true),
			PreflightPaths: r.PreflightPaths(),
			FileServers:    fileServers,
		}
		ierr := r.IterateActions(func(a *design.ActionDefinition) error {
			context := fmt.Sprintf("%s%sContext", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			unmarshal := fmt.Sprintf("unmarshal%s%sPayload", codegen.Goify(a.Name, true), codegen.Goify(r.Name, true))
			action := map[string]interface{}{
				"Name":            codegen.Goify(a.Name, true),
				"Routes":          a.Routes,
				"Context":         context,
				"Unmarshal":       unmarshal,
				"Payload":         a.Payload,
				"PayloadOptional": a.PayloadOptional,
				"Security":        a.Security,
			}
			data.Actions = append(data.Actions, action)
			return nil
		})
		if ierr != nil {
			return ierr
		}
		if len(data.Actions) > 0 || len(data.FileServers) > 0 {
			data.Encoders = encoders
			data.Decoders = decoders
			data.Origins = r.AllOrigins()
			controllersData = append(controllersData, data)
		}
		return nil
	})
	if err != nil {
		return err
	}
	g.genfiles = append(g.genfiles, ctlFile)
	if err = ctlWr.Execute(controllersData); err != nil {
		return err
	}
	return ctlWr.FormatCode()
}
