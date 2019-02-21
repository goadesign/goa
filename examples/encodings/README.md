# goa v2 Encodings Example

This simple example demonstrates HTTP responses with the 'text/html' content type encoding.  It consists of a simple
endpoint that concatenates strings.

## Design

Objects cannot be encoded with 'text/html', so the design illustrates several ways to specify a text response that will
be compatible with the encoding.  For example, a response type String:
```go
Result(String)

HTTP(func() {
    // The payload fields are encoded as path parameters.
    GET("/concatstrings/{a}/{b}")
    Response(StatusOK, func() {
        // Respond with text/html
        ContentType("text/html")
    })
})
```


