package apidsl

import (
	"github.com/kyokomi/goa-v1/design"
	"github.com/kyokomi/goa-v1/dslengine"
)

// Setup API DSL roots.
func init() {
	design.Design = design.NewAPIDefinition()
	design.GeneratedMediaTypes = make(design.MediaTypeRoot)
	design.ProjectedMediaTypes = make(design.MediaTypeRoot)
	dslengine.Register(design.Design)
	dslengine.Register(design.GeneratedMediaTypes)
}
