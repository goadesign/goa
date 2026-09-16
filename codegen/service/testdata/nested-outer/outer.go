// Package nestedouter supplies an external conversion shape whose child types
// come from two distinct Go packages.
package nestedouter

import (
	unusedalpha "goa.design/goa/v3/codegen/service/testdata/a-nested-alpha"
	nestedalpha "goa.design/goa/v3/codegen/service/testdata/nested-alpha"
	nestedbeta "goa.design/goa/v3/codegen/service/testdata/nested-beta"
)

// Envelope contains mapped same-named child types and one deliberately
// unmapped child whose package name collides with the mapped alpha package.
type Envelope struct {
	Alpha  *nestedalpha.Child
	Beta   *nestedbeta.Child
	Unused *unusedalpha.Child
}
