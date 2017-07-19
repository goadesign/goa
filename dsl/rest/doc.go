/*
Package rest implements the goa DSL used to define REST APIs.

The REST DSL adds a "HTTP" function to the generic DSL constructs that require
HTTP specific information. These include the API, Service, Method and Error
DSLs.

For example:

    var _ = API("name", func() {
        Description("Optional description")
        // HTTP specific properties
        HTTP(func() {
            // Base path for all the API requests.
            Path("/path")
        })
    })

The HTTP function defines the mapping of the data type attributes used
in the generic DSL to HTTP parameters (for requests), headers and body fields.

For example:

    var _ = Service("name", func() {
        Method("name", func() {
            Request(RequestType)     // has attributes rq1, rq2, rq3 and rq4
            Response(ResponseType)   // has attributes rp1 and rp2
            Error("name", ErrorType) // has attributes er1 and er2

            HTTP(func() {
                GET("/{rq1}")            // rq1 read from path parameter
                Request(func() {
                    Params(func() {
                        Param("rq2")     // rq2 read from query string
                    })
                    Headers(func() {
                        Header("rq3")    // rq3 read from header
                    })
                    Body(func() {
                        Attribute("rq4") // rq4 read from body field
                    })
                })
                Response(func() {
                    Code(StatusOK)
                    Headers(func() {
                        Header("rp1")    // rp1 written to header
                    })
                    Body(func() {
                        Attribute("rp2") // rp2 written to body field
                    })
                })
                Error("name", func() {
                    Code(StatusBadRequest)
                    Headers(func() {
                        Header("er1")    // er1 written to header
                    })
                    Body(func() {
                        Attribute("er2") // er2 written to body field
                    })
                })
            })
        })
    })

By default the top level type attributes define the request and response bodies.
Also the default response status code is 200 OK for response types other than
Empty and 204 NoContent for the Empty response type. So the following:

    var _ = Service("name", func() {
        Method("name", func() {
            Request(RequestType)
            Response(ResponseType)
            HTTP(func() {
                POST("/")
            })
        })
    })

is equivalent to:

    var _ = Service("name", func() {
        Method("name", func() {
            Request(RequestType)   // has attributes rq1 and rq2
            Response(ResponseType) // has attributes rp1 and rp2
            HTTP(func() {
                POST("/")
                Request(func() {
                    Body(func() {
                        Attribute("rq1") // rq1 read from body field
                        Attribute("rq2") // rq2 read from body field
                    })
                })
                Response(func() {
                    Code(StatusOK)
                    Body(func() {
                        Attribute("rp1") // rp1 written to body field
                        Attribute("rp2") // rp2 written to body field
                    })
                })
            })
        })
    })

The error types also describe the corresponding HTTP response body fields by
default. The default HTTP response status code for errors is 400 Bad Request
except for errors whose name matches one of the built-in names: ErrBadRequest,
ErrUnauthorized, ErrForbidden, ErrNotFound, ErrConflict or
ErrInternalServerError in which case the status code is given by the constant
name (400, 401, 403, 404, 409 and 500 respectively).

The HTTP DSL may override the request or response type attributes or even define
new ones. Attributes listed in the Request or Response HTTP DSLs inherit the
properties (type, description, validations etc.) of the attribute with the same
name defined in the top level request or response if any. For example:

    var _ = Service("name", func() {
        Method("name", func() {
            Response(ResponseType) // has attributes rp1 and rp2
            HTTP(func() {
                GET("/")
                Response(func() {
                    Body(func() {
                        Attribute("rp1") // inherits properties from ResponseType
                        Attribute("http_only", String, "HTTP only field")
                    })
                })
            })
        })
    })

*/
package rest
