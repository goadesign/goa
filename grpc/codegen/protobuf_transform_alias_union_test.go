package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"

	"goa.design/goa/v3/codegen"
	. "goa.design/goa/v3/dsl"
	"goa.design/goa/v3/expr"
)

// DSL defining a union whose members are distinct user types defined in the
// same package and all alias the same underlying type. This mirrors a common
// pattern where member user types share an underlying alias type.
var aliasUnionSamePkgDSL = func() {
	var Alias = Type("Alias", String)

	var SourceMetric = Type("SourceMetric", Alias)
	var SourceOutput = Type("SourceOutput", Alias)
	var SourceSetting = Type("SourceSetting", Alias)
	var SourceControlPoint = Type("SourceControlPoint", Alias)

	Type("SeriesSource", func() {
		Attribute("Source", func() {
			OneOf("Source", func() {
				Attribute("Metric", SourceMetric)
				Attribute("Output", SourceOutput)
				Attribute("Setting", SourceSetting)
				Attribute("ControlPoint", SourceControlPoint)
			})
		})
	})
}

// Ensure Go -> protobuf transform does not require wrapper types when the
// union and its member user types live in the same package and are unique.
func TestProtoBufTransform_AliasUnionSamePkg_NoWrap_GoToProto(t *testing.T) {
	root := codegen.RunDSL(t, aliasUnionSamePkgDSL)

	sd := &ServiceData{Name: "S", Scope: codegen.NewNameScope()}
	svcCtx := serviceTypeContext("proto", sd.Scope)
	pbCtx := protoBufTypeContext("proto", sd.Scope, true)

	// Source: service type; Target: protobuf message for same type name
	source := &expr.AttributeExpr{Type: root.UserType("SeriesSource")}
	target := &expr.AttributeExpr{Type: root.UserType("SeriesSource")}

	// Build protobuf target message
	target = makeProtoBufMessage(expr.DupAtt(target), target.Type.Name(), sd)

	code, _, err := protoBufTransform(source, target, "source", "target", svcCtx, pbCtx, true, true)
	require.NoError(t, err)

	// Generated switch type cases must not use any "Wrap" types.
	require.NotContains(t, code, "Wrap", "Go->Proto union transform should not use wrapper types when not generated")
}
