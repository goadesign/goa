# `goa`, the Language

The goa DSL provides a simple way to describe your API design. The DSL consists of global Go
functions that can be nested to build up *definitions*. The top level definition is the API
definition. This definition is what the DSL builds as it executes. There are 3 other top level
definitions: the Resource, MediaType and Type definitions.

Resource definitions describe your API resources. This includes the default media type used to
represent the resource as well as all the actions that can be run on it.

MediaType definitions describe the media types used throughout the API. A media type describes the
body of action responses by specifying their fields in a recursive manner. This description can
also include JSON schema like validation rules that `goa` will use to produce validation code. A
MediaType definition also describes one or more *views* and for each view which fields to render.
Finally a MediaType definition may also define *links* to other resources.

The last top level definition is the Type definition. Type definitions describe data structures in
a similar way that MediaType definitions describe response bodies. In fact MediaType definitions
are a special kind of type definitions. Type definitions can be used to describe the request
payloads and as with MediaType definitions they can include validation rules that `goa` can leverage
to validate the incoming requests.

## API Definition

Typically the first definition to be declared is the API's. This defines the API name, description
and other global properties such as the base path to all the API resource actions. Here is an
example showing all the possible API sub-definitions:
```go
API("API name", func() {
	Title("title")							
	Description("description")
	BasePath("/base/:param")
	BaseParams(func() {
		Param("param")
	})
	ResponseTemplate("static", func() {
		Description("description")
		Status(404)
		MediaType("application/json")
	})
	ResponseTemplate("dynamic", func(arg1, arg2 string) {
		Description(arg1)
		Status(200)
		MediaType(arg2)
	})
	Trait("Authenticated", func() {
		Headers(func() {
			Header("header")
			Required("header")
		})
	})
}
```
#### Title
API title used for documentation
#### Description
API description used for documentation and generated code comments.
#### BasePath
Common request path prefix for all API resource actions
#### BaseParams
Parameters used in `BasePath`, uses the `Attribute` DSL. See [Attribute](#attribute).
#### ResponseTemplate
Response template that action definitions can use to describe their responses. Includes the response
HTTP status, HTTP header specifications and body media type. A response template may accept optional
parameters used to define the response fields. These parameters are set by the action definition
when using the response template to describe a possible response. `goa` comes with a set of
predefined response templates (one per standard HTTP status code). This DSL function makes it
possible to define additional response templates specific to the API.
#### Trait<a name="trait"></a>
Traits that can be leveraged in resource, action and attribute definitions. A trait encapsulates
arbitrary DSL that gets executed wherever the trait is used.

## Resource Definition

There is one resource definition per resource exposed by the API. Each definition describes the
resource attributes via a MediaType definition and the resource actions via Action definitions.
The resource definition can also specify a parent resource, `goa` uses that information coupled with
the BasePath to infer how to build hrefs to the resource collection items. By default `goa` uses
the `show` action if present to compute a resource href (basically concatenating the parent route
with the base path and show route). The resource definition may specify a *canonical action* to
override that default. Here is an example showing all the possible Resource sub-definitions:
```go
Resource("name", func() {
	Description("description")
	DefaultMedia(MediaType)
	BasePath("/path")
	Parent("parent")
	CanonicalActionName("show")
	Trait("Authenticated")
	Action("show", func() {
		Routing(GET("/:id"))
		Response(OK, MediaType.Identifier)
	})
	Response(NotFound)
}
```
#### Description
Resource description used for documentation and generated code comments.
#### DefaultMedia
Resource default media type. This media type is the default representation of the resource in action
responses. Each action may override the default to provide an action specific media type. See
[Media](#media).
#### BasePath
Common request path prefix for all resource actions.
#### Parent
Name of parent resource if any.
#### CanonicalActionName
Name of canonical action if not `show`.
#### Trait
Executes the trait with given name. See [Trait](#trait).
#### Action
Action that can be executed on the resource or resource collection. See [Action](#action).
#### <a name="response"></a>Response
Response defines common responses to all the resource actions. A given action may override the
definition using Response as well. The Response DSL always accepts the name of the response as
first parameter. `goa` defines all the standard HTTP status as global variables so they can be
readily used as response names. Response accepts optional arguments that correspond to the arguments
defined by the corresponding response template (the response template with the same name) if there
is one. A response may also optionally use an anonymous function as last argument to specify the
response status code, media type and headers overriding what the response template specifies:
```go
Response(OK, "vnd.foo.app", func() {
	Headers(func() {
		Header("X-Bar")
	})
	Media(FooMedia) // Overrides "vnd.foo.app"
})
```

## MediaType Definition

A MediaType definition describes the representation of a resource used in a response body. This
includes listing all the *potential* resource attributes that can appear in the body. Views specify
which of the attributes are actually rendered so that the same MediaType definition can represent
multiple rendering of the same resource representation. The MediaType definition attributes can
also define links to other resources. A link is defined using the name of one of the other MediaType
attributes. The attribute type defines the linked resource media type. Links are rendered using the
special "link" view. Media types that are linked to must define that view. Here is an example
showing all the possible MediaType sub-definitions:
```go
MediaType("identifier", func() {
	Description("description")
	Attributes(func() {
		Attribute("name1", Integer, "description1")
		Attribute("name2", String, "description2")
		Attribute("related", RelatedMediaType, "description3")
		Attribute("inline", func() {
			Description("inline data structure")
			Attribute("child1", String, "description")
			Attribute("child2", func() {
				Attribute("grandchild", String, "can nest indefinitely")
			})
			Required("child1")
		})
		Links(func() {
			Link("related")
		})
		Required("name1")
	})
	View("default", func() {
		Attribute("name1")
		Attribute("name2")
		Attribute("links") // Render all the links
	})
	View("extended", func() {
		Attribute("name1")
		Attribute("name2")
		Attribute("related", func() {
			View("extended") // Render "related" using the "extended" view
		})
	})
}
```
#### Description
Media type description used for documentation and generated code comments.
#### Attributes
List of attributes that make up the media type. Each attribute uses the Attribute DSL. In addition
`Links` can be used to define the media type links if any. The `Links` definition also uses the
Attribute DSL where each sub-attribute is declared using `Link`. See [Attribute](#attribute).
#### View
View definition. A view simply lists all the attributes that are rendered. In addition it may
specify the view used to render a specific attribute if the attribute type is a media type and a
different view than the "default" view should be used to render it.

## Type Definition

A type definition describes a data structure attributes. Types can then be used to describe an
action payload (i.e. the shape of incoming request bodies) or any other attribute such as the ones
appearing in other Type or MediaType definitions. Here is an example:
```go
Type("name", func() {
	Description("description")
	Attribute("name", String, "description")
	Required("name")
})
```
#### Description
Type description used for documentation and generated code comments.
#### Attribute
Type attribute. Each attribute uses the Attribute DSL. See [Attribute](#attribute).
#### Required
List of required attribute names. A required attribute must be present in all representations of the
type. This is useful for example to define the required attributes in a request payload.

## Action Definition

Action definitions appear in Resource definitions. They describe actions that can be executed on the
resource collection as a whole (e.g. `list` or `create`) or on elements of the collection (e.g.
`show` or `delete`). An action definition describes the action routes, more than one is possible,
the first route is the default route used to create hrefs to the resource if the action is the
canonical action. The definition also describes the parameters (route parameters and/or query string
parameters) as well as the shape of the payload (request body) and request HTTP headers if any.
Finally the action definition also lists the possible responses together with the HTTP status,
headers and body. Here is an example showing all the possible sub-definitions:
```go
Action("name", func() {
	Description("description")
	Routing(
		PUT("/:id"),
	)
	Params(func() {
		Param("id", Integer)
	})
	Headers(func() {
		Header("Authorization", String)
		Header("X-Account", Integer)
		Required("Authorization", "X-Account")
	})
	Payload(Payload)
	Response(OK)
	Response(NotFound)
})
```
#### Description
Type description used for documentation and generated code comments.
#### Routing
Routing lists the action route. Each route is defined with a function named after the HTTP method.
The route function takes the path as argument. Route paths may use wildcards as described in the
[httprouter](https://godoc.org/github.com/julienschmidt/httprouter) package documentation. These
wildcards define parameters using the `:name` or `*name` syntax where `:name` matches a path
segment and `*name` is a catch-all that matches the path until the end.
#### Params
Params describe the action parameters, either path parameters identified via wildcards or query
string parameters. Each parameter is described via the `Param` function which uses the same DSL
as the Attribute DSL. See [Attribute](#attribute).
#### Headers
Headers list the action request HTTP headers. The DSL syntax is identical to the Attribute DSL. See
[Attribute](#attribute).
#### Payload
Payload describes the HTTP request payload (body data structure). The function accepts either a type
or a DSL that describes the payload members using the Member DSL:
```go
Payload(func() {
	Member("foo", String)
	Member("bar", Integer)
	Required("bar")
})
```
The Member DSL is the same as the Attribute DSL. See [Attribute](#attribute).
#### Response
Each action response is defined using the Response DSL. This DSL is the same as the Response DSL
used to define responses common to all actions at the resource level. See [Response](#response).

## Attribute Definition

Attributes allow describing the data structures that flow through the API. This includes request
and response payloads as well as parameter and header specifications. An attribute definition
is recursive: attributes may include other attributes. At the basic level an attribute has a name,
a type and optionally a default value and validation rules. The type of an attribute can be one of:
* The primitive types `boolean`, `integer`, `number` or `string`.
* A type defined via the `Type` DSL.
* A media type defined via the `MediaType` DSL.
* An object described recursively with child attributes.
* An array defined using the `ArrayOf` DSL keyword.
* An hashmap defined using the `HashOf` DSL keyword.
Attributes can be defined using the `Attribute`, `Param`, `Member` or `Header` DSL keyword depending
on where the definition appears. The syntax for all these definitions is the same. An attribute
definition must specify a name, may specify a type (if it doesn't then the type defaults to String
unless child attributes are defined in which case the type defaults to Object) and may specify a
description. Additionally an attribute definition accepts an optional anonymous function that can
be used to define validations and / or child attributes. All the possible syntax are shows in the
example below:
```go
Attribute("name")

Attribute("name", func() {
	Description("description")
})

Attribute("name", Integer)

Attribute("name", Integer, func() {
	Default(42)
})

Attribute("name", Integer, "description")

Attribute("name", Integer, "description", func() {
	Enum(1, 2)
})
```
Nested attributes:
```go
Attribute("nested", func() {
	Description("description")
	Attribute("child")
	Attribute("child2", func() {
		# ....
	})
})
```
#### Description
Attribute description used for documentation and generated code comments
#### Default
Attribute default value, the type of the argument must match the type of the attribute.
#### Attribute
Child attribute definition. An attribute with children describes a nested data structure which can
be useful to describe request payloads for example.

### Validations

Attribute definitions may include validation rules. These rules are validated against instances of
the corresponding data structure. For example a request payload definition may specify the list of
required child attributes. Validation occurs when the attribute is loaded (for example when `goa`
builds a controller action context from the request state) and when it is dumped (for example when
`goa` renders an HTTP response body).
#### Enum
Enum lists the possible values for the attribute. As with the Attribute Default DSL the type of the
argument given to Enum depends on the type of the attribute.
#### Format
Format only applies to attributes of type String. It specifies the format as defined by the JSON
schema standard that the string must validate. The supported formats are global variables in the
`goa` package starting with the `Format` prefix. At the time of writing these are:
```go
// FormatDateTime defines RFC3339 date time values.
FormatDateTime Format = "date-time"

// FormatEmail defines RFC5322 email addresses.
FormatEmail = "email"

// FormatHostname defines RFC1035 Internet host names.
FormatHostname = "hostname"

// FormatIPv4 defines RFC2373 IPv4 address values.
FormatIPv4 = "ipv4"

// FormatIPv6 defines RFC2373 IPv6 address values.
FormatIPv6 = "ipv6"

// FormatURI defines RFC3986 URI values.
FormatURI = "uri"

// FormatMAC defines IEEE 802 MAC-48, EUI-48 or EUI-64 MAC address values.
FormatMAC = "mac"

// FormatCIDR defines RFC4632 and RFC4291 CIDR notation IP address values.
FormatCIDR = "cidr"

// FormatRegexp Regexp defines regular expression syntax accepted by RE2.
FormatRegexp = "regexp"
```
#### Minimum / Maximum
Minimum and maximum specify a the minimum and maximum value that an Integer or Number attribute may
take.
#### MinLength / MaxLength
MinLength and MaxLength specify the minimum and maximum length that a String or Array attribute may
have.
#### Required
Required lists the required child attributes by name.
