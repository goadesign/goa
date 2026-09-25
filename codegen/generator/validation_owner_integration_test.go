// This file compiles private plugin validators against the original service
// types. The plugin keeps Goa's layouts and validation rules through rendering.
package generator

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	d "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

type (
	validationFixture struct {
		pkg         *codegen.GeneratedPackage
		imports     *codegen.GeneratedImportPlan
		bind        codegen.GoTypeBinder
		values      []*validationFixtureValue
		definitions map[validationFixtureKey]*validationFixtureValue
		roots       map[string]*validationFixtureValue
	}

	validationFixtureKey struct {
		definition  *expr.AttributeExpr
		declaration *codegen.TypeDeclaration
		policy      codegen.GoLayoutPolicy
	}

	validationFixtureValue struct {
		named       *codegen.GoTypePlan
		definition  *codegen.GoTypePlan
		validation  *codegen.ValidationPlan
		declaration *codegen.NameDeclaration
	}

	validationFixtureOrder int

	validationFixtureFunction struct {
		Name, Ref, Body string
	}
)

func TestGeneratePrivateValueValidation(t *testing.T) {
	root := codegen.RunDSL(t, validationOwnerDSL)
	fixture := &validationFixture{
		definitions: make(map[validationFixtureKey]*validationFixtureValue),
		roots:       make(map[string]*validationFixtureValue),
	}
	registry := newDefaultRegistry()
	registry.registerPlugin("private-value-validation", "gen", pluginNormal, func() Plugin {
		return Plugin{
			Plan: func(plan *Plan) error {
				return fixture.plan(plan.Generation(), root)
			},
			Generate: func(_ *Plan, files []*codegen.File) ([]*codegen.File, error) {
				if err := fixture.imports.Link(); err != nil {
					return nil, err
				}
				file, err := fixture.file()
				if err != nil {
					return nil, err
				}
				return append(files, file), nil
			},
		}
	})
	run, err := newGenerationRun("gen", registry)
	require.NoError(t, err)
	result, err := run.execute("generated.local/gen", []eval.Root{root})
	require.NoError(t, err)
	counts := make(map[string]int)
	for _, value := range fixture.values {
		require.Equal(t, "generated.local/gen/types", value.named.Owner())
		counts[value.named.TypeDeclaration().Name()]++
	}
	require.Equal(t, map[string]int{
		"Container": 1, "Entry": 1, "Node": 1, "NodeList": 1,
		"NodeMap": 1, "Branch": 1, "MapNode": 1, "SliceNode": 1,
	}, counts, "recursive calls share their active definition")

	dir := t.TempDir()
	writeGeneratedModule(t, dir, "generated.local")
	for _, file := range result.files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	types := result.plan.Generation().Package("generated.local/gen/types")
	choice := expr.AsObject(root.UserType("Node").Attribute().Type).Attribute("choice")
	branch, err := types.UnionBranch(choice, "node")
	require.NoError(t, err)
	data := map[string]string{
		"NodeChoice": fixture.pkg.ImportName(types.ImportPath()) + "." + branch.Constructor(),
	}
	for name, value := range fixture.roots {
		data[name+"Type"] = value.named.Link(fixture.pkg.ImportPath(), fixture.pkg.ImportName).Name()
		data[name+"Validator"] = value.declaration.Name()
	}
	for _, name := range []string{"Entry", "Branch", "NodeList", "NodeMap"} {
		declaration, err := types.Type(root.UserType(name))
		require.NoError(t, err)
		data[name+"Type"] = fixture.pkg.ImportName(types.ImportPath()) + "." + declaration.Name()
	}
	testFile := &codegen.File{
		Path: filepath.Join(fixture.pkg.OutputDirectory(), "validation_test.go"),
		SectionTemplates: []*codegen.SectionTemplate{
			codegen.Header("Private validation tests", "checks", []*codegen.ImportSpec{
				codegen.SimpleImport("testing"), fixture.pkg.Import(types.ImportPath()),
			}),
			{Name: "validation-tests", Source: validationOwnerTests, Data: data},
		},
	}
	_, err = testFile.Render(dir)
	require.NoError(t, err)

	// Once generation has finished, changed expressions cannot change the
	// saved parameter types, fields or validation rules.
	before, err := os.ReadFile(filepath.Join(dir, fixture.pkg.OutputDirectory(), "validation.go"))
	require.NoError(t, err)
	for _, userType := range root.Types {
		userType.SetAttribute(&expr.AttributeExpr{Type: expr.String})
	}
	retained, err := fixture.file()
	require.NoError(t, err)
	retainedDir := t.TempDir()
	_, err = retained.Render(retainedDir)
	require.NoError(t, err)
	after, err := os.ReadFile(filepath.Join(retainedDir, retained.Path))
	require.NoError(t, err)
	require.Equal(t, before, after)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "-v", "./...")
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off")
	output, err := command.CombinedOutput()
	t.Logf("generated validation command: go test -mod=mod -v ./...\n%s", output)
	require.NoError(t, err)
}

