// This file verifies that service generation allocates relocated union symbols
// once across every design root that contributes to a generated Go package.
package generator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	servicecodegen "goa.design/goa/v3/codegen/service"
	"goa.design/goa/v3/dsl"
	"goa.design/goa/v3/eval"
	"goa.design/goa/v3/expr"
)

func TestRelocatedUnionPackageNamesCompile(t *testing.T) {
	t.Cleanup(func() { Generators = generators })
	Generators = func(_ string) ([]Genfunc, error) {
		return []Genfunc{Service, Transport}, nil
	}

	root := func() {
		dsl.API("relocated union package names", func() {})

		firstInput := dsl.Type("FirstInput", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Field(1, "text", dsl.String)
			})
			dsl.Required("Value")
		})
		secondInput := dsl.Type("SecondInput", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.OneOf("Value", func() {
				dsl.Field(1, "number", dsl.Int)
			})
			dsl.Required("Value")
		})
		dsl.Service("First", func() {
			dsl.Method("Read", func() {
				dsl.Payload(firstInput)
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/first")
					dsl.Response(200)
				})
				dsl.GRPC(func() {})
			})
		})
		dsl.Service("Second", func() {
			dsl.Method("Read", func() {
				dsl.Payload(secondInput)
				dsl.Result(dsl.String)
				dsl.HTTP(func() {
					dsl.POST("/second")
					dsl.Response(200)
				})
				dsl.GRPC(func() {})
			})
		})
	}
	codegen.RunDSL(t, root)

	dir := t.TempDir()
	genDir := filepath.Join(dir, codegen.Gendir)
	writeGeneratedModule(t, genDir, "gen")
	_, err := Generate(dir, "gen", false)
	require.NoError(t, err)
	runGeneratedTests(t, genDir)
}

func TestServiceRelocatedUnionNamesSpanDesignRoots(t *testing.T) {
	roots := []eval.Root{
		codegen.RunDSL(t, unusedRelocatedValueRoot()),
		codegen.RunDSL(t, relocatedUnionRoot("ZExistingValue", "FirstService")),
		codegen.RunDSL(t, relocatedUnionRoot("MExistingValue", "SecondService")),
		codegen.RunDSL(t, relocatedUnionRoot("AAddedValue", "ThirdService")),
		codegen.RunDSL(t, relocatedDifferentUnionRoot()),
		codegen.RunDSL(t, relocatedTopLevelValueRoot()),
	}
	files, err := Service("goa.design/goa/example", roots)
	require.NoError(t, err)

	var generated strings.Builder
	for _, file := range files {
		if !strings.Contains(file.Path, filepath.Join("gen", "types")) {
			continue
		}
		for _, section := range file.SectionTemplates {
			require.NoError(t, section.Write(&generated))
		}
	}
	code := generated.String()
	require.Equal(t, 1, strings.Count(code, "type Value struct {"), code)
	require.Equal(t, 1, strings.Count(code, "type Value2 struct {"), code)
	require.Equal(t, 1, strings.Count(code, "type ValueKind string"), code)
	require.Equal(t, 1, strings.Count(code, "type Value2Kind string"), code)
	require.Equal(t, 1, strings.Count(code, "type Value3 struct {"), code)
	require.Contains(t, code, "type Value3 struct {\n\tText string", code)
	require.Equal(t,
		[]string{"Value", "Value", "Value", "Value2"},
		[]string{
			unionFieldType(code, "ZExistingValue"),
			unionFieldType(code, "MExistingValue"),
			unionFieldType(code, "AAddedValue"),
			unionFieldType(code, "DifferentValue"),
		},
	)
}

func TestServiceSelectiveRelocatedUnionOwnerCompiles(t *testing.T) {
	root := codegen.RunDSL(t, selectiveRelocatedUnionRoot())
	services := servicecodegen.NewServicesData(root)
	data := services.Get(root.Services[1].Name)
	servicecodegen.SetUserTypeImports("generated.local/gen", data)
	files := servicecodegen.Files(
		"generated.local/gen",
		root.Services[1],
		services,
		make(map[string][]string),
	)
	addServiceImports(files, data)
	dir := t.TempDir()
	for _, file := range files {
		_, err := file.Render(dir)
		require.NoError(t, err)
	}
	writeGeneratedModule(t, dir, "generated.local")
	runGeneratedTests(t, dir)
}

// unusedRelocatedValueRoot declares a relocated type that no service reaches
// and does not force generation. It must not reserve a generated package name.
func unusedRelocatedValueRoot() func() {
	return func() {
		dsl.Type("Value", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("unused", dsl.String)
		})
	}
}

// selectiveRelocatedUnionRoot declares the same relocated union from two
// services so rendering only the later service must still emit its definition.
func selectiveRelocatedUnionRoot() func() {
	return func() {
		first := relocatedValueType("FirstValue")
		second := relocatedValueType("SecondValue")
		dsl.Service("FirstService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(first)
			})
		})
		dsl.Service("SecondService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(second)
			})
		})
	}
}

// relocatedValueType defines one force-generated owner of the shared Value
// union in the generated types package.
func relocatedValueType(name string) expr.UserType {
	return dsl.Type(name, func() {
		dsl.Meta("struct:pkg:path", "types")
		dsl.Meta("type:generate:force")
		dsl.OneOf("Value", func() {
			dsl.Attribute("Bool", dsl.Boolean)
			dsl.Attribute("Enum", dsl.String)
			dsl.Attribute("Number", dsl.Float64)
		})
	})
}

// relocatedTopLevelValueRoot declares an emitted top-level Value after the
// union definitions, so it receives the next available package-wide name.
func relocatedTopLevelValueRoot() func() {
	return func() {
		value := dsl.Type("Value", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Attribute("text", dsl.String)
			dsl.Required("text")
		})
		dsl.Service("FifthService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(value)
			})
		})
	}
}

// relocatedDifferentUnionRoot defines a union with the same natural name but
// a different branch shape, so it must receive the next package-wide name.
func relocatedDifferentUnionRoot() func() {
	return func() {
		value := dsl.Type("DifferentValue", func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Meta("type:generate:force")
			dsl.OneOf("Value", func() {
				dsl.Attribute("Bool", dsl.Boolean)
				dsl.Attribute("Text", dsl.String)
			})
		})
		dsl.Service("FourthService", func() {
			dsl.Method("Read", func() {
				dsl.Payload(value)
			})
		})
	}
}

// relocatedUnionRoot defines one independently declared union with the same
// name and branch shape as the declarations in the other design roots.
func relocatedUnionRoot(typeName, serviceName string) func() {
	return func() {
		value := dsl.Type(typeName, func() {
			dsl.Meta("struct:pkg:path", "types")
			dsl.Meta("type:generate:force")
			dsl.OneOf("Value", func() {
				dsl.Attribute("Bool", dsl.Boolean)
				dsl.Attribute("Enum", dsl.String)
				dsl.Attribute("Number", dsl.Float64)
			})
		})
		dsl.Service(serviceName, func() {
			dsl.Method("Read", func() {
				dsl.Payload(value)
			})
		})
	}
}

// unionFieldType returns the generated union type referenced by owner.
func unionFieldType(code, owner string) string {
	prefix := "type " + owner + " struct {\n\tValue "
	start := strings.Index(code, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)
	end := strings.IndexByte(code[start:], '\n')
	if end == -1 {
		return ""
	}
	return code[start : start+end]
}
