# Design of Interface Between Generated Code and User Code

## Service

### Decoding / Encoding

The goals are:

* Allow injecting low-level encoder or decoder. Generated code should use
  encoders and decoders that match the design "Produces" and "Consumes"
  expressions.

* Allow serializing errors in arbitrary data structures. Default should use
  goa's Error type.

Non-goals:

* Allow serializing responses with arbitrary data structures. Changing the
  response data structures should be done in design.

Request Decoding:

User code provides a function which given a request returns a decoder. This makes
it possible to use different low-level decodes for different request (i.e. JSON
decoder for JSON payloads, XML decoder for XML payloads etc.).

The user code has no control on how the low-level decoder is used. In particular
it cannot change the payload struct given to the service implementation.

The generated code provides a default implementation for the user provided
function which looks at the request content-type and the decoders loaded as a
result of the design "Consumes" expression to figure out the best match.

Response Encoding:

User code provides a function which given a request and a response returns a low
level encoder. Having the request available makes it possible to look at the
Accept header and implement content type negotiation.

The user code has no control on how the low-level encoder is used. In particular
it cannot change the response struct serialized to the response body.

The generated code provides a default implementation for the user provided
function which looks at the request Accept header and the encoders loaded as a
result of the design "Produces" expression to figure out the best match.
