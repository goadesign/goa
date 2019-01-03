package testdata

const UnaryRPCsProtoCode = `// Service is the ServiceUnaryRPCs service interface.
service ServiceUnaryRPCs {
	// MethodUnaryRPCA implements MethodUnaryRPCA.
	rpc MethodUnaryRPCA (MethodUnaryRPCARequest) returns (MethodUnaryRPCAResponse);
	// MethodUnaryRPCB implements MethodUnaryRPCB.
	rpc MethodUnaryRPCB (MethodUnaryRPCBRequest) returns (MethodUnaryRPCBResponse);
}

message MethodUnaryRPCARequest {
	sint32 int = 1;
	string string_ = 2;
}

message MethodUnaryRPCAResponse {
	repeated bool array_field = 1;
	map<string, double> map_field = 2;
}

message MethodUnaryRPCBRequest {
	uint32 u_int = 1;
	float float32 = 2;
}

message MethodUnaryRPCBResponse {
	repeated bool array_field = 1;
	map<string, double> map_field = 2;
}
`

const UnaryRPCNoPayloadProtoCode = `// Service is the ServiceUnaryRPCNoPayload service interface.
service ServiceUnaryRPCNoPayload {
	// MethodUnaryRPCNoPayload implements MethodUnaryRPCNoPayload.
	rpc MethodUnaryRPCNoPayload (MethodUnaryRPCNoPayloadRequest) returns (MethodUnaryRPCNoPayloadResponse);
}

message MethodUnaryRPCNoPayloadRequest {
}

message MethodUnaryRPCNoPayloadResponse {
	string field = 1;
}
`

const UnaryRPCNoResultProtoCode = `// Service is the ServiceUnaryRPCNoResult service interface.
service ServiceUnaryRPCNoResult {
	// MethodUnaryRPCNoResult implements MethodUnaryRPCNoResult.
	rpc MethodUnaryRPCNoResult (MethodUnaryRPCNoResultRequest) returns (MethodUnaryRPCNoResultResponse);
}

message MethodUnaryRPCNoResultRequest {
	repeated string field = 1;
}

message MethodUnaryRPCNoResultResponse {
}
`

const ServerStreamingRPCProtoCode = `// Service is the ServiceServerStreamingRPC service interface.
service ServiceServerStreamingRPC {
	// MethodServerStreamingRPC implements MethodServerStreamingRPC.
	rpc MethodServerStreamingRPC (MethodServerStreamingRPCRequest) returns (stream MethodServerStreamingRPCResponse);
}

message MethodServerStreamingRPCRequest {
	sint32 field = 1;
}

message MethodServerStreamingRPCResponse {
	string field = 1;
}
`

const ClientStreamingRPCProtoCode = `// Service is the ServiceClientStreamingRPC service interface.
service ServiceClientStreamingRPC {
	// MethodClientStreamingRPC implements MethodClientStreamingRPC.
	rpc MethodClientStreamingRPC (stream MethodClientStreamingRPCStreamingRequest) returns (MethodClientStreamingRPCResponse);
}

message MethodClientStreamingRPCStreamingRequest {
	sint32 field = 1;
}

message MethodClientStreamingRPCResponse {
	string field = 1;
}
`

const BidirectionalStreamingRPCProtoCode = `// Service is the ServiceBidirectionalStreamingRPC service interface.
service ServiceBidirectionalStreamingRPC {
	// MethodBidirectionalStreamingRPC implements MethodBidirectionalStreamingRPC.
	rpc MethodBidirectionalStreamingRPC (stream MethodBidirectionalStreamingRPCStreamingRequest) returns (stream MethodBidirectionalStreamingRPCResponse);
}

message MethodBidirectionalStreamingRPCStreamingRequest {
	sint32 field = 1;
}

message MethodBidirectionalStreamingRPCResponse {
	sint32 a = 1;
	string b = 2;
}
`

const MessageUserTypeWithPrimitivesMessageCode = `
message MethodMessageUserTypeWithPrimitivesRequest {
	bool boolean_field = 1;
	sint32 int_field = 2;
	sint32 int32_field = 3;
	sint64 int64_field = 4;
	uint32 u_int_field = 5;
	uint32 u_int32_field = 6;
	uint64 u_int64_field = 7;
}

message MethodMessageUserTypeWithPrimitivesResponse {
	float float32_field = 1;
	double float64_field = 2;
	string string_field = 3;
	bytes bytes_field = 4;
}
`

const MessageUserTypeWithNestedUserTypesCode = `
message MethodMessageUserTypeWithNestedUserTypesRequest {
	bool boolean_field = 1;
	sint32 int_field = 2;
	UTLevel1 ut_level1 = 3;
}

message UTLevel1 {
	sint32 int32_field = 1;
	sint64 int64_field = 2;
	UTLevel2 ut_level2 = 3;
}

message UTLevel2 {
	sint64 int64_field = 2;
}

message MethodMessageUserTypeWithNestedUserTypesResponse {
	RecursiveT recursive = 1;
}

message RecursiveT {
	RecursiveT recursive = 1;
}
`

const MessageResultTypeCollectionCode = `
message MethodMessageUserTypeWithNestedUserTypesRequest {
}

message RTCollection {
	repeated RT field = 1;
}

message RT {
	sint32 int_field = 1;
}
`

const MessageArrayCode = `
message MethodMessageArrayRequest {
	repeated uint32 array_of_primitives = 1;
	repeated ArrayOfBytes twod_array = 2;
	repeated ArrayOfArrayOfBytes threed_array = 3;
	repeated MapOfStringDouble array_of_maps = 4;
}

message ArrayOfBytes {
	repeated bytes field = 1;
}

message ArrayOfArrayOfBytes {
	repeated ArrayOfBytes field = 1;
}

message MapOfStringDouble {
	map<string, double> field = 1;
}

message MethodMessageArrayResponse {
	repeated UT field = 1;
}

message UT {
	repeated uint32 array_of_primitives = 1;
	repeated ArrayOfBytes twod_array = 2;
	repeated ArrayOfArrayOfBytes threed_array = 3;
	repeated MapOfStringDouble array_of_maps = 4;
}
`

const MessageMapCode = `
message MethodMessageMapRequest {
	map<sint32, UT> field = 1;
}

message UT {
	map<uint32, bool> map_of_primitives = 1;
	map<sint32, ArrayOfUTLevel1> map_of_primitiveut_array = 2;
}

message ArrayOfUTLevel1 {
	repeated UTLevel1 field = 1;
}

message UTLevel1 {
	map<string, MapOfSint32Uint32> map_of_map_of_primitives = 1;
}

message MapOfSint32Uint32 {
	map<sint32, uint32> field = 1;
}

message MethodMessageMapResponse {
	map<uint32, bool> map_of_primitives = 1;
	map<sint32, ArrayOfUTLevel1> map_of_primitiveut_array = 2;
}
`

const MessagePrimitiveCode = `
message MethodMessagePrimitiveRequest {
	uint32 field = 1;
}

message MethodMessagePrimitiveResponse {
	sint32 field = 1;
}
`

const MessageWithMetadataCode = `
message MethodMessageWithMetadataRequest {
	bool boolean_field = 1;
	UTLevel1 ut_level1 = 3;
}

message UTLevel1 {
	sint32 int32_field = 1;
	sint64 int64_field = 2;
}

message MethodMessageWithMetadataResponse {
	UTLevel1 ut_level1 = 3;
}
`

const MessageWithSecurityAttrsCode = `
message MethodMessageWithSecurityRequest {
	string oauth_token = 3;
	bool boolean_field = 1;
}

message MethodMessageWithSecurityResponse {
}
`
