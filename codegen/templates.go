package codegen

import (
	"embed"

	"goa.design/goa/v3/codegen/template"
)

// Template constants
const (
	// Header template
	headerT = "header"

	// Transform templates
	transformGoArrayTmplName          = "transform_go_array"
	transformGoMapTmplName            = "transform_go_map"
	transformGoUnionTmplName          = "transform_go_union"
	transformGoUnionToObjectTmplName  = "transform_go_union_to_object"
	transformGoObjectToUnionTmplName  = "transform_go_object_to_union"

	// Validation templates
	validationArrayT     = "validation/array"
	validationMapT       = "validation/map"
	validationUnionT     = "validation/union"
	validationUserT      = "validation/user"
	validationEnumT      = "validation/enum"
	validationPatternT   = "validation/pattern"
	validationFormatT    = "validation/format"
	validationExclMinMaxT = "validation/excl_min_max"
	validationMinMaxT    = "validation/min_max"
	validationLengthT    = "validation/length"
	validationRequiredT  = "validation/required"
)

//go:embed templates/*.go.tpl templates/validation/*.go.tpl
var templateFS embed.FS

// codegenTemplates is the shared template reader for the codegen package (package-private).
var codegenTemplates = &template.TemplateReader{FS: templateFS}