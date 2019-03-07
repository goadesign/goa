/*Package grpc contains code generation logic to produce a server that
serves gRPC requests and a client to encode requests to and decode responses
from a gRPC server. It produces gRPC service definition (.proto files) from the
design, compiles the definition using the protocol buffer compiler (protoc)
using the gRPC in Go plugin, and generates code that hooks up the compiled
protocol buffer types with the goa generated types. It uses "proto3" syntax to
generate gRPC service and protocol buffer message definitions.

In addition to the code generation logic, the grpc package contains:

	* A customizable server and client handler interface to handle unary and streaming RPCs.
	* Encoder and decoder interface to convert a protocol buffer type to goa type and vice versa.
	* Error handlers to encode and decode error responses.
	* Interceptors (a.k.a middlewares) to wrap additional functionality around unary and streaming RPCs.
*/
package grpc
