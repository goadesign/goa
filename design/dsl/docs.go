// Package dsl implements the goa design language.
//
// The goa DSL provides a simple way to describe your API design. The DSL consists of global Go
// functions that can be nested to build up *definitions*. The top level definition is the API
// definition. This definition is what the DSL builds as it executes. There are 3 other top level
// definitions: the resource, media type and type definitions all created using the corresponding
// global functions (Resource, MediaType and Type).
//
// Resource definitions describe your API resources. This includes the default media type used to
// represent the resource as well as all the actions that can be run on it.
//
// Media type definitions describe the media types used throughout the API. A media type describes
// the body of HTTP responses by listing their attributes (think object fields) in a recursive
// manner. This description can also include JSON schema-like validation rules that goa uses to
// produce validation code. A Media type definition also describes one or more *views* and for each
// view which fields to render. Finally a media type definition may also define *links* to other
// resources. The media type used to render the link on a resource defines a special "link" view
// used by default by goa to render the "links" child attributes.
//
// The last top level definition is the type definition. Type definitions describe data structures
// in a similar way that media type definitions describe response body attributes. In fact, media
// type definitions are a special kind of type definitions that add views and links. Type
// definitions can be used to describe the request payloads as a whole or any attribute appearing
// anywhere (payloads, media types, headers, params etc.) and as with media type definitions they
// can include validation rules that goa leverages to validate attributes of that type.
package dsl
