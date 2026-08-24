// This file runs the public plugin registration APIs in fresh child processes.
// Each child uses the real default registry, so the tests cover rejecting late
// registrations and running released callbacks across generation commands.
package generator

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	servicecodegen "goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
	httpcodegen "goa.design/goa/v3/http/codegen"
)

type (
	// publicHTTPPluginOrder gives declarations in the compile test a stable
	// order within the generated server package.
	publicHTTPPluginOrder string

	// publicHTTPPluginData supplies three generated names to the plugin source
	// template after Goa has chosen all generated names.
	publicHTTPPluginData struct {
		// Wrapper is the handler wrapper name chosen with core server names.
		Wrapper *codegen.NameDeclaration
		// EndpointWrapper is the private wrapper chosen for the Read endpoint.
		EndpointWrapper *codegen.NameDeclaration
		// Mount is the extra mount function name chosen with core server names.
		Mount *codegen.NameDeclaration
	}
)

const publicPluginChildMode = "GOA_PUBLIC_PLUGIN_CHILD"

// TestPublicPluginRegistrationUsesDefaultGenerationRun verifies released and
// factory registrations using fresh package globals in separate processes.
func TestPublicPluginRegistrationUsesDefaultGenerationRun(t *testing.T) {
	switch os.Getenv(publicPluginChildMode) {
	case "run":
		runPublicPluginChild(t)
		return
	case "duplicate":
		runPublicPluginDuplicateChild(t)
		return
	case "repeat":
		runPublicPluginRepeatedRunChild(t)
		return
	case "http-extension":
		runPublicHTTPServerExtensionChild(t)
		return
	case "legacy-http-endpoint":
		runPublicLegacyHTTPEndpointChild(t)
		return
	}

	for _, mode := range []string{"run", "duplicate", "repeat", "http-extension", "legacy-http-endpoint"} {
		t.Run(mode, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestPublicPluginRegistrationUsesDefaultGenerationRun$")
			command.Env = append(os.Environ(), publicPluginChildMode+"="+mode)
			output, err := command.CombinedOutput()
			require.NoErrorf(t, err, "child process failed:\n%s", output)
		})
	}
}

// runPublicLegacyHTTPEndpointChild checks that a released plugin can add an
// endpoint using the public handler name without knowing Goa's private plan.
func runPublicLegacyHTTPEndpointChild(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Read", func() {
				dsl.HTTP(func() {
					dsl.GET("/items")
				})
			})
		})
	})
	codegen.RegisterPlugin("legacy-http-endpoint", "gen", nil, func(_ string, _ []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
		for _, file := range files {
			if file.Path != filepath.Join(codegen.Gendir, "http", "calc", "server", "server.go") {
				continue
			}
			for _, section := range file.SectionTemplates {
				if section.Name != "server-init" {
					continue
				}
				data := section.Data.(*httpcodegen.ServiceData)
				data.Endpoints = append(data.Endpoints, &httpcodegen.EndpointData{
					Method:       &servicecodegen.MethodData{VarName: "CORS"},
					MountHandler: "MountCORSHandler",
					HandlerInit:  "NewCORSHandler",
				})
				section.Source = strings.ReplaceAll(
					section.Source,
					`e.{{ .Method.VarName }}, mux, {{ if .MultipartRequestDecoder }}{{ .MultipartRequestDecoder.InitName }}(mux, {{ .MultipartRequestDecoder.VarName }}){{ else }}decoder{{ end }}, encoder, errhandler, formatter{{ if isWebSocketEndpoint . }}, upgrader, configurer.{{ .Method.VarName }}Fn{{ end }})`,
					`{{ if ne .Method.VarName "CORS" }}e.{{ .Method.VarName }}, mux, {{ if .MultipartRequestDecoder }}{{ .MultipartRequestDecoder.InitName }}(mux, {{ .MultipartRequestDecoder.VarName }}){{ else }}decoder{{ end }}, encoder, errhandler, formatter{{ if isWebSocketEndpoint . }}, upgrader, configurer.{{ .Method.VarName }}Fn{{ end }}{{ end }})`,
				)
			}
		}
		return files, nil
	})

	run, err := newGenerationRun("gen", defaultRegistry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	for _, file := range result.files {
		if file.Path != filepath.Join(codegen.Gendir, "http", "calc", "server", "server.go") {
			continue
		}
		code := codegen.SectionsCode(t, file.Section("server-init"))
		require.Contains(t, code, "CORS: NewCORSHandler()")
		mount := codegen.SectionsCode(t, file.Section("server-mount"))
		require.Contains(t, mount, "MountCORSHandler(mux, h.CORS)")
		return
	}
	t.Fatal("generated HTTP server file is missing")
}

