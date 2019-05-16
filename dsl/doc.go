/*
Package dsl implements the Goa DSL.

The Goa DSL consists of Go functions that can be composed to describe a remote
service API. The functions are composed using anonymous function arguments, for
example:

    var Person = Type("Person", func() {
        Attribute("name", String)
    })

The package defines a set of "top level" DSL functions - functions that do not
appear within other functions such as Type above and a number of functions that
are meant to be used within others such as Attribute above.

The comments for each function describe the intent, parameters and usage of the
function. A number of DSL functions leverage variadic arguments to emulate
optional arguments, for example these are all valid use of Attribute:

    Attribute("name", String)
    Attribute("name", String, "The name of the person")
    Attribute("name", String, "The name of the person", func() {
        Meta("struct:field:type", "json.RawMessage")
    })

It is recommended to use "dot" import when importing the DSL package to improve
the readability of designs:

    import . "goa.design/goa/v3/dsl"

Importing the DSL package this way makes it possible to write the designs as
shown in the examples above instead of having to prefix each DSL function call
with "dsl." (note: the authors are aware that using "dot" imports is bad
practice in general when writing standard Go code and Goa in particular makes no
use of them outside of writing DSLs. However they DO make designs much easier to
read and maintain).

The general structure of the DSL is shown below (partial list):

    API                 Service          Type            ResultType
    ├── Title           ├── Description  ├── Extend      ├── TypeName
    ├── Description     ├── Docs         ├── Reference   ├── ContentType
    ├── Version         ├── Security     ├── ConvertTo   ├── Extend
    ├── Docs            ├── Error        ├── CreateFrom  ├── Reference
    ├── License         ├── GRPC         ├── Attribute   ├── ConvertTo
    ├── TermsOfService  ├── HTTP         ├── Field       ├── CreateFrom
    ├── Contact         ├── Method       └── Required    ├── Attributes
    ├── Server          │   ├── Payload                  └── View
    └── HTTP            │   ├── Result
                        │   ├── Error
                        │   ├── GRPC
                        │   └── HTTP
                        └── Files
*/
package dsl
