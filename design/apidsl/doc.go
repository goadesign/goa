/*
Package apidsl implements the goa design language.

The goa design language provides a simple way to describe an API design. The language consists of
global Go functions that can be nested to build up *definitions*. The root definition is the API
definition. This definition is what the language builds as it executes. There are 3 other top level
definitions: the resource, media type and type definitions all created using the corresponding
global functions (Resource, MediaType and Type).

Resource definitions describe the API resources. This includes the default media type used to
represent the resource as well as all the actions that can be run on it.

Media type definitions describe the media types used throughout the API. A media type describes
the body of HTTP responses by listing their attributes (think object fields) in a recursive manner.
This description can also include JSON schema-like validation rules that goa uses to produce
validation code. A Media type definition also describes one or more *views* and for each view which
fields to render. Finally a media type definition may also define *links* to other resources. The
media type used to render the link on a resource defines a special "link" view used by default by
goa to render the "links" child attributes.

The last top level definition is the type definition. Type definitions describe data structures
in a similar way that media type definitions describe response body attributes. In fact, media
type definitions are a special kind of type definitions that add views and links. Type definitions
can be used to describe the request payloads as a whole or any attribute appearing anywhere
(payloads, media types, headers, params etc.) and as with media type definitions they can include
validation rules that goa leverages to validate attributes of that type.

Package apidsl also provides a generic DSL engine that other DSLs can plug into. Adding a DSL
implementation consists of registering the root DSL object in the design package Roots variable.
The runner iterates through all root DSL definitions and executes the definition sets they expose.

In general there should be one root definition per DSL (the built-in API DSL uses the APIDefinition
as root definition). The root definition can in turn list sets of definitions where a set defines
a unit of execution and allows to control the ordering of execution. Each definition set consists
of a list of definitions. Definitions must implement the design.Definition interface and may
additionally implement the design.Source and design.Validate interfaces.
*/
package apidsl
