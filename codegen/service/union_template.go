// This file gives plugins the same union renderer used by Goa service, view and
// HTTP types. Plugins supply planned declaration names and branch layouts, so
// the generated constructors and JSON mappings follow the authored Goa union.
package service

import "goa.design/goa/v3/internal/uniontemplate"

// UnionTypeSource renders one UnionTypeData as a codegen.SectionTemplate source.
// The data must contain final declarations and branch type references from the
// owning generated package. The containing file supplies bytes, encoding/json,
// fmt and goa.design/goa/v3 imports under their standard package names.
// Generated JSON methods follow the union's tagged, flat or untagged mapping;
// transport-specific strict input checks remain the caller's responsibility.
var UnionTypeSource = uniontemplate.Source
