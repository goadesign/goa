package codegen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// TestAnonymousUserUnionArrayNoWrappersFromProto verifies that protobuf
// oneofs nested in arrays convert directly into value-backed service unions.
func TestAnonymousUserUnionArrayNoWrappersFromProto(t *testing.T) {
	root := codegen.RunDSL(t, func() {
		// Two distinct user types
		Alpha := func() {
			// Force a concrete package location to simulate cross-package members
			Meta("struct:pkg:path", "types")
			Attribute("a", Int)
		}
		Beta := func() {
			Meta("struct:pkg:path", "types")
			Attribute("b", Int)
		}
		Type("Alpha", Alpha)
		Type("Beta", Beta)

		// Event with embedded oneof named "Details" of user types only
		Type("Event", func() {
			Attribute("id", String)
			OneOf("Details", func() {
				Attribute("alpha", "Alpha")
				Attribute("beta", "Beta")
			})
		})

		// Container with array of events
		Type("Container", func() {
			Attribute("events", ArrayOf("Event"))
		})
	})

	sd := &ServiceData{Name: "Svc", Scope: codegen.NewNameScope()}
	svcCtx := serviceTypeContext("proto", sd.Scope)
	pbCtx := protoBufTypeContext("proto", sd.Scope, true)

	// Transform protobuf -> Go for Container
	target := &expr.AttributeExpr{Type: root.UserType("Container")}
	source := makeProtoBufMessage(expr.DupAtt(target), target.Type.Name(), sd)

	code, _, err := protoBufTransform(source, target, "source", "target", pbCtx, svcCtx, false, true)
	require.NoError(t, err)
	out := codegen.FormatTestCode(t, "package foo\nfunc transform(){\n"+code+"}")

	// Ensure no per-branch wrapper casts (e.g., types.DetailsAlpha(...)).
	require.NotContains(t, out, "types.Details")
	require.NotContains(t, out, "Details(")
	require.NotContains(t, out, "detailsValue")
	// Sanity check: union switch on Details field exists.
	require.True(t, strings.Contains(out, "switch") && strings.Contains(out, ".Details"))
}
