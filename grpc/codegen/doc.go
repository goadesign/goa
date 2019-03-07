/*Package codegen contains the code generation logic to generate gRPC service
definitions (.proto files) from the design DSLs and the corresponding server
and client code that wraps the goa-generated endpoints with the protocol buffer
compiler (protoc) generated clients and servers.

The code generator uses "proto3" syntax for generating the proto files.

The code generator compiles the proto files using the protocol buffer compiler
(protoc) with the gRPC in Go plugin. It hooks up the generated protocol buffer
types to the goa generated types as follows:

	* It generates a server that implements the protoc-generated gRPC server interface.
	* It generates a client that invokes the protoc-generated gRPC client.
	* It generates encoders and decoders that transforms the protocol buffer types and gRPC metadata into goa types and vice versa.
	* It generates validations to validate the protocol buffer message types and gRPC metadata fields with the validations set in the design.
*/
package codegen
