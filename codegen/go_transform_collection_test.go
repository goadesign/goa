// These tests compile collection conversions with the final import names used
// by both ordinary attribute scopes and retained generated type layouts.
package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/expr"
)

func TestGoTransformCollectionsUseFinalCustomTypeImportAlias(t *testing.T) {
	const (
		owner      = "example.com/transformtest"
		customPath = owner + "/custom/strconv"
	)
	generation := mustTestGeneration(t, owner, nil)
	pkg := mustClaimTestPackage(t, generation, owner)
	require.NoError(t, pkg.DeclareImport(NewImport("strconv", customPath)))
	require.NoError(t, pkg.RequireImport(SimpleImport("strconv")))
	require.NoError(t, generation.Freeze())
	require.Equal(t, "strconv2", pkg.ImportName(customPath))

	token := &expr.AttributeExpr{
		Type: expr.String,
		Meta: expr.MetaExpr{
			"struct:field:type": {"strconv.Token", customPath, "strconv"},
		},
	}
	array := &expr.AttributeExpr{Type: &expr.Array{ElemType: token}}
	shapes := []struct {
		name      string
		attribute *expr.AttributeExpr
		typeName  string
	}{
		{"array", array, "[]strconv2.Token"},
		{"map", &expr.AttributeExpr{Type: &expr.Map{KeyType: token, ElemType: token}}, "map[strconv2.Token]strconv2.Token"},
		{"nested", &expr.AttributeExpr{Type: &expr.Array{ElemType: array}}, "[][]strconv2.Token"},
		{"object", &expr.AttributeExpr{Type: &expr.Array{ElemType: &expr.AttributeExpr{
			Type: &expr.Object{{Name: "token", Attribute: token}},
		}}}, "[]*struct { Token *strconv2.Token }"},
	}

	var source strings.Builder
	source.WriteString("package transformtest\nimport (\n\"strconv\"\nstrconv2 \"" + customPath + "\"\n)\nvar _ = strconv.Itoa\n")
	for _, retained := range []bool{false, true} {
		for _, shape := range shapes {
			context := NewAttributeContext(false, false, true, "", pkg.Scope())
			if retained {
				// The wrapped scope deliberately has no final import binding;
				// the linked type owns the alias in this caller.
				context = NewAttributeContext(false, false, true, "", NewNameScope())
				plan, err := PlanGoType(shape.attribute, GoTypePlanOptions{
					Owner:  owner,
					Policy: context.LayoutPolicy(),
				})
				require.NoError(t, err)
				context, err = context.WithGoTypeLayout(plan.Link(owner, pkg.ImportName))
				require.NoError(t, err)
			}
			linked, err := context.Scope.(GoTypeLayoutResolver).GoTypeLayout(shape.attribute, context.LayoutPolicy())
			require.NoError(t, err)
			require.Equal(t, []GoTypeImport{{Name: "strconv2", Path: customPath}}, linked.Imports())
			leaves := linked.plan.PlansForOccurrence(token)
			require.NotEmpty(t, leaves)
			for _, leaf := range leaves {
				require.Equal(t, "strconv2", linked.Enter(leaf).Package())
			}
			// Reattaching the returned layout exercises the same public path
			// used by a generator that retains a layout for later conversion.
			reused, err := context.WithGoTypeLayout(linked)
			require.NoError(t, err)
			code, helpers, err := GoTransform(
				shape.attribute, shape.attribute, "source", "target",
				context, context, "", true,
			)
			require.NoError(t, err)
			require.Empty(t, helpers)
			require.Contains(t, code, "strconv2.Token")
			require.NotContains(t, code, "strconv.Token")
			reusedCode, reusedHelpers, err := GoTransform(
				shape.attribute, shape.attribute, "source", "target",
				reused, reused, "", true,
			)
			require.NoError(t, err)
			require.Equal(t, code, reusedCode)
			require.Empty(t, reusedHelpers)
			fmt.Fprintf(&source, "func transform%s%t(source %s) %s {\n%s\nreturn target\n}\n",
				shape.name, retained, shape.typeName, shape.typeName, code)
		}
	}

	directory := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(directory, "custom", "strconv"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "go.mod"), []byte("module "+owner+"\n\ngo 1.26.0\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "custom", "strconv", "token.go"), []byte("package strconv\ntype Token string\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(directory, "transform.go"), []byte(source.String()), 0o600))
	command := exec.Command("go", "test", "./...")
	command.Dir = directory
	output, err := command.CombinedOutput()
	require.NoError(t, err, "generated collection conversions did not compile:\n%s", output)
}