// runPublicPluginRepeatedRunChild registers through the released API once and
// checks that the same functions receive each later run's package and root.
func runPublicPluginRepeatedRunChild(t *testing.T) {
	var prepared, generated []string
	codegen.RegisterPlugin(
		"repeat",
		"gen",
		func(genpkg string, roots []eval.Root) error {
			prepared = append(prepared, genpkg+":"+roots[0].(*expr.RootExpr).API.Name)
			return nil
		},
		func(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
			name := roots[0].(*expr.RootExpr).API.Name
			generated = append(generated, genpkg+":"+name)
			return append(files, &codegen.File{Path: "released-" + name}), nil
		},
	)

	packages := []string{"generated.local/first", "generated.local/second"}
	for index, name := range []string{"first", "second"} {
		root := expr.RunDSL(t, func() {
			dsl.API(name, func() {
			})
		})
		run, err := newGenerationRun("gen", defaultRegistry)
		require.NoError(t, err)
		result, err := run.execute(packages[index], []eval.Root{root})
		require.NoError(t, err)
		require.Equal(t, "released-"+name, result.files[len(result.files)-1].Path)
	}

	require.Equal(t, []string{
		"generated.local/first:first",
		"generated.local/second:second",
	}, prepared)
	require.Equal(t, prepared, generated)
}

// runPublicHTTPServerExtensionChild generates and compiles an HTTP service with
// a public per-run plugin that defines both extension function bodies.
func runPublicHTTPServerExtensionChild(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		dsl.Service("Calc", func() {
			dsl.Method("Read", func() {
				dsl.Payload(func() {
					dsl.Attribute("id", dsl.String)
				})
				dsl.HTTP(func() {
					dsl.GET("/items/{id}")
				})
			})
		})
	})
	RegisterPlugin("http-server-extension", "gen", func() Plugin {
		data := &publicHTTPPluginData{}
		return Plugin{
			Plan: func(plan *Plan) error {
				httpPlan, ok := plan.HTTP(root)
				if !ok {
					return fmt.Errorf("ordinary HTTP plan is missing")
				}
				service := root.API.HTTP.Services[0]
				var err error
				data.Wrapper, err = httpPlan.DeclareServerHandlerWrapper(service, "WrapExtension", publicHTTPPluginOrder("wrapper"))
				if err != nil {
					return err
				}
				data.EndpointWrapper, err = httpPlan.DeclareServerEndpointHandlerWrapper(service.HTTPEndpoints[0], "wrapReadExtension", publicHTTPPluginOrder("endpoint wrapper"))
				if err != nil {
					return err
				}
				data.Mount, err = httpPlan.DeclareServerMount(service, "MountExtension", publicHTTPPluginOrder("mount"), []httpcodegen.ServerMountPoint{{
					Method:  "Extension preflight",
					Verb:    "OPTIONS",
					Pattern: "/items/{id}",
				}})
				return err
			},
			Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
				return append(files, publicHTTPServerExtensionFile(data)), nil
			},
		}
	})

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "generated.local/gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	serverSource, err := os.ReadFile(filepath.Join(genDir, "http", "calc", "server", "server.go"))
	require.NoError(t, err)
	require.Contains(t, string(serverSource), "h = WrapExtension(wrapReadExtension(h))")
	require.Contains(t, string(serverSource), "MountReadHandler(mux, h.Read)")
	require.NotContains(t, string(serverSource), "MountReadHandler(mux, WrapExtension")
	runGeneratedTests(t, genDir)
}

