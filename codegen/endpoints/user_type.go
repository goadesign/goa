package genserver

import (
	"github.com/goadesign/goa/codegen"
	"github.com/goadesign/goa/design"
)

type (
	// UserTypesWriter generate code for a goa application user types.
	// User types are data structures defined in the DSL with "Type".
	UserTypesWriter struct {
		*codegen.SourceFile
	}
)

// NewUserTypesWriter returns a contexts code writer.
// User types contain custom data structured defined in the DSL with "Type".
func NewUserTypesWriter(filename string) (*UserTypesWriter, error) {
	file, err := codegen.SourceFileFor(filename)
	if err != nil {
		return nil, err
	}
	return &UserTypesWriter{SourceFile: file}, nil
}

// Execute writes the code for the context types to the writer.
func (w *UserTypesWriter) Execute(t *design.UserTypeExpr) error {
	return w.ExecuteTemplate("types", userTypeT, nil, t)
}

// userTypeT generates the code for a user type.
// template input: UserTypeTemplateData
const userTypeT = `// {{ gotypedesc . false }}{{ $privateTypeName := gotypename . .AllRequired 0 true }}
type {{ $privateTypeName }} {{ gotypedef . 0 true true }}
{{ $assignment := recursiveFinalizer .AttributeDefinition "ut" 1 }}{{ if $assignment }}// Finalize sets the default values for {{$privateTypeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 true }}) Finalize() {
{{ $assignment }}
}{{ end }}
{{ $validation := recursiveValidate .AttributeDefinition false false false "ut" "response" 1 true }}{{ if $validation }}// Validate validates the {{$privateTypeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 true }}) Validate() (err error) {
{{ $validation }}
	return
}{{ end }}
{{ $typeName := gotypename . .AllRequired 0 false }}
// Publicize creates {{ $typeName }} from {{ $privateTypeName }}
func (ut {{ gotyperef . .AllRequired 0 true }}) Publicize() {{ gotyperef . .AllRequired 0 false }} {
	var pub {{ gotypename . .AllRequired 0 false }}
	{{ recursivePublicizer .AttributeDefinition "ut" "pub" 1 }}
	return &pub
}

// {{ gotypedesc . true }}
type {{ $typeName }} {{ gotypedef . 0 true false }}
{{ $validation := recursiveValidate .AttributeDefinition false false false "ut" "response" 1 false }}{{ if $validation }}// Validate validates the {{$typeName}} type instance.
func (ut {{ gotyperef . .AllRequired 0 false }}) Validate() (err error) {
{{ $validation }}
	return
}{{ end }}
`
