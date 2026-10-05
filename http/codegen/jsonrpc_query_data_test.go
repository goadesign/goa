// These tests verify that JSON-RPC path and query generation retain the HTTP plan's
// field names and conversion facts while callers receive independent copies of
// mutable types, defaults, examples and map bindings.
package codegen

import (
	"testing"

	"github.com/stretchr/testify/require"
	"goa.design/goa/v3/expr"
)

func TestJSONRPCQueryDataIsIndependent(t *testing.T) {
	binding := "values"
	item := &expr.UserTypeExpr{TypeName: "Key", AttributeExpr: &expr.AttributeExpr{Type: expr.String, Validation: &expr.ValidationExpr{Pattern: "^[a-z]+$"}}}
	source := &PayloadData{Request: &RequestData{QueryParams: []*ParamData{{
		Element: &Element{HTTPName: "key", AttributeData: &AttributeData{
			Name: "credential", VarName: "credential", FieldName: "QueryValue", Type: item, FieldType: item,
			TypeRef: "*string", ValueTypeRef: "string", Pointer: true, FieldPointer: true,
			DefaultValue: []string{"original"}, Example: map[string]string{"key": "original"},
		}},
		MapQueryParams: &binding,
	}}}}
	snapshot := copyJSONRPCPayload(source)
	require.Len(t, snapshot.Request.QueryParams, 1)
	query := snapshot.Request.QueryParams[0]
	require.Equal(t, "key", query.HTTPName)
	require.Equal(t, "QueryValue", query.FieldName)
	require.True(t, query.Pointer)
	require.Equal(t, "*string", query.TypeRef)
	query.HTTPName = "changed"
	query.Type.(*expr.UserTypeExpr).AttributeExpr.Validation.Pattern = "changed"
	query.DefaultValue.([]string)[0] = "changed"
	query.Example.(map[string]string)["key"] = "changed"
	*query.MapQueryParams = "changed"
	fresh := copyJSONRPCPayload(source).Request.QueryParams[0]
	require.Equal(t, "key", fresh.HTTPName)
	require.Equal(t, "^[a-z]+$", fresh.Type.(*expr.UserTypeExpr).AttributeExpr.Validation.Pattern)
	require.Equal(t, []string{"original"}, fresh.DefaultValue)
	require.Equal(t, map[string]string{"key": "original"}, fresh.Example)
	require.Equal(t, "values", *fresh.MapQueryParams)
}

// TestJSONRPCPathDataIsIndependent checks that changing a copied route field
// cannot change the HTTP type and selector used by another generated file.
func TestJSONRPCPathDataIsIndependent(t *testing.T) {
	item := &expr.UserTypeExpr{TypeName: "Organization", AttributeExpr: &expr.AttributeExpr{
		Type: expr.String, Validation: &expr.ValidationExpr{Pattern: "^[a-z]+$"},
	}}
	source := &PayloadData{Request: &RequestData{PathParams: []*ParamData{{
		Element: &Element{HTTPName: "organization_id", AttributeData: &AttributeData{
			Name: "organization_id", VarName: "organizationID", FieldName: "Organization",
			Type: item, FieldType: item, TypeRef: "string", ValueTypeRef: "string",
			Required: true, FieldPointer: true,
		}},
	}}}}
	snapshot := copyJSONRPCPayload(source)
	require.Len(t, snapshot.Request.PathParams, 1)
	route := snapshot.Request.PathParams[0]
	require.Equal(t, "Organization", route.FieldName)
	require.True(t, route.Required)
	require.True(t, route.FieldPointer)
	route.FieldName = "Changed"
	route.Type.(*expr.UserTypeExpr).AttributeExpr.Validation.Pattern = "changed"
	fresh := copyJSONRPCPayload(source).Request.PathParams[0]
	require.Equal(t, "Organization", fresh.FieldName)
	require.Equal(t, "^[a-z]+$", fresh.Type.(*expr.UserTypeExpr).AttributeExpr.Validation.Pattern)
	require.Equal(t, "^[a-z]+$", fresh.FieldType.(*expr.UserTypeExpr).AttributeExpr.Validation.Pattern)
}
