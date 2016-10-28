# Generators

Generators provide the entry points for code generation. A generator takes an API expression
produced from an API design and generates one or more files. The process is done in two steps: the
`Writers` function exposed by the generator packages takes the API expression and returns a slice of
`FileWriter`. Each `FileWriter` contains the data required to generate a single output file. The
data includes a slice of `Section`. A `Section` contains a template and the data required to render
it.  The goa code generator package iterates through the list and renders all the templates.

This two step process (first call `Writers` then iterate over the writer sections and render the
templates) makes it possible to modify the templates prior to the code generator iterating. This
provides the basic for writing code generators that modify the output of existing ones. An example
would be a middleware generator that modifies the controller generator templates to inject code
prior and/or after the action is run.

```go
// Writers accepts the API expression and returns the file writers used to generate the output.
func Writers(api *design.ApiExpression) []FileWriter, error

// A FileWriter exposes a set of Sections and the relative path to the output file.
type FileWriter interface {
    Sections() []Section
    RelPath string
}

// A Section consists of a template and accompaying render data.
type Section struct {
    // Template used to render section text.
    Template text.Template
    // Data used as input of template.
    Data interface{}
}
```
