/*
Package generator contains the code generation algorithms for a service server,
client, and OpenAPI specification.

Server and Client

The code generated for the service server and client includes:

    - A `service` package that contains the declarations for the service
      interfaces and endpoints which wrap the service methods.
    - A `views` package that contains code to render a result type using a view.
    - transport specific packages for each of the transports defined in the
      design.
    - An example implementation of the client, server, and the service.

OpenAPI

The OpenAPI generator generates a OpenAPI v2 specification for the service
REST endpoints. This generator requires the design to define the HTTP transport.
*/
package generator
