/*
Package dsl implements the goa DSL used to define HTTP APIs.

The HTTP DSL adds a "HTTP" function to the DSL constructs that require HTTP
specific information. These include the API, Service, Method and Error DSLs.

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
            Payload(PayloadType)     // has attributes rq1, rq2, rq3 and rq4
            Result(ResultType)       // has attributes rp1 and rp2
            Error("name", ErrorType) // has attributes er1 and er2

            HTTP(func() {
                GET("/{rq1}")    // rq1 read from path parameter
                Param("rq2")     // rq2 read from query string
                Header("rq3")    // rq3 read from header
                Body(func() {
                    Attribute("rq4") // rq4 read from body field, default
                })
                Response(StatusOK, func() {
                    Header("rp1")    // rp1 written to header
                    Body(func() {
                        Attribute("rp2") // rp2 written to body field, default
                    })
                })
                Response(StatusBadRequest, func() {
                    Header("er1")    // er1 written to header
                    Body(func() {
                        Attribute("er2") // er2 written to body field, default
                    })
                })
            })
        })
    })

By default the payload, result and error type attributes define the request and
response body fields respectively. Any attribute that is not explicitly mapped
is used to define the request or response body. The default response status code
is 200 OK for response types other than Empty and 204 NoContent for the Empty
response type. The default response status code for errors is 400.

The example above can thus be simplified to:

    var _ = Service("name", func() {
        Method("name", func() {
            Payload(PayloadType)     // has attributes rq1, rq2, rq3 and rq4
            Result(ResultType)       // has attributes rp1 and rp2
            Error("name", ErrorType) // has attributes er1 and er2

            HTTP(func() {
                GET("/{rq1}")    // rq1 read from path parameter
                Param("rq2")     // rq2 read from query string
                Header("rq3")    // rq3 read from header
                Response(StatusOK, func() {
                    Header("rp1")    // rp1 written to header
                })
                Response("name", StatusBadRequest, func() {
                    Header("er1")    // er1 written to header
                })
            })
        })
    })

*/
package dsl
