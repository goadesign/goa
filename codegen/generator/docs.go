/*
Package generator contains the code generation algorithms for a service server,
client and OpenAPI specification.

Server and Client

The code generated for the service server and client includes:

    - A `service' package that contains the declarations for the service
      interfaces.
    - A `endpoint' package that contains the declarations for the endpoints
      which wrap the service methods.
    - transport specific packages for each of the transports defined in the
      design.

OpenAPI

The OpenAPI generator generates a OpenAPI v2 specification for the service
REST endpoints. This generator requires the design to define the HTTP transport.
*/
package generator
