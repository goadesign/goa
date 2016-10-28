# The Principles Behind the DSL of goa v2

Like in v1 the top level DSL function in v2 is `API`. The `API` DSL lists the
global properties of the API such as its hostname, its version number etc.

```go
var _ = API("cellar", func() {
	Title("The virtual wine cellar")
	Version("1.0")
	Description("An example of an API implemented with goa")
	Contact(func() {
		Name("goa team")
		Email("admin@goa.design")
		URL("http://goa.design")
				})
	License(func() {
		Name("MIT")
	})
	Docs(func() {
		Description("goa guide")
		URL("http://goa.design/getting-started.html")
	})
	Host("cellar.goa.design")
})
```

The `Service` DSL defines a group of endpoints. This maps to a resource in REST
or a `service` declaration in gRPC. A service may define a default type (with
`DefaultType`). The default type lists common attributes that may be reused
throughout the service endpoint request and response types.

```go
// The "account" service exposes the account resource endpoints.
var _ = Service("account", func() {
	DefaultType(Account)
	HTTP(func() {
		BasePath("/accounts")
	})
```

The service endpoints are described using `Endpoint`. This function defines the
endpoint request and response types. It may also list an arbitrary number of
error responses. An error response has a name an optionally a type. If the
`Endpoint` DSL omits the response type then the service default type is used
instead. The built-in type `Empty` denotes an empty response (no response body
in HTTP, Empty message in gRPC).

```go
	Endpoint("update", func() {
		Description("Change account name")
		Request(UpdateAccount)
		Response(Empty)
		Error(ErrNotFound)
		Error(ErrBadRequest, ErrorResponse)
```

The request, response and error types define the request and responses
*independently of the transport*. The `HTTP` function then defines the mapping
between the type attributes and the actual HTTP requests and responses (which
attributes define HTTP headers, which ones define request path elements or
query string values and which ones map to the body). The `HTTP` function also
defines other HTTP specific properties such as the request path, the service
base path, the response HTTP status codes etc.

```go
		HTTP(func() {
			PUT("/{accountID}")
			Body(func() {
				Attribute("name")
				Required("name")
			})
			Response(NoContent)
			Error(ErrNotFound, NotFound)
			Error(ErrBadRequest, BadRequest, ErrorResponse)
		})
```
In the example above the `accountID` request path parameter is defined by the
attribute of same name of the `UpdateAccount` type and so is the body attribute
`name`.

While a service may only define one response type the `HTTP` function may list
multiple responses. Each response defines the HTTP status code, response body
shape if any and may also list HTTP headers. By default the shape of the body of
responses with HTTP status code 200 is described by the endpoint response type.
The `HTTP` function may optionnally use response type attributes to define
response headers. Any attribute of the response type that is not explicitly used
to define a response header defines a field of the response body implcitly. This
alleviates the need to repeat all the response type attributes to define the
body since in most cases only a few would map to headers.

The response body may also be explicitly described using the function `Body`.
This function may be given a list of response type attributes in which case
the body shape is an object with corresponding fields. The `Body` function may
also be given the name of a specific attribute of the response type in which
case the response body shape is dictated by the type of the attribute. This
makes it possible to define response bodies that contain arrays instead of
objects for example:

```go
	Endpoint("index", func() {
		Description("Index all accounts")
		Request(ListAccounts)
		Response(func() {
			Attribute("marker", String, "Pagination marker")
			Attribute("accounts", CollectionOf(Account), "list of accounts")
		})
		HTTP(func() {
			GET("")
			Response(OK, func() {
				Header("marker")
				Body("accounts")
			})
		})
	})
```

The example produces response bodies of the form
`[{"name"="foo"},{"name"="bar"}]` assuming the type `Account` only has a `name`
attribute. The same example as above but with the line defining the response
body (`Body("accounts")`) removed produces response bodies of the form:
`{"accounts":[{"name"="foo"},{"name"="bar"}]` since `accounts` isn't used
for headers or parameters.
