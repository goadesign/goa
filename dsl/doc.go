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

The GRPC DSL adds a "GRPC" function to the DSL constructs that require gRPC
specific information. These include the API, Service, Method, and Error DSLs.

For example:

    var _ = API("name", func() {
        Description("Optional description")
        // gRPC specific properties
        GRPC(func() {
        })
    })

The GRPC function defines the mapping of the data type attributes used in the
generic DSL to gRPC messages and metadata.

For example:

    var PayloadType = Type("Payload", func() {
        TypeName("Payload")        // mapped to gRPC message with name "Payload"
        Field(1, "rq1", String)    // mapped to field in "Payload" message
        													 // with name "rq1" and tag number 1
        Field(2, "rq2", String)    // mapped to field in "Payload" message
        													 // with name "rq2" and tag number 2
        Attribute("rq3", Int)
        Attribute("rq4", Int)
    })

    var ResultType = ResultType("application/vnd.result", func() {
        TypeName("Result")         // mapped to gRPC message with name "Result"
        Attributes(func() {
           Attribute("rp1", Int)
           Field(1, "rp2", String) // mapped to field in "Result" message
        													 // with name "rp2" and tag number 1
        })
    })

    var _ = Service("name", func() {
        Method("name", func() {
           Payload(PayloadType)
           Result(ResultType)       // has attributes rp1 and rp2
           Error("name")

    			 GRPC(func() {
               Metadata(func() {   // rq3 and rq4 present in gRPC request metadata
                   Attribute("rq3)
                   Attribute("rq4")
               })
               Message(func() {
                   Attribute("rq1") // rq1 and rq2 present in gRPC request message
               })
               Response(CodeOK, func() {
                   Metadata(func() {
                       Attribute("rp1") // rp1 present in gRPC response metadata
                   })
                   Message(func() {
                       Attribute("rp2") // rp2 present in gRPC response message
                   })
               })
               Response("name", CodeInternal) // responds with error message
                                              // defined by error "name"
           })
    	  })
    })

By default the payload and result type attributes define the request and
response message fields respectively with the exception of security attributes
in payload which are mapped to request metadata unless specified explicitly.
The default response status code is CodeOK for success response and
CodeUnknown for error responses. See google.golang.org/grpc/codes package for
more information on the status codes.
*/
package dsl
