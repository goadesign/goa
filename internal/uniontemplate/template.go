// Package uniontemplate supplies one template for the selected-branch types
// emitted in service, views, HTTP and JSON-RPC packages. Planning provides the
// exact declarations and JSON mapping; rendering chooses no names or types.
package uniontemplate

import _ "embed"

// Source is the template shared by service and transport union declarations.
//
//go:embed union_type.go.tpl
var Source string