// plan uses the same original declarations that the service generator emits.
// Only the functions and their imports belong to the plugin's checks package.
func (f *validationFixture) plan(generation *codegen.Generation, root *expr.RootExpr) error {
	var err error
	f.pkg, err = generation.ClaimPackage(path.Join(generation.GenPkg(), "checks"))
	if err != nil {
		return err
	}
	f.imports = codegen.NewGeneratedImportPlan(f.pkg)
	f.bind = func(request codegen.GoTypeBindingRequest) (codegen.GoTypeBinding, error) {
		owner := request.InheritedOwner
		if location := codegen.UserTypeLocation(request.Attribute.Type); location != nil {
			owner = path.Join(generation.GenPkg(), location.RelImportPath)
		}
		pkg := generation.Package(owner)
		binding := codegen.GoTypeBinding{Owner: owner}
		var err error
		switch request.Kind {
		case codegen.GoNamed:
			binding.Type, err = pkg.Type(request.Attribute.Type.(expr.UserType))
		case codegen.GoUnion:
			binding.Union, err = pkg.Union(request.Attribute)
		default:
			return binding, fmt.Errorf("unexpected named binding %s", request.Kind)
		}
		return binding, err
	}
	for _, name := range []string{"Container", "Node", "MapNode", "SliceNode"} {
		attribute := &expr.AttributeExpr{Type: root.UserType(name)}
		layout, err := codegen.PlanGoType(attribute, codegen.GoTypePlanOptions{
			Owner:  path.Join(generation.GenPkg(), "types"),
			Policy: codegen.GoLayoutPolicy{UseDefault: true, SumType: true},
			Bind:   f.bind,
		})
		if err != nil {
			return err
		}
		value, err := f.add(attribute, layout)
		if err != nil {
			return err
		}
		f.roots[name] = value
	}
	return nil
}

