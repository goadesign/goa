/*
Package encoding provide goa adapters to many different encoders.

Use the Consumes and Produces DSL in your design to specify the encoding supported by the service.

        var _ = API("MyAPI", func() {
                // ...
                Consumes("application/json", "application/gob")
                Produces("application/json")
                // ...
        })

The built-in encoder and decoder media types are:

	- application/json
	- application/xml
	- application/gob and application/x-gob
	- application/msgpack and application/x-msgpack
	- application/binc and application/x-binc
	- application/cbor and application/x-cbor

External encoders and decoders can also be specified via the DSL:

	Produces("application/json", func() {   // Custom encoder
		Package("github.com/goadesign/goa/encoding/json")
	})

The DSL above causes the generated code to use the JSON encoder implemented in the package at
github.com/goadesign/goa/encoding/json rather than the stdlib JSON encoder. Third party encoders
can easily be used via adapter packages that expose the NewDecoder and NewEcoder methods expected
by the generated code, see the json package as an example.
*/
package encoding