// runPublicPluginChild mixes both public APIs and checks the arguments and
// files passed through the real default generation run.
func runPublicPluginChild(t *testing.T) {
	root := expr.RunDSL(t, func() {})
	var events []string
	registerPublicFactoryPlugin("a-first", pluginFirst, &events)
	registerPublicReleasedPlugin("z-first", pluginFirst, root, &events)
	registerPublicReleasedPlugin("a-normal", pluginNormal, root, &events)
	registerPublicFactoryPlugin("z-normal", pluginNormal, &events)
	registerPublicReleasedPlugin("a-last", pluginLast, root, &events)
	registerPublicFactoryPlugin("z-last", pluginLast, &events)

	run, err := newGenerationRun("gen", defaultRegistry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	require.Equal(t, []string{
		"prepare:a-first", "prepare:z-first", "prepare:a-normal", "prepare:z-normal", "prepare:a-last", "prepare:z-last",
		"generate:a-first", "generate:z-first:factory:a-first", "generate:a-normal:released:z-first",
		"generate:z-normal:released:a-normal", "generate:a-last:factory:z-normal", "generate:z-last:released:a-last",
	}, events)
	require.Equal(t, "factory:z-last", result.files[len(result.files)-1].Path)
	require.PanicsWithValue(t, "plugin registry is sealed", func() {
		codegen.RegisterPlugin("late", "gen", nil, publicUnchangedFiles)
	})
	require.PanicsWithValue(t, "generator plugin registry is sealed", func() {
		RegisterPlugin("late", "gen", func() Plugin {
			return Plugin{}
		})
	})
}

// runPublicPluginDuplicateChild proves that registrations for another command
// do not block or run during this command.
func runPublicPluginDuplicateChild(t *testing.T) {
	called := false
	codegen.RegisterPlugin("duplicate", "example", nil, func(_ string, _ []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
		called = true
		return files, nil
	})
	RegisterPlugin("duplicate", "example", func() Plugin {
		called = true
		return Plugin{}
	})

	_, err := newGenerationRun("gen", defaultRegistry)
	require.NoError(t, err)
	require.False(t, called)
}

// registerPublicReleasedPlugin adds one old-style callback pair through the
// exact API used by released Goa v3 plugins.
func registerPublicReleasedPlugin(name string, position pluginPosition, root eval.Root, events *[]string) {
	prepare := func(genpkg string, roots []eval.Root) error {
		if genpkg != "generated.local/gen" {
			return fmt.Errorf("prepare received package %q", genpkg)
		}
		if len(roots) != 1 || roots[0] != root {
			return fmt.Errorf("prepare received another run's roots")
		}
		*events = append(*events, "prepare:"+name)
		return nil
	}
	generate := func(genpkg string, roots []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
		if genpkg != "generated.local/gen" {
			return nil, fmt.Errorf("generate received package %q", genpkg)
		}
		if len(roots) != 1 || roots[0] != root {
			return nil, fmt.Errorf("generate received another run's roots")
		}
		*events = append(*events, "generate:"+name+":"+files[len(files)-1].Path)
		return append(files, &codegen.File{Path: "released:" + name}), nil
	}
	switch position {
	case pluginFirst:
		codegen.RegisterPluginFirst(name, "gen", prepare, generate)
	case pluginNormal:
		codegen.RegisterPlugin(name, "gen", prepare, generate)
	case pluginLast:
		codegen.RegisterPluginLast(name, "gen", prepare, generate)
	}
}

// registerPublicFactoryPlugin adds one planning-aware plugin through the new
// API and records the file left by the preceding plugin.
func registerPublicFactoryPlugin(name string, position pluginPosition, events *[]string) {
	factory := func() Plugin {
		return Plugin{
			Prepare: func(_ string, _ []eval.Root) error {
				*events = append(*events, "prepare:"+name)
				return nil
			},
			Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
				event := "generate:" + name
				if len(files) > 0 {
					event += ":" + files[len(files)-1].Path
				}
				*events = append(*events, event)
				return append(files, &codegen.File{Path: "factory:" + name}), nil
			},
		}
	}
	switch position {
	case pluginFirst:
		RegisterPluginFirst(name, "gen", factory)
	case pluginNormal:
		RegisterPlugin(name, "gen", factory)
	case pluginLast:
		RegisterPluginLast(name, "gen", factory)
	}
}

// publicUnchangedFiles is a valid callback used to test late registration.
func publicUnchangedFiles(_ string, _ []eval.Root, files []*codegen.File) ([]*codegen.File, error) {
	return files, nil
}

// publicHTTPServerExtensionFile writes the two functions promised during
// plugin planning into the generated Calc server package.
func publicHTTPServerExtensionFile(data *publicHTTPPluginData) *codegen.File {
	return &codegen.File{
		Path: filepath.Join(codegen.Gendir, "http", "calc", "server", "plugin.go"),
		SectionTemplates: []*codegen.SectionTemplate{
			codegen.Header("Calc HTTP server plugin", "server", []*codegen.ImportSpec{
				codegen.SimpleImport("net/http"),
				codegen.GoaNamedImport("http", "goahttp"),
			}),
			{
				Name: "http-server-extension",
				Source: `// {{ .Wrapper.Name }} wraps a handler mounted from the Calc design.
func {{ .Wrapper.Name }}(handler http.Handler) http.Handler {
	return handler
}

// {{ .EndpointWrapper.Name }} wraps only the Read endpoint handler.
func {{ .EndpointWrapper.Name }}(handler http.Handler) http.Handler {
	return handler
}

// {{ .Mount.Name }} adds the Calc preflight route.
func {{ .Mount.Name }}(mux goahttp.Muxer) {
	mux.Handle("OPTIONS", "/items/{id}", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
}`,
				Data: data,
			},
		},
	}
}

// ComparePackageName gives public plugin declarations a stable order.
func (o publicHTTPPluginOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	return cmp.Compare(string(o), string(other.(publicHTTPPluginOrder)))
}