// add reserves a function before planning its body. Recursive calls therefore
// reuse the exact active definition and policy instead of expanding it again.
func (f *validationFixture) add(attribute *expr.AttributeExpr, layout *codegen.GoTypePlan) (*validationFixtureValue, error) {
	userType := attribute.Type.(expr.UserType)
	key := validationFixtureKey{userType.Attribute(), layout.TypeDeclaration(), layout.Policy()}
	if value := f.definitions[key]; value != nil {
		return value, nil
	}
	definition, err := codegen.PlanGoType(userType.Attribute(), codegen.GoTypePlanOptions{
		Owner: layout.Owner(), Policy: layout.Policy(), Bind: f.bind,
	})
	if err != nil {
		return nil, err
	}
	declaration, err := f.pkg.DeclareDependentName(codegen.NameFunction, layout.TypeDeclaration().Declaration(),
		"validate", "", validationFixtureOrder(len(f.values)))
	if err != nil {
		return nil, err
	}
	value := &validationFixtureValue{named: layout, definition: definition, declaration: declaration}
	f.definitions[key] = value
	f.values = append(f.values, value)
	value.validation, err = codegen.NewValidationPlan(userType.Attribute(), definition, codegen.ValidationPlanOptions{
		Required: true,
		Alias:    expr.IsAlias(userType),
		Bind: func(request codegen.ValidatorBindingRequest) (*codegen.NameDeclaration, error) {
			nested, err := f.add(request.Attribute, request.Layout)
			if err != nil {
				return nil, err
			}
			return nested.declaration, nil
		},
	})
	if err != nil {
		return nil, err
	}
	for _, spec := range layout.ImportPreferences() {
		if err := f.imports.AddGenerated(codegen.NewImport("gentypes", spec.Path)); err != nil {
			return nil, err
		}
	}
	for _, spec := range value.validation.ImportPreferences() {
		if spec.Path == codegen.GoaImport("").Path || spec.Path == "unicode/utf8" {
			err = f.imports.Require(codegen.NewImport(spec.Name, spec.Path))
		} else {
			err = f.imports.AddGenerated(codegen.NewImport(spec.Name, spec.Path))
		}
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

// file renders only retained layouts, rules and completed declaration names.
func (f *validationFixture) file() (*codegen.File, error) {
	functions := make([]validationFixtureFunction, 0, len(f.values))
	for _, value := range f.values {
		body, err := value.validation.Link(value.definition.Link(f.pkg.ImportPath(), f.pkg.ImportName))
		if err != nil {
			return nil, err
		}
		functions = append(functions, validationFixtureFunction{
			Name: value.declaration.Name(),
			Ref:  value.named.Link(f.pkg.ImportPath(), f.pkg.ImportName).Ref(),
			Body: body.Render("value", "value"),
		})
	}
	return &codegen.File{
		Path: filepath.Join(f.pkg.OutputDirectory(), "validation.go"),
		SectionTemplates: []*codegen.SectionTemplate{
			codegen.Header("Private original-value validation", "checks", f.imports.Imports()),
			{Name: "validators", Source: `{{range .}}
func {{.Name}}(value {{.Ref}}) (err error) {
	{{.Body}}
	return
}
{{end}}`, Data: functions},
		},
	}, nil
}

// ComparePackageName orders fixture helpers by their explicit planning order.
func (o validationFixtureOrder) ComparePackageName(other codegen.PackageNameOrder) int {
	return cmp.Compare(o, other.(validationFixtureOrder))
}

func validationOwnerDSL() {
	d.API("original-values", func() {})
	located := func() {
		d.Meta("struct:pkg:path", "types")
		d.Meta("type:generate:force")
	}
	entry := d.Type("Entry", func() {
		located()
		d.Attribute("label", d.String, func() {
			d.MinLength(1)
		})
		d.Required("label")
	})
	container := d.Type("Container", func() {
		located()
		d.Attribute("single", entry)
		d.Attribute("loose", d.ArrayOf(entry))
		d.Attribute("tight", d.ArrayOfRequired(entry))
		d.Attribute("count", d.Int, func() {
			d.Default(5)
		})
		d.Attribute("enabled", d.Boolean, func() {
			d.Default(true)
		})
		d.Attribute("text", d.String, func() {
			d.Default("fallback")
		})
		d.Required("single", "loose", "tight", "count", "enabled", "text")
	})
	d.Type("Branch", func() {
		located()
		d.Attribute("nodes", "NodeList")
	})
	d.Type("Node", func() {
		located()
		d.Attribute("label", d.String, func() {
			d.MinLength(1)
		})
		d.Attribute("children", "NodeList")
		d.Attribute("by_name", "NodeMap")
		d.Attribute("branch", "Branch")
		d.OneOf("choice", func() {
			d.Attribute("node", "Node")
			d.Attribute("text", d.String)
		})
		d.Required("label")
	})
	d.Type("NodeList", d.ArrayOf("Node"), located)
	d.Type("NodeMap", d.MapOf(d.String, "Node"), located)
	// Register each collection before connecting its element to itself.
	mapNode := d.Type("MapNode", func() {
		located()
		d.MaxLength(2)
	})
	mapNode.Attribute().Type = d.MapOf(d.String, mapNode)
	sliceNode := d.Type("SliceNode", func() {
		located()
		d.MaxLength(2)
	})
	sliceNode.Attribute().Type = d.ArrayOf(sliceNode)
	d.Service("records", func() {
		d.Method("store", func() {
			d.Payload(container)
		})
	})
}

const validationOwnerTests = `
func TestRequiredOriginalValues(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(*{{.ContainerType}})
		invalid bool
	}{
		{name: "empty collections and zero values"},
		{name: "nullable child", edit: func(v *{{.ContainerType}}) {
			v.Loose = []*{{.EntryType}}{nil}
		}},
		{name: "required object", edit: func(v *{{.ContainerType}}) {
			v.Single = nil
		}, invalid: true},
		{name: "required nullable collection", edit: func(v *{{.ContainerType}}) {
			v.Loose = nil
		}, invalid: true},
		{name: "required strict collection", edit: func(v *{{.ContainerType}}) {
			v.Tight = nil
		}, invalid: true},
		{name: "strict null element", edit: func(v *{{.ContainerType}}) {
			v.Tight = []*{{.EntryType}}{nil}
		}, invalid: true},
		{name: "nested constraint", edit: func(v *{{.ContainerType}}) {
			v.Single.Label = ""
		}, invalid: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := &{{.ContainerType}}{Single: &{{.EntryType}}{Label: "valid"}, Loose: []*{{.EntryType}}{}, Tight: []*{{.EntryType}}{}}
			if test.edit != nil {
				test.edit(value)
			}
			if err := {{.ContainerValidator}}(value); (err != nil) != test.invalid {
				t.Fatalf("validation error = %v, invalid = %v", err, test.invalid)
			}
		})
	}
}

func TestFiniteRecursiveOriginalValues(t *testing.T) {
	for _, test := range []struct {
		name string
		connect func(*{{.NodeType}}, *{{.NodeType}})
	}{
		{name: "named slice", connect: func(root, leaf *{{.NodeType}}) {
			root.Children = {{.NodeListType}}{leaf}
		}},
		{name: "named map", connect: func(root, leaf *{{.NodeType}}) {
			root.ByName = {{.NodeMapType}}{"leaf": leaf}
		}},
		{name: "mutual object", connect: func(root, leaf *{{.NodeType}}) {
			root.Branch = &{{.BranchType}}{Nodes: {{.NodeListType}}{leaf}}
		}},
		{name: "union", connect: func(root, leaf *{{.NodeType}}) {
			root.Choice = {{.NodeChoice}}(leaf)
		}},
		{name: "shared DAG", connect: func(root, leaf *{{.NodeType}}) {
			root.Children = {{.NodeListType}}{leaf, leaf}
			root.ByName = {{.NodeMapType}}{"same": leaf}
		}},
		{name: "mixed", connect: func(root, leaf *{{.NodeType}}) {
			root.Children = {{.NodeListType}}{&{{.NodeType}}{
				Label: "middle",
				Choice: {{.NodeChoice}}(&{{.NodeType}}{
					Label: "branch", Branch: &{{.BranchType}}{Nodes: {{.NodeListType}}{leaf}},
				}),
			}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, leaf := &{{.NodeType}}{Label: "root"}, &{{.NodeType}}{Label: "leaf"}
			test.connect(root, leaf)
			if err := {{.NodeValidator}}(root); err != nil {
				t.Fatal(err)
			}
			leaf.Label = ""
			if err := {{.NodeValidator}}(root); err == nil {
				t.Fatal("invalid descendant accepted")
			}
		})
	}
}

func TestFiniteSelfRecursiveCollections(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		value := {{.MapNodeType}}{"child": {{.MapNodeType}}{"leaf": {{.MapNodeType}}{}}}
		if err := {{.MapNodeValidator}}(value); err != nil {
			t.Fatal(err)
		}
		value["child"]["leaf"] = {{.MapNodeType}}{"one": nil, "two": nil, "three": nil}
		if err := {{.MapNodeValidator}}(value); err == nil {
			t.Fatal("invalid nested map accepted")
		}
	})
	t.Run("slice", func(t *testing.T) {
		value := {{.SliceNodeType}}{ {{.SliceNodeType}}{ {{.SliceNodeType}}{} } }
		if err := {{.SliceNodeValidator}}(value); err != nil {
			t.Fatal(err)
		}
		value[0][0] = {{.SliceNodeType}}{nil, nil, nil}
		if err := {{.SliceNodeValidator}}(value); err == nil {
			t.Fatal("invalid nested slice accepted")
		}
	})
}
`
